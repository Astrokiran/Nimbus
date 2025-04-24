package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"nimbus-service/internal/config"
	"nimbus-service/modules/auth/models"
	"nimbus-service/modules/users"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrTokenRevoked = errors.New("token has been revoked")
)

// Claims represents the data stored in the JWT access token.
type Claims struct {
	UserID   uint           `json:"user_id"`
	UserType users.UserType `json:"user_type"`
	jwt.RegisteredClaims
}

// TokenService handles JWT access and refresh token generation and validation.
type TokenService struct {
	cfg *config.Config
	db  *gorm.DB
}

// NewTokenService creates a new TokenService.
func NewTokenService(cfg *config.Config, db *gorm.DB) *TokenService {
	return &TokenService{cfg: cfg, db: db}
}

// GenerateAccessToken creates a new JWT access token.
func (s *TokenService) GenerateAccessToken(userID uint, userType users.UserType) (string, error) {
	expirationTime := time.Now().Add(time.Duration(s.cfg.Auth.Jwt.AccessExpiryMinutes) * time.Minute)
	claims := &Claims{
		UserID:   userID,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "nimbus-service", // Optional: identify the issuer
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.Auth.Jwt.Secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// VerifyAccessToken validates an access token string.
func (s *TokenService) VerifyAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.Auth.Jwt.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrInvalidToken
		}
		return nil, err // Other parsing errors
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GenerateRefreshToken creates, stores, and returns a new refresh token.
func (s *TokenService) GenerateRefreshToken(userID uint, deviceType models.DeviceType) (string, error) {
	// Generate a random token string
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	plainToken := hex.EncodeToString(b)

	// Hash the token for storage
	hasher := sha256.New()
	hasher.Write([]byte(plainToken))
	tokenHash := hex.EncodeToString(hasher.Sum(nil))

	expiresAt := time.Now().AddDate(0, 0, s.cfg.Auth.Jwt.RefreshExpiryDays)

	refreshToken := models.RefreshToken{
		UserID:     userID,
		TokenHash:  tokenHash,
		DeviceType: deviceType,
		ExpiresAt:  expiresAt,
		Revoked:    false,
	}

	if err := s.db.Create(&refreshToken).Error; err != nil {
		return "", err
	}

	// Return the plain (non-hashed) token to the client
	return plainToken, nil
}

// hashToken hashes a plain token string using SHA256.
func hashToken(plainToken string) string {
	hasher := sha256.New()
	hasher.Write([]byte(plainToken))
	return hex.EncodeToString(hasher.Sum(nil))
}

// VerifyAndUseRefreshToken validates a refresh token and returns the associated user.
// TODO: Implement refresh token rotation for enhanced security.
func (s *TokenService) VerifyAndUseRefreshToken(tokenString string) (*users.User, error) {
	tokenHash := hashToken(tokenString)
	var refreshToken models.RefreshToken

	// Find the token by its hash
	result := s.db.Preload("User").Where("token_hash = ?", tokenHash).First(&refreshToken)
	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err // Other DB error
	}

	// Check if revoked
	if refreshToken.Revoked {
		return nil, ErrTokenRevoked
	}

	// Check expiry
	if time.Now().After(refreshToken.ExpiresAt) {
		// Optionally revoke expired tokens here or via a background job
		s.db.Model(&refreshToken).Update("revoked", true)
		return nil, ErrInvalidToken
	}

	// Basic validation passed. User is preloaded.
	if refreshToken.User.ID == 0 { // Check if User struct was loaded correctly
		return nil, errors.New("failed to load user associated with refresh token")
	}

	// --- Optional: Implement Token Rotation ---
	// 1. Revoke the current token:
	//    s.db.Model(&refreshToken).Update("revoked", true)
	// 2. Issue a new refresh token for the same user/device:
	//    _, err := s.GenerateRefreshToken(refreshToken.UserID, refreshToken.DeviceType)
	//    if err != nil { /* handle error */ }
	// 3. The client would need to receive and store this new refresh token.
	// --- End Token Rotation ---

	return &refreshToken.User, nil
}

// RevokeRefreshToken marks a refresh token as revoked in the database.
func (s *TokenService) RevokeRefreshToken(tokenString string) error {
	tokenHash := hashToken(tokenString)
	result := s.db.Model(&models.RefreshToken{}).Where("token_hash = ?", tokenHash).Update("revoked", true)

	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		// Token didn't exist, but logout shouldn't necessarily fail.
		// Log this potentially? For now, treat as success.
		return nil
	}
	return nil
}
