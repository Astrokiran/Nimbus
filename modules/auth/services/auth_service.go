package services

import (
	"errors"
	"fmt"
	customermodel "nimbus-service/modules/customers/models"
	guidemodel "nimbus-service/modules/guides/models"
	"strings"
	"time"

	"nimbus-service/internal/config"
	authmodel "nimbus-service/modules/auth/models"
	usermodel "nimbus-service/modules/users"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

const maxOtpAttempts = 3

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrProfileNotFound     = errors.New("profile not found for the given type and mobile number")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserInactive        = errors.New("user account is inactive")
	ErrOtpBlocked          = errors.New("account blocked due to excessive failed OTP attempts. Please contact support")
	ErrUnsupportedUserType = errors.New("unsupported user type for this operation")
	ErrProfileUpdateFailed = errors.New("failed to update user profile")
	ErrUserCreationFailed  = errors.New("failed to create associated user account")
)

// AuthService orchestrates authentication logic.
type AuthService struct {
	db              *gorm.DB
	cfg             *config.Config
	tokenService    *TokenService
	otpService      *OtpService
	passwordService *PasswordService
	logService      *LogService
	logger          zerolog.Logger
}

// NewAuthService creates a new AuthService.
func NewAuthService(db *gorm.DB, cfg *config.Config, ts *TokenService, os *OtpService, ps *PasswordService, ls *LogService, logger zerolog.Logger) *AuthService {
	return &AuthService{
		db:              db,
		cfg:             cfg,
		tokenService:    ts,
		otpService:      os,
		passwordService: ps,
		logService:      ls,
		logger:          logger.With().Str("service", "AuthService").Logger(),
	}
}

