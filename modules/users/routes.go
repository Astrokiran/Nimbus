package users

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// RegisterRoutes sets up the routes for the users module.
// It requires a database connection and logger.
func RegisterRoutes(r *chi.Mux, db *gorm.DB, log zerolog.Logger) {
	// You would typically create a handler struct that holds dependencies like db and logger
	// For simplicity now, we pass them directly or use package-level handlers (less ideal)

	// Example route group
	r.Route("/users", func(r chi.Router) {
		r.Get("/", GetUsersHandler)    // GET /users
		r.Post("/", CreateUserHandler) // POST /users
		// Add other user routes here (e.g., /users/{id})
	})
}
