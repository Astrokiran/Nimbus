package models

import (
	"time"

	"nimbus-service/modules/users"

	"gorm.io/gorm"
)

// OtpAttemptLog represents a single attempt to verify an OTP.
type OtpAttemptLog struct {
	gorm.Model
	UserID     uint       `gorm:"index;not null"` // Link to the central users.User ID
	User       users.User `gorm:"foreignKey:UserID"`
	ProfileID  uint       `gorm:"index;not null"`                     // The ID of the Customer or Guide profile being logged into
	UserType   string     `gorm:"type:varchar(50);index;not null"`    // 'CUSTOMER' or 'GUIDE'
	Timestamp  time.Time  `gorm:"default:CURRENT_TIMESTAMP;not null"` // Let DB handle default
	WasValid   bool       `gorm:"not null"`                           // Did this specific attempt succeed?
	IPAddress  string     `gorm:"type:varchar(45)"`
	UserAgent  string     `gorm:"type:varchar(255)"`
	DeviceType DeviceType `gorm:"type:varchar(10)"` // Reusing DeviceType from refreshtoken.go
}

// TableName specifies the table name for the OtpAttemptLog model.
func (OtpAttemptLog) TableName() string {
	return "auth_otp_attempt_logs"
}
