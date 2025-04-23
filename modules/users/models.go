package users

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system.
// Add GORM tags and more fields as needed.
type User struct {
	gorm.Model        // Includes fields like ID, CreatedAt, UpdatedAt, DeletedAt
	Name       string `gorm:"not null"`
	Email      string `gorm:"uniqueIndex;not null"`
	Password   string `gorm:"not null"` // Store hashed passwords only!
	LastLogin  *time.Time
}

// Add other user-related models here if necessary.
