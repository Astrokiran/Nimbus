package models

import (
	"time"

	"gorm.io/gorm"
)

// Guide represents the guide details in the database
type Guide struct {
	GuideID     uint           `gorm:"primaryKey;column:guide_id" json:"guide_id"`
	UserID      uint           `gorm:"index;not null" json:"user_id"` // Foreign key to users.User
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Name        string         `gorm:"type:varchar(255);not null" validate:"required"`
	AreaCode    string         `gorm:"type:varchar(10);not null;uniqueIndex:idx_guide_phone" validate:"required"`
	PhoneNumber string         `gorm:"type:varchar(20);not null;uniqueIndex:idx_guide_phone" validate:"required"`
	Gender      string         `gorm:"type:varchar(50)"`
	Skills      string         `gorm:"type:text"` // Store as comma-separated values or consider a separate skills table for more flexibility
	Languages   string         `gorm:"type:text"` // Store as comma-separated values or consider a separate languages table
	PhotoURL    string         `gorm:"type:varchar(255)"`
	// Additional extendable fields can be added here
	IsActive bool `gorm:"default:true"`

	// Auth related fields
	OtpSecret       string     `json:"-"` // Store encrypted
	OtpGeneratedAt  *time.Time `json:"-"`
	OtpAttemptCount int        `gorm:"default:0" json:"-"`
	IsOtpBlocked    bool       `gorm:"default:false" json:"-"`
	// Other potential fields: Rating, Bio, Experience, etc.
}

// TableName specifies the table name for the Guide model
func (Guide) TableName() string {
	return "guides"
}
