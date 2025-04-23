package health

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers the health check routes.
func RegisterRoutes(r *chi.Mux) {
	r.Get("/health", HealthCheckHandler)
}
