package users

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

// Placeholder handler - Replace with actual user logic
func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GET /users called (placeholder)")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Users endpoint (placeholder)"))
}

// Placeholder handler - Replace with actual user logic
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("POST /users called (placeholder)")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User created (placeholder)"))
}
