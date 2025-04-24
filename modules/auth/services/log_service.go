package services

import (
	"nimbus-service/modules/auth/models"

	"gorm.io/gorm"
)

// LogService handles saving authentication-related log entries.
type LogService struct {
	db *gorm.DB
}

// NewLogService creates a new LogService.
func NewLogService(db *gorm.DB) *LogService {
	return &LogService{db: db}
}

// LogActivity creates a record for a login or logout event.
func (s *LogService) LogActivity(userID uint, action models.LoginAction, ipAddress, userAgent string, deviceType models.DeviceType) error {
	activity := models.LoginActivity{
		UserID:     userID,
		Action:     action,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		DeviceType: deviceType,
		// Timestamp is handled by DB default
	}
	return s.db.Create(&activity).Error
}

// LogOtpAttempt creates a record for an OTP verification attempt.
func (s *LogService) LogOtpAttempt(userID, profileID uint, userType, ipAddress, userAgent string, deviceType models.DeviceType, wasValid bool) error {
	attempt := models.OtpAttemptLog{
		UserID:     userID,
		ProfileID:  profileID,
		UserType:   userType,
		WasValid:   wasValid,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		DeviceType: deviceType,
		// Timestamp is handled by DB default
	}
	return s.db.Create(&attempt).Error
}
