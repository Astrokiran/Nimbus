package models

import (
	"time"

	"nimbus-service/modules/users"

	"gorm.io/gorm"
)

// DeviceType represents the type of client device.
type DeviceType string

const (
	MobileDevice DeviceType = "MOBILE"
	WebDevice    DeviceType = "WEB"
)

// RefreshToken stores information about issued refresh tokens.
type RefreshToken struct {
	gorm.Model
	UserID     uint       `gorm:"index;not null"`    // Foreign key to users.User.ID
	User       users.User `gorm:"foreignKey:UserID"` // Belongs To relationship
	TokenHash  string     `gorm:"index;not null"`
	DeviceType DeviceType `gorm:"type:varchar(10);not null"`
	ExpiresAt  time.Time  `gorm:"not null"`
	Revoked    bool       `gorm:"default:false"`
}

// TableName specifies the table name for the RefreshToken model.
func (RefreshToken) TableName() string {
	return "auth_refresh_tokens"
}
