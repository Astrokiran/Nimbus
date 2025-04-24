package models

import (
	"time"

	"nimbus-service/modules/users"

	"gorm.io/gorm"
)

// LoginAction represents the type of login activity.
type LoginAction string

const (
	ActionLogin  LoginAction = "LOGIN"
	ActionLogout LoginAction = "LOGOUT"
)

// LoginActivity stores records of user login and logout events.
type LoginActivity struct {
	gorm.Model
	UserID     uint        `gorm:"index;not null"`
	User       users.User  `gorm:"foreignKey:UserID"`
	Action     LoginAction `gorm:"type:varchar(10);not null"`
	Timestamp  time.Time   `gorm:"default:CURRENT_TIMESTAMP;not null"` // Let DB handle default
	IPAddress  string      `gorm:"type:varchar(45)"`
	UserAgent  string      `gorm:"type:varchar(255)"`
	DeviceType DeviceType  `gorm:"type:varchar(10)"` // Reusing DeviceType from refreshtoken.go
}

// TableName specifies the table name for the LoginActivity model.
func (LoginActivity) TableName() string {
	return "auth_login_activities"
}
