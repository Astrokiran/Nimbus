package users

import (
	"time"

	"gorm.io/gorm"
)

// UserType represents the type of user in the system.
type UserType string

const (
	AdminUser    UserType = "ADMIN"
	CustomerUser UserType = "CUSTOMER"
	GuideUser    UserType = "GUIDE"
)

// User represents a user in the system.
// Add GORM tags and more fields as needed.
type User struct {
	gorm.Model          // Includes fields like ID, CreatedAt, UpdatedAt, DeletedAt
	Mobile     string   `gorm:"uniqueIndex;not null"` // Added unique mobile number
	Name       *string  // Made optional
	Email      *string  `gorm:"uniqueIndex"` // Made optional, but still unique if provided
	Password   *string  // Made optional (Store hashed passwords only!)
	UserType   UserType `gorm:"type:varchar(50);not null;default:'ADMIN'"`
	IsActive   bool     `gorm:"default:true"`
	LastLogin  *time.Time
}

// Add other user-related models here if necessary.
