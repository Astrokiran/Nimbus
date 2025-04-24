package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"nimbus-service/modules/auth/services"
	"nimbus-service/modules/users"
	"strings"

	"gorm.io/gorm"
)

// Define a custom type for the context key to avoid collisions.
type contextKey string

// UserContextKey is the key used to store the authenticated user in the request context.
const UserContextKey contextKey = "user"

// Helper function to write JSON error responses for middleware
func respondMiddlewareError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	// Simple JSON structure for errors
	jsonStr := `{"error":"` + message + `"}`
	w.Write([]byte(jsonStr))
}

// AuthMiddleware creates a standard net/http middleware for verifying JWT access tokens.
func AuthMiddleware(tokenService *services.TokenService, db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondMiddlewareError(w, http.StatusUnauthorized, "Authorization header required")
				return
			}

			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || strings.ToLower(headerParts[0]) != "bearer" {
				respondMiddlewareError(w, http.StatusUnauthorized, "Authorization header format must be Bearer {token}")
				return
			}

			tokenString := headerParts[1]

			claims, err := tokenService.VerifyAccessToken(tokenString)
			if err != nil {
				errorMsg := "Invalid or expired token"
				if errors.Is(err, services.ErrInvalidToken) {
					// Keep generic message
				} else {
					log.Printf("Unexpected error verifying access token: %v", err)
					errorMsg = "Token verification failed"
				}
				respondMiddlewareError(w, http.StatusUnauthorized, errorMsg)
				return
			}

			var user users.User
			if err := db.Where("id = ? AND is_active = ?", claims.UserID, true).First(&user).Error; err != nil {
				errorMsg := "Authenticated user not found or inactive"
				status := http.StatusUnauthorized
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					log.Printf("Error fetching user %d during auth middleware: %v", claims.UserID, err)
					errorMsg = "Failed to verify user status"
					status = http.StatusInternalServerError // Or keep 401?
				}
				respondMiddlewareError(w, status, errorMsg)
				return
			}

			// Store the user information in the request context.
			ctx := context.WithValue(r.Context(), UserContextKey, &user)

			// Serve the next handler with the updated context.
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves the authenticated user from the standard request context.
// Returns nil if the user is not found in the context.
func GetUserFromContext(ctx context.Context) *users.User {
	userVal := ctx.Value(UserContextKey)
	if userVal == nil {
		return nil
	}

	user, ok := userVal.(*users.User)
	if !ok {
		log.Printf("Unexpected type found in context for UserContextKey: %T", userVal)
		return nil
	}
	return user
}
