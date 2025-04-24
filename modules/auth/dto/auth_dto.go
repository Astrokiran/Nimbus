package dto

// OtpRequest is the DTO for requesting an OTP.
type OtpRequest struct {
	AreaCode     string `json:"areaCode" validate:"required,numeric"`
	MobileNumber string `json:"mobileNumber" validate:"required,numeric"`
	UserType     string `json:"userType" validate:"required,oneof=CUSTOMER GUIDE"` // Use constants from user model?
}

// OtpVerifyRequest is the DTO for verifying an OTP.
type OtpVerifyRequest struct {
	AreaCode     string `json:"areaCode" validate:"required,numeric"`
	MobileNumber string `json:"mobileNumber" validate:"required,numeric"`
	Otp          string `json:"otp" validate:"required,numeric,len=6"`
	UserType     string `json:"userType" validate:"required,oneof=CUSTOMER GUIDE"`
	DeviceType   string `json:"deviceType" validate:"required,oneof=MOBILE WEB"` // Use constants from auth model?
}

// AdminLoginRequest is the DTO for admin login.
type AdminLoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	DeviceType string `json:"deviceType" validate:"required,oneof=MOBILE WEB"`
}

// RefreshTokenRequest is the DTO for refreshing a token.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// LogoutRequest is the DTO for logout (primarily needs refresh token).
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// TokenResponse is the DTO for returning auth tokens.
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// AccessTokenResponse is the DTO for returning just a new access token (on refresh).
type AccessTokenResponse struct {
	AccessToken string `json:"accessToken"`
}

// AdminUnblockOtpRequest is the DTO for the admin OTP unblock endpoint.
type AdminUnblockOtpRequest struct {
	UserType string `json:"userType" validate:"required,oneof=CUSTOMER GUIDE"`
}
