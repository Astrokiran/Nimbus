package routes

import (
	"nimbus-service/modules/auth/handlers"

	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterAuthRoutes defines the routes for the authentication module.
// It requires the AuthHandlers struct which contains the handler methods.
func RegisterAuthRoutes(router chi.Router, h *handlers.AuthHandlers, authMW func(http.Handler) http.Handler) {
	// Public auth routes
	authGroup := router.Group(nil)
	{
		authGroup.Post("/otp/request", h.RequestOtp)
		authGroup.Post("/otp/verify", h.VerifyOtp)
		authGroup.Post("/admin/login", h.LoginAdmin)
		authGroup.Post("/token/refresh", h.RefreshToken)
		authGroup.Post("/logout", h.Logout)
	}

	// Admin routes (require authentication and admin privileges)
	adminGroup := router.Group(nil)
	adminGroup.Use(authMW)
	{
		adminGroup.Put("/admin/users/{userId}/otp/unblock", h.AdminUnblockOtpUserHandler)
		// Add other admin-related auth/user management routes here
	}
}