// RequestOtp handles the request for sending an OTP.
// If userType is CUSTOMER and the customer doesn't exist, a new one is created.
// If userType is GUIDE and the guide doesn't exist, an error is returned.
func (s *AuthService) RequestOtp(areaCode, mobileNumber, userTypeString string) (err error) { // Return named error
	var otpSecret string
	var profileID uint
	now := time.Now()
	userType := usermodel.UserType(userTypeString)

	// Use transaction for find/create/update profile logic
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit().Error // Commit transaction if no error occurred within
		}
	}()

	// Find profile and get/generate secret
	switch userType {
	case usermodel.CustomerUser:
		var customer customermodel.Customer
		res := tx.Where("area_code = ? AND mobile_number = ?", areaCode, mobileNumber).First(&customer)

		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			// --- Customer Not Found: Create a new User and Customer profile ---
			s.logger.Info().Msgf("Customer not found for %s-%s, creating new user and profile.", areaCode, mobileNumber)

			// 1. Create the associated User record first with placeholder data
			fullMobile := fmt.Sprintf("%s-%s", areaCode, mobileNumber)
			placeholderEmail := fmt.Sprintf("%s@placeholder.nimbus", fullMobile) // Ensure uniqueness and a placeholder domain
			placeholderName := fmt.Sprintf("User %s", fullMobile)
			// Generate a secure random password hash (or store nil/empty if schema allows & handle login logic)
			// For now, using a placeholder hash - REPLACE with actual secure random hash generation
			placeholderPasswordHash, _ := s.passwordService.HashPassword("defaultSecurePasswordPlaceholder") // TODO: Replace with random secure password

			newUser := usermodel.User{
				Mobile:   fullMobile,               // Set the mandatory mobile number
				Name:     &placeholderName,         // Pass address of string
				Email:    &placeholderEmail,        // Pass address of string
				Password: &placeholderPasswordHash, // Pass address of string
				UserType: usermodel.CustomerUser,
				IsActive: true,
			}
			if err = tx.Create(&newUser).Error; err != nil {
				// Handle potential unique constraint violation on email gracefully
				if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "Duplicate entry") {
					s.logger.Warn().Err(err).Msgf("Potential race condition or duplicate placeholder user for %s. Attempting to find existing.", placeholderEmail)
					// Try to fetch the potentially existing user
					if findErr := tx.Where("email = ?", placeholderEmail).First(&newUser).Error; findErr != nil {
						s.logger.Error().Err(findErr).Msgf("Failed to find existing user %s after unique constraint error", placeholderEmail)
						err = ErrUserCreationFailed // Keep original error type
						return
					}
					// Found user, proceed using this newUser.ID
					s.logger.Info().Msgf("Found existing user %d with placeholder email %s", newUser.ID, placeholderEmail)
				} else {
					s.logger.Error().Err(err).Msgf("Error creating associated user for %s", fullMobile)
					err = ErrUserCreationFailed // Return specific error
					return
				}
			}

			// 2. Now create the Customer profile, linking the UserID
			otpSecret, err = s.otpService.GenerateNewSecret()
			if err != nil {
				return // Return error from GenerateNewSecret
			}
			customer = customermodel.Customer{
				UserID:          newUser.ID, // Link to the newly created or found user
				AreaCode:        areaCode,
				MobileNumber:    mobileNumber,
				IsActive:        true, // Default to active
				OtpSecret:       otpSecret,
				OtpGeneratedAt:  &now,
				OtpAttemptCount: 0,
				IsOtpBlocked:    false,
			}
			if err = tx.Create(&customer).Error; err != nil {
				// If customer creation fails with unique constraint, it might mean user existed but customer didn't (e.g., previous error)
				if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "Duplicate entry") {
					s.logger.Warn().Err(err).Msgf("Customer unique constraint violation for user %d / mobile %s. Assuming customer already exists.", newUser.ID, fullMobile)
					// Attempt to fetch the existing customer profile to proceed
					if findCustErr := tx.Where("user_id = ? AND area_code = ? AND mobile_number = ?", newUser.ID, areaCode, mobileNumber).First(&customer).Error; findCustErr != nil {
						s.logger.Error().Err(findCustErr).Msgf("Failed to find existing customer for user %d after unique constraint error", newUser.ID)
						return // Return the find error
					}
					// Found existing customer, update OTP details on this one
					profileID = customer.CustomerID
					// Need to handle updating the found customer's OTP details like in the 'else' block below
					if !customer.IsActive {
						err = ErrUserInactive
						return
					}
					if customer.IsOtpBlocked {
						err = ErrOtpBlocked
						return
					}
					if customer.OtpSecret == "" {
						customer.OtpSecret, err = s.otpService.GenerateNewSecret()
						if err != nil {
							return
						}
					}
					otpSecret = customer.OtpSecret
					customer.OtpGeneratedAt = &now
					customer.OtpAttemptCount = 0
					if err = tx.Save(&customer).Error; err != nil {
						s.logger.Error().Err(err).Msgf("Error updating existing customer profile %d after constraint recovery", profileID)
						err = ErrProfileUpdateFailed
						return
					}
					// If save successful, continue to OTP generation/send phase
				} else {
					s.logger.Error().Err(err).Msgf("Error creating new customer profile for %s-%s after user creation", fullMobile)
					return // Return other DB creation error
				}
			} else {
				// Customer created successfully
				profileID = customer.CustomerID // Get ID of the newly created customer
				// otpSecret is already set from GenerateNewSecret
			}

		} else if res.Error != nil {
			s.logger.Error().Err(res.Error).Msgf("Error finding customer %s-%s", areaCode, mobileNumber)
			err = res.Error // Return other DB errors
			return
		} else {
			// --- Customer Found: Update existing ---
			if !customer.IsActive {
				err = ErrUserInactive
				return
			}
			if customer.IsOtpBlocked {
				err = ErrOtpBlocked
				return
			}
			profileID = customer.CustomerID
			if customer.OtpSecret == "" {
				customer.OtpSecret, err = s.otpService.GenerateNewSecret()
				if err != nil {
					return
				}
			}
			otpSecret = customer.OtpSecret
			customer.OtpGeneratedAt = &now
			customer.OtpAttemptCount = 0 // Reset attempts on new OTP request
			if err = tx.Save(&customer).Error; err != nil {
				s.logger.Error().Err(err).Msgf("Error updating customer profile %d", profileID)
				err = ErrProfileUpdateFailed
				return
			}
		}

	case usermodel.GuideUser:
		var guide guidemodel.Guide
		res := tx.Where("area_code = ? AND phone_number = ?", areaCode, mobileNumber).First(&guide)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			err = ErrProfileNotFound // Guides are NOT created automatically
			return
		}
		if res.Error != nil {
			s.logger.Error().Err(res.Error).Msgf("Error finding guide %s-%s", areaCode, mobileNumber)
			err = res.Error
			return
		}
		// --- Guide Found: Update existing ---
		if !guide.IsActive {
			err = ErrUserInactive
			return
		}
		if guide.IsOtpBlocked {
			err = ErrOtpBlocked
			return
		}
		profileID = guide.GuideID
		if guide.OtpSecret == "" {
			guide.OtpSecret, err = s.otpService.GenerateNewSecret()
			if err != nil {
				return
			}
		}
		otpSecret = guide.OtpSecret
		guide.OtpGeneratedAt = &now
		guide.OtpAttemptCount = 0 // Reset attempts on new OTP request
		if err = tx.Save(&guide).Error; err != nil {
			s.logger.Error().Err(err).Msgf("Error updating guide profile %d", profileID)
			err = ErrProfileUpdateFailed
			return
		}

	default:
		err = ErrUnsupportedUserType
		return
	}

	// If we reach here without an error, the profile exists (or was created) and is updated.
	// The transaction will be committed by the defer function unless an error occurred below.

	// Generate and Send OTP
	otpCode, err := s.otpService.generateOtpCode(otpSecret)
	if err != nil {
		s.logger.Error().Err(err).Msgf("Error generating OTP code for profile %d (%s)", profileID, userTypeString)
		err = errors.New("failed to generate OTP") // Don't expose internal details
		return
	}

	fullMobile := areaCode + mobileNumber
	if err = s.otpService.SendOtp(fullMobile, otpCode); err != nil {
		s.logger.Error().Err(err).Msgf("Error sending OTP for profile %d (%s) to %s", profileID, userTypeString, fullMobile)
		err = errors.New("failed to send OTP") // Don't expose internal details
		return
	}

	return // Returns nil if commit succeeds, or commit error
}

