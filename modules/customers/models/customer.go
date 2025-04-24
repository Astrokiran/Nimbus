package models

import (
	"time"

	"gorm.io/gorm"
)

// Customer represents the customer details in the database.
type Customer struct {
	// gorm.Model                // Includes ID, CreatedAt, UpdatedAt, DeletedAt
	CustomerID     uint           `gorm:"primaryKey;column:customer_id" json:"customer_id"`
	UserID         uint           `gorm:"index;not null" json:"user_id"` // Foreign key to users.User
	CreatedAt      time.Time      `json:"-"`                             // Exclude GORM timestamps from direct JSON marshal if needed
	UpdatedAt      time.Time      `json:"-"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	AreaCode       string         `gorm:"type:varchar(10);not null;uniqueIndex:idx_customer_phone" validate:"required"`
	MobileNumber   string         `gorm:"type:varchar(20);not null;uniqueIndex:idx_customer_phone" validate:"required"`
	Name           string         `gorm:"type:varchar(255)"`
	Gender         string         `gorm:"type:varchar(50)"` // Consider using an enum or specific type if applicable
	DateOfBirth    *time.Time     `gorm:"type:date"`        // Use pointer for optional field
	TimeOfBirth    *string        `gorm:"type:time"`        // Use pointer for optional field, store as string HH:MM:SS
	PlaceOfBirth   string         `gorm:"type:varchar(255)"`
	CurrentAddress string         `gorm:"type:text"`
	City           string         `gorm:"type:varchar(100)"`
	State          string         `gorm:"type:varchar(100)"`
	Country        string         `gorm:"type:varchar(100)"`
	Pincode        string         `gorm:"type:varchar(20)"`

	// Auth related fields
	OtpSecret       string     `json:"-"` // Store encrypted
	OtpGeneratedAt  *time.Time `json:"-"`
	OtpAttemptCount int        `gorm:"default:0" json:"-"`
	IsOtpBlocked    bool       `gorm:"default:false" json:"-"`
	IsActive        bool       `gorm:"default:true" json:"-"`
}

// TableName specifies the table name for the Customer model.
func (Customer) TableName() string {
	return "customers"
}
