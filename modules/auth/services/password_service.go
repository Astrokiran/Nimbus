package services

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordService handles password hashing and verification.
type PasswordService struct{}

// NewPasswordService creates a new PasswordService.
func NewPasswordService() *PasswordService {
	return &PasswordService{}
}

// HashPassword generates a bcrypt hash for the given password.
func (s *PasswordService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword compares a hashed password with a plain text password.
func (s *PasswordService) VerifyPassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}
