package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"nimbus-service/internal/middleware"
	"nimbus-service/modules/auth/dto"
	"nimbus-service/modules/auth/services"
	"nimbus-service/modules/users"

	"strconv"
	"strings"

	"github.com/go-chi/chi/v5" // Added for URLParam
	"github.com/go-playground/validator/v10"
	// No more gin-gonic/gin
)

// Helper function to write JSON responses
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

// Helper function to write error JSON responses
func respondError(w http.ResponseWriter, status int, message string, details ...string) {
	errPayload := map[string]interface{}{"error": message}
	if len(details) > 0 && details[0] != "" {
		errPayload["details"] = details[0]
	}
	respondJSON(w, status, errPayload)
}

// AuthHandlers encapsulates the dependencies for auth handlers.
type AuthHandlers struct {
	authService *services.AuthService
	validate    *validator.Validate // Assuming validator is initialized elsewhere
}

// NewAuthHandlers creates a new instance of AuthHandlers.
func NewAuthHandlers(authService *services.AuthService, validate *validator.Validate) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		validate:    validate,
	}
}

// RequestOtp handles the POST /otp/request endpoint.
func (h *AuthHandlers) RequestOtp(w http.ResponseWriter, r *http.Request) {
	var req dto.OtpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	defer r.Body.Close()

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	err := h.authService.RequestOtp(req.AreaCode, req.MobileNumber, req.UserType)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := "Failed to request OTP"
		if errors.Is(err, services.ErrProfileNotFound) {
			statusCode = http.StatusNotFound
			errorMsg = err.Error()
		} else if errors.Is(err, services.ErrOtpBlocked) {
			statusCode = http.StatusForbidden // Or 429 Too Many Requests?
			errorMsg = err.Error()
		} else if errors.Is(err, services.ErrUserInactive) {
			statusCode = http.StatusForbidden
			errorMsg = err.Error()
		} // Add more specific error mappings as needed

		respondError(w, statusCode, errorMsg)
		return
	}

	w.WriteHeader(http.StatusOK) // Or http.StatusAccepted
}

// VerifyOtp handles the POST /otp/verify endpoint.
func (h *AuthHandlers) VerifyOtp(w http.ResponseWriter, r *http.Request) {
	var req dto.OtpVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	defer r.Body.Close()

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Prefer X-Forwarded-For if available, otherwise use RemoteAddr
	ipAddress := r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}
	userAgent := r.UserAgent()

	accessToken, refreshToken, err := h.authService.VerifyOtpAndLogin(req.AreaCode, req.MobileNumber, req.Otp, req.UserType, req.DeviceType, ipAddress, userAgent)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := "OTP verification failed"
		if errors.Is(err, services.ErrOtpInvalid) || errors.Is(err, services.ErrOtpExpired) {
			statusCode = http.StatusUnauthorized
			errorMsg = "Invalid or expired OTP"
		} else if errors.Is(err, services.ErrOtpBlocked) {
			statusCode = http.StatusForbidden
			errorMsg = err.Error()
		} else if errors.Is(err, services.ErrProfileNotFound) {
			statusCode = http.StatusNotFound
			errorMsg = err.Error()
		} else if errors.Is(err, services.ErrUserInactive) {
			statusCode = http.StatusForbidden
			errorMsg = err.Error()
		}

		respondError(w, statusCode, errorMsg)
		return
	}

	respondJSON(w, http.StatusOK, dto.TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken})
}

// LoginAdmin handles the POST /admin/login endpoint.
func (h *AuthHandlers) LoginAdmin(w http.ResponseWriter, r *http.Request) {
	var req dto.AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	defer r.Body.Close()

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	ipAddress := r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}
	userAgent := r.UserAgent()

	accessToken, refreshToken, err := h.authService.LoginAdmin(req.Email, req.Password, req.DeviceType, ipAddress, userAgent)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := "Admin login failed"
		if errors.Is(err, services.ErrInvalidCredentials) {
			statusCode = http.StatusUnauthorized
			errorMsg = err.Error()
		} else if errors.Is(err, services.ErrUserInactive) {
			statusCode = http.StatusForbidden
			errorMsg = err.Error()
		}
		respondError(w, statusCode, errorMsg)
		return
	}

	respondJSON(w, http.StatusOK, dto.TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken})
}

// RefreshToken handles the POST /token/refresh endpoint.
func (h *AuthHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	defer r.Body.Close()

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	newAccessToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		statusCode := http.StatusUnauthorized // Default to unauthorized for refresh failures
		errorMsg := "Failed to refresh token"
		if errors.Is(err, services.ErrInvalidToken) || errors.Is(err, services.ErrTokenRevoked) {
			errorMsg = err.Error()
		} else {
			// Internal server error for unexpected issues
			statusCode = http.StatusInternalServerError
		}
		respondError(w, statusCode, errorMsg)
		return
	}

	respondJSON(w, http.StatusOK, dto.AccessTokenResponse{AccessToken: newAccessToken})
}

// Logout handles the POST /logout endpoint.
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	var req dto.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	defer r.Body.Close()

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// We might need the access token for logging who logged out,
	// but the core action only needs the refresh token.
	// Extracting it from header is an option if needed later.
	// accessToken := r.Header.Get("Authorization") // Need to parse Bearer token

	err := h.authService.Logout(req.RefreshToken)
	if err != nil {
		// Log the error, but usually return success to client anyway
		// as the goal is to invalidate the token best-effort.
		// log.Printf("Error during logout token revocation: %v", err)
		// respondError(w, http.StatusInternalServerError, "Logout failed")
		// return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AdminUnblockOtpUserHandler handles the PUT /admin/users/{userId}/otp/unblock endpoint.
func (h *AuthHandlers) AdminUnblockOtpUserHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Authorization: Ensure caller is an Admin
	// Retrieve user from the standard request context
	callingUser := middleware.GetUserFromContext(r.Context())
	if callingUser == nil || callingUser.UserType != users.AdminUser {
		respondError(w, http.StatusForbidden, "Forbidden: Admin privileges required")
		return
	}

	// 2. Get target UserID from path parameter
	targetUserIDStr := chi.URLParam(r, "userId") // Use chi URLParam
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 32)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	// 3. Bind request body
	var req dto.AdminUnblockOtpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	defer r.Body.Close()

	// 4. Validate request body
	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// 5. Call the service method
	err = h.authService.AdminUnblockOtpUser(uint(targetUserID), req.UserType)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := "Failed to unblock user OTP"
		if errors.Is(err, services.ErrUserNotFound) {
			statusCode = http.StatusNotFound
			errorMsg = err.Error()
		} else if errors.Is(err, services.ErrProfileNotFound) || strings.Contains(err.Error(), "profile not found") {
			statusCode = http.StatusNotFound
			errorMsg = err.Error()
		} else if errors.Is(err, services.ErrUnsupportedUserType) {
			statusCode = http.StatusBadRequest
			errorMsg = err.Error()
		}
		respondError(w, statusCode, errorMsg)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