// VerifyOtpAndLogin validates OTP and logs in a Customer or Guide.
func (s *AuthService) VerifyOtpAndLogin(areaCode, mobileNumber, otp, userTypeString, deviceTypeString, ipAddress, userAgent string) (accessToken string, refreshToken string, err error) {
	var profile interface{}
	var userID uint
	var profileID uint
	var otpSecret string
	var generatedAt *time.Time
	var attemptCount int
	var isBlocked bool
	var isActive bool
	userType := usermodel.UserType(userTypeString)
	deviceType := authmodel.DeviceType(deviceTypeString)

	// --- 1. Find Profile & Check Status ---
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			// Handle panic if necessary, re-panic or log
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit().Error
		}
	}()

	switch userType {
	case usermodel.CustomerUser:
		var customer customermodel.Customer
		if err = tx.Where("area_code = ? AND mobile_number = ?", areaCode, mobileNumber).First(&customer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = ErrProfileNotFound
			}
			return
		}
		profile = &customer
		userID = customer.UserID // Might be 0 if not linked yet
		profileID = customer.CustomerID
		otpSecret = customer.OtpSecret
		generatedAt = customer.OtpGeneratedAt
		attemptCount = customer.OtpAttemptCount
		isBlocked = customer.IsOtpBlocked
		isActive = customer.IsActive
	case usermodel.GuideUser:
		var guide guidemodel.Guide
		if err = tx.Where("area_code = ? AND phone_number = ?", areaCode, mobileNumber).First(&guide).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = ErrProfileNotFound
			}
			return
		}
		profile = &guide
		userID = guide.UserID // Might be 0 if not linked yet
		profileID = guide.GuideID
		otpSecret = guide.OtpSecret
		generatedAt = guide.OtpGeneratedAt
		attemptCount = guide.OtpAttemptCount
		isBlocked = guide.IsOtpBlocked
		isActive = guide.IsActive
	default:
		err = ErrUnsupportedUserType
		return
	}

	if isBlocked {
		err = ErrOtpBlocked
		return
	}
	if !isActive {
		err = ErrUserInactive
		return
	}
	if generatedAt == nil {
		// OTP was never generated or requested
		err = ErrOtpInvalid // Or a more specific error
		return
	}

	// --- 2. Handle Test OTP ---
	var isOtpValid bool
	if s.cfg.Auth.Otp.TestModeEnabled && otp == s.cfg.Auth.Otp.TestValue {
		s.logger.Info().Msgf("Test OTP used for profile %d (%s)", profileID, userType)
		isOtpValid = true
	} else {
		// --- 3. Verify Real OTP ---
		otpErr := s.otpService.VerifyOtpCode(otpSecret, otp, *generatedAt)
		isOtpValid = (otpErr == nil)
		if otpErr != nil && !errors.Is(otpErr, ErrOtpInvalid) && !errors.Is(otpErr, ErrOtpExpired) {
			// Internal error during validation
			s.logger.Error().Err(otpErr).Msgf("Internal error verifying OTP for profile %d (%s)", profileID, userType)
			err = errors.New("OTP verification failed") // Generic error
			return
		}
	}

	// --- 4. Log Attempt ---
	// We need the *final* UserID after potential creation/lookup
	// Defer logging? Or log with potentially zero UserID initially?
	// Let's get/create the user first.

	// --- 5. Find or Create Central User ---
	var user usermodel.User
	if userID != 0 {
		// UserID was already linked in the profile
		if err = tx.First(&user, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.logger.Error().Msgf("Profile %d (%s) has UserID %d, but user record not found.", profileID, userType, userID)
				err = ErrUserNotFound // Indicates data inconsistency
			} else {
				s.logger.Error().Err(err).Msgf("Error fetching user %d", userID)
			}
			return
		}
		// Ensure UserType matches profile type (consistency check)
		if user.UserType != userType {
			s.logger.Warn().Msgf("UserType mismatch for UserID %d. DB: %s, Profile: %s. Updating user record.", userID, user.UserType, userType)
			user.UserType = userType
			if err = tx.Save(&user).Error; err != nil {
				s.logger.Error().Err(err).Msgf("Error updating user type for user %d", userID)
				// Decide if this is fatal, maybe not
			}
		}
	} else {
		// No UserID linked, need to create a new user record
		user = usermodel.User{
			UserType: userType,
			IsActive: true,
			// Name/Email might be empty initially for OTP users
		}
		if err = tx.Create(&user).Error; err != nil {
			s.logger.Error().Err(err).Msgf("Error creating central user for profile %d (%s)", profileID, userType)
			err = ErrUserCreationFailed
			return
		}
		userID = user.ID // Get the newly assigned ID
		// Link it back to the profile
		switch p := profile.(type) {
		case *customermodel.Customer:
			p.UserID = userID
		case *guidemodel.Guide:
			p.UserID = userID
		}
	}

	// Now log the attempt with the correct UserID
	if logErr := s.logService.LogOtpAttempt(userID, profileID, userTypeString, ipAddress, userAgent, deviceType, isOtpValid); logErr != nil {
		// Log the logging error, but don't fail the login for it
		s.logger.Warn().Err(logErr).Msgf("Error logging OTP attempt for UserID %d", userID)
	}

	// --- 6. Handle Validation Result & Update Profile ---
	if !isOtpValid {
		attemptCount++
		if attemptCount >= maxOtpAttempts {
			isBlocked = true
		}
		err = ErrOtpInvalid // Set the final error
	} else {
		attemptCount = 0  // Reset on success
		isBlocked = false // Ensure unblocked on success
	}

	switch p := profile.(type) {
	case *customermodel.Customer:
		p.OtpAttemptCount = attemptCount
		p.IsOtpBlocked = isBlocked
		if errUpdate := tx.Save(p).Error; errUpdate != nil {
			s.logger.Error().Err(errUpdate).Msgf("Error updating customer profile %d after OTP attempt", profileID)
			// If OTP was valid, we might still proceed but log this failure?
			if isOtpValid {
				// Potentially mask this error if login should still succeed?
			} else {
				// If OTP was invalid, this update error is secondary to ErrOtpInvalid
			}
			// For now, let the original error (ErrOtpInvalid) take precedence if OTP failed
		}
	case *guidemodel.Guide:
		p.OtpAttemptCount = attemptCount
		p.IsOtpBlocked = isBlocked
		if errUpdate := tx.Save(p).Error; errUpdate != nil {
			s.logger.Error().Err(errUpdate).Msgf("Error updating guide profile %d after OTP attempt", profileID)
			// Similar error handling logic as customer
		}
	}

	if !isOtpValid {
		if isBlocked {
			err = ErrOtpBlocked // Override ErrOtpInvalid if blocked
		}
		return // Return ErrOtpInvalid or ErrOtpBlocked
	}

	// --- 7. Generate Tokens on Success ---
	accessToken, err = s.tokenService.GenerateAccessToken(userID, userType)
	if err != nil {
		s.logger.Error().Err(err).Msgf("Error generating access token for UserID %d", userID)
		return
	}
	refreshToken, err = s.tokenService.GenerateRefreshToken(userID, deviceType)
	if err != nil {
		s.logger.Error().Err(err).Msgf("Error generating refresh token for UserID %d", userID)
		return
	}

	// --- 8. Log Successful Login ---
	if logErr := s.logService.LogActivity(userID, authmodel.ActionLogin, ipAddress, userAgent, deviceType); logErr != nil {
		s.logger.Warn().Err(logErr).Msgf("Error logging successful login for UserID %d", userID)
		// Don't fail the login for logging error
	}

	// Update LastLogin on User record (async? or in transaction?)
	now := time.Now()
	user.LastLogin = &now
	if errUpdate := tx.Save(&user).Error; errUpdate != nil {
		s.logger.Error().Err(errUpdate).Msgf("Error updating LastLogin for UserID %d", userID)
	}

	// Commit handled by defer
	return
}

