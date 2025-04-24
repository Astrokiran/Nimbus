package services

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"log"
	"nimbus-service/internal/config"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

var (
	ErrOtpInvalid       = errors.New("invalid OTP code")
	ErrOtpExpired       = errors.New("OTP code has expired")
	ErrSecretGeneration = errors.New("failed to generate OTP secret")
)

// OtpService handles OTP generation, validation, and interaction.
type OtpService struct {
	cfg *config.Config
	// smsClient *SmsClient // Placeholder for actual SMS gateway client
}

// NewOtpService creates a new OtpService.
func NewOtpService(cfg *config.Config /*, smsClient *SmsClient */) *OtpService {
	return &OtpService{
		cfg: cfg,
		// smsClient: smsClient,
	}
}

// GenerateNewSecret creates a new base32 encoded secret key for OTP.
func (s *OtpService) GenerateNewSecret() (string, error) {
	// Generate a random secret
	secret := make([]byte, 20) // 160 bits is recommended
	_, err := rand.Read(secret)
	if err != nil {
		return "", ErrSecretGeneration
	}
	// Encode to base32
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret), nil
}

// generateOtpCode generates a TOTP code based on the secret.
func (s *OtpService) generateOtpCode(secret string) (string, error) {
	// Ensure secret is uppercase as required by some libraries/authenticators
	secret = strings.ToUpper(secret)
	return totp.GenerateCode(secret, time.Now())
}

// VerifyOtpCode validates a provided OTP code against the secret and generation time.
func (s *OtpService) VerifyOtpCode(secret string, providedOtp string, generatedAt time.Time) error {
	// Ensure secret is uppercase
	secret = strings.ToUpper(secret)

	// Basic validation: check if the OTP is within a reasonable time window
	// Example: Allow OTPs generated within the last 5 minutes
	validWindow := 5 * time.Minute
	if time.Since(generatedAt) > validWindow {
		return ErrOtpExpired
	}

	// Validate the code using the library
	valid, err := totp.ValidateCustom(providedOtp, secret, time.Now(), totp.ValidateOpts{
		Period:    30,                // Standard TOTP period
		Skew:      1,                 // Allow 1 time step skew
		Digits:    otp.DigitsSix,     // Use Digits constant from otp package
		Algorithm: otp.AlgorithmSHA1, // Use Algorithm constant from otp package
	})

	if err != nil {
		// Log internal error potentially
		log.Printf("Error during OTP validation: %v\n", err)
		return ErrOtpInvalid // Return a generic error to the user
	}

	if !valid {
		return ErrOtpInvalid
	}

	return nil // OTP is valid
}

// SendOtp sends the OTP code via the configured SMS gateway.
// This is a placeholder and needs integration with a real SMS service.
func (s *OtpService) SendOtp(mobileNumber, otpCode string) error {
	if s.cfg.Auth.Otp.TestModeEnabled {
		// In test mode, just log the OTP instead of sending
		log.Printf("Test Mode: Skipping SMS to %s. OTP: %s\n", mobileNumber, otpCode)
		return nil
	}

	// --- TODO: Integrate with actual SMS Gateway Client ---
	log.Printf("Simulating SMS send to %s: OTP = %s\n", mobileNumber, otpCode)
	// err := s.smsClient.Send(mobileNumber, fmt.Sprintf("Your verification code is: %s", otpCode))
	// if err != nil {
	// 	 log.Printf("Failed to send SMS OTP to %s: %v\n", mobileNumber, err)
	// 	 return err
	// }
	// ----------------------------------------------------

	log.Printf("Successfully sent OTP to %s (simulated)\n", mobileNumber)
	return nil
}
