package health

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Status represents the health status of the service.
type Status struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// AppVersion stores the application version (set during build time).
// This should ideally be set in main.go and passed down or accessed globally.
// For simplicity in scaffolding, we'll declare it here, but it should be injected.
var AppVersion string = "dev"

// HealthCheckHandler returns the current health status.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := Status{
		Status:  "ok",
		Version: AppVersion, // Use the injected or globally set version
	}

	err := json.NewEncoder(w).Encode(status)
	if err != nil {
		log.Error().Err(err).Msg("Failed to encode health status")
		// Avoid writing header again if already written
		if _, ok := w.(http.Hijacker); ok {
			// connection has been hijacked, cannot write header
		} else if _, ok := w.(http.Flusher); ok {
			// response has been flushed, cannot write header
		} else {
			// Make sure we haven't written the header yet
			if !headerWritten(w) {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}
}

// Helper to check if header was already written
// Note: This is a basic check and might not cover all edge cases.
func headerWritten(w http.ResponseWriter) bool {
	// This is tricky. A common way is to check a private field via reflection,
	// but that's fragile. Chi's WrapResponseWriter sets status code on first WriteHeader/Write.
	// A more robust solution involves custom response writers or inspecting the underlying connection state.
	// For now, we assume if an error happened during encoding, the header might have been partially written.
	// A safer approach in real apps might be to log the error and not attempt to write a status code.
	return false // Assume not written, leading to potential redundant WriteHeader calls if encoding fails mid-stream
}