// LoginAdmin authenticates an admin user with email and password.
func (s *AuthService) LoginAdmin(email, password, deviceTypeString, ipAddress, userAgent string) (accessToken string, refreshToken string, err error) {
	var user usermodel.User
	deviceType := authmodel.DeviceType(deviceTypeString)

	if err = s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = ErrInvalidCredentials
		} else {
			s.logger.Error().Err(err).Msgf("Database error finding admin user %s", email)
		}
		return
	}

	if user.UserType != usermodel.AdminUser {
		s.logger.Warn().Msgf("Attempt to login non-admin user '%s' via admin endpoint.", email)
		err = ErrInvalidCredentials // Treat as invalid credentials
		return
	}

	if !user.IsActive {
		err = ErrUserInactive
		return
	}

	// Check if password field is set before verifying
	if user.Password == nil {
		s.logger.Warn().Msgf("Admin login attempt for user %s (ID: %d) with nil password field.", email, user.ID)
		err = ErrInvalidCredentials
		return
	}

	if err = s.passwordService.VerifyPassword(*user.Password, password); err != nil { // Dereference pointer
		// Log failed attempt? (Avoid logging password)
		s.logger.Warn().Msgf("Failed admin login attempt for user %s (ID: %d)", email, user.ID)
		err = ErrInvalidCredentials // Generic error for wrong password
		return
	}

	// --- Credentials Valid: Generate Tokens & Log ---
	accessToken, err = s.tokenService.GenerateAccessToken(user.ID, user.UserType)
	if err != nil {
		s.logger.Error().Err(err).Msgf("Error generating access token for Admin UserID %d", user.ID)
		return
	}
	refreshToken, err = s.tokenService.GenerateRefreshToken(user.ID, deviceType)
	if err != nil {
		s.logger.Error().Err(err).Msgf("Error generating refresh token for Admin UserID %d", user.ID)
		return
	}

	if logErr := s.logService.LogActivity(user.ID, authmodel.ActionLogin, ipAddress, userAgent, deviceType); logErr != nil {
		s.logger.Warn().Err(logErr).Msgf("Error logging successful admin login for UserID %d", user.ID)
	}

	// Update LastLogin
	now := time.Now()
	user.LastLogin = &now
	if errUpdate := s.db.Save(&user).Error; errUpdate != nil {
		s.logger.Error().Err(errUpdate).Msgf("Error updating LastLogin for Admin UserID %d", user.ID)
	}

	return
}

// RefreshToken validates a refresh token and issues a new access token.
func (s *AuthService) RefreshToken(refreshTokenString string) (accessToken string, err error) {
	user, err := s.tokenService.VerifyAndUseRefreshToken(refreshTokenString)
	if err != nil {
		// Map token service errors if needed, e.g., ErrInvalidToken, ErrTokenRevoked
		return "", err
	}

	// Generate a new access token
	accessToken, err = s.tokenService.GenerateAccessToken(user.ID, user.UserType)
	if err != nil {
		s.logger.Error().Err(err).Msgf("Error generating new access token during refresh for UserID %d", user.ID)
		return "", err
	}

	return accessToken, nil
}

// Logout revokes the provided refresh token.
func (s *AuthService) Logout(refreshTokenString string /*, optionalAccessToken string */) error {
	// Optional: Verify access token first to get UserID for logging?
	// var userID uint
	// if optionalAccessToken != "" {
	// 	 claims, err := s.tokenService.VerifyAccessToken(optionalAccessToken)
	// 	 if err == nil { userID = claims.UserID }
	// }

	err := s.tokenService.RevokeRefreshToken(refreshTokenString)
	if err != nil {
		s.logger.Error().Err(err).Msgf("Error revoking refresh token")
		// Decide if this should be exposed to the user or just logged
		return err // Or return nil if client logout should succeed anyway?
	}

	// Optional: Log logout activity
	// Need UserID. If not obtained from access token, maybe store UserID with refresh token hash?
	// Or maybe the client provides UserID during logout? Less secure.
	// For now, skipping detailed logout logging in this basic implementation.
	// if userID != 0 {
	// 	 s.logService.LogActivity(userID, authmodel.ActionLogout, "", "", "") // Use ActionLogout
	// }

	return nil
}

// AdminUnblockOtpUser allows an admin to unblock OTP for a specific user.
func (s *AuthService) AdminUnblockOtpUser(targetUserID uint, userTypeString string) error {
	targetUserType := usermodel.UserType(userTypeString)

	tx := s.db.Begin()
	var err error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit().Error
		}
	}()

	// Find the target user first to ensure they exist
	var targetUser usermodel.User
	if err = tx.Where("id = ?", targetUserID).First(&targetUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound // Target user doesn't exist
		}
		s.logger.Error().Err(err).Msgf("DB error finding target user %d for OTP unblock", targetUserID)
		return err // Internal DB error
	}

	// Find and update the specific profile
	switch targetUserType {
	case usermodel.CustomerUser:
		var customer customermodel.Customer
		// Find customer profile linked to the target UserID
		if err = tx.Where("user_id = ?", targetUserID).First(&customer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("customer profile not found for user ID %d", targetUserID)
			}
			return err
		}
		customer.IsOtpBlocked = false
		customer.OtpAttemptCount = 0
		if err = tx.Save(&customer).Error; err != nil {
			s.logger.Error().Err(err).Msgf("Error unblocking customer profile %d (UserID: %d)", customer.CustomerID, targetUserID)
			return ErrProfileUpdateFailed
		}

	case usermodel.GuideUser:
		var guide guidemodel.Guide
		// Find guide profile linked to the target UserID
		if err = tx.Where("user_id = ?", targetUserID).First(&guide).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("guide profile not found for user ID %d", targetUserID)
			}
			return err
		}
		guide.IsOtpBlocked = false
		guide.OtpAttemptCount = 0
		if err = tx.Save(&guide).Error; err != nil {
			s.logger.Error().Err(err).Msgf("Error unblocking guide profile %d (UserID: %d)", guide.GuideID, targetUserID)
			return ErrProfileUpdateFailed
		}

	default:
		return ErrUnsupportedUserType
	}

	s.logger.Info().Msgf("Admin successfully unblocked OTP for UserID %d (%s)", targetUserID, targetUserType)
	return nil // Commit handled by defer
}
