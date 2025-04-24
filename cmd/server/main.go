package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	"nimbus-service/cmd/server/routes" // Import routes package
	"nimbus-service/internal/config"
	"nimbus-service/internal/database"
	appLogger "nimbus-service/internal/logger"         // Alias to avoid conflict
	appMiddleware "nimbus-service/internal/middleware" // Alias to avoid conflict
	"nimbus-service/modules/health"

	// Not needed anymore due to centralized routes registration
	// "nimbus-service/modules/health"
	// "nimbus-service/modules/users"
	// customerHandler "nimbus-service/modules/customers/handler"
	// customerRepo "nimbus-service/modules/customers/repository"
	// customerRouter "nimbus-service/modules/customers/router"
	// customerService "nimbus-service/modules/customers/service"
	"strings" // Ensure strings is imported
)

// Variables to be set by linker flags during build
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func main() {
	// --- Configuration ---
	cfg, err := config.LoadConfig("./configs") // Load from ./configs/config.yaml
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// --- Logger ---
	appLogger.Init(cfg)
	health.AppVersion = Version // Inject version into health module
	log.Info().Msgf("Starting service - Version: %s, Commit: %s, BuildDate: %s", Version, Commit, BuildDate)

	// --- Database ---
	if err := database.Init(cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database connection")
	}

	// --- Router ---
	r := chi.NewRouter()

	// --- Middleware ---
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(appMiddleware.StructuredLogger)          // Use our custom JSON logger
	r.Use(chiMiddleware.Recoverer)                 // Recover from panics
	r.Use(chiMiddleware.Timeout(60 * time.Second)) // Set a reasonable timeout
	// Add CORS, other middleware as needed

	// --- Register Module Routes ---
	// Register all module API routes using the centralized function.
	// It's assumed RegisterAllModules internally handles the /api/v1 prefix now.
	routes.RegisterAllModules(r, database.DB, cfg, appLogger.Logger)

	// --- Setup Swagger/OpenAPI Documentation ---
	// 1. Serve the static swagger.yaml file
	docPath := "./docs"
	specFile := "swagger.yaml"
	filePath := filepath.Join(docPath, specFile)
	r.Get("/swagger/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filePath)
	})

	// Restore original FileServer logic
	swaggerUIPath := "./docs/swagger-ui/"
	fs := http.Dir(swaggerUIPath)
	FileServer(r, "/swagger", fs)

	log.Info().Msg("Registered routes including Swagger UI at /swagger/")

	// --- Server ---
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeoutSeconds) * time.Second,
	}

	// --- Graceful Shutdown ---
	go func() {
		log.Info().Msgf("Server starting on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Wait 10 seconds
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exiting")
}

// FileServer is a helper function to serve static files using Chi router.
// It mimics the behavior of chi.FileServer (which is not directly exported).
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, ":*") {
		panic("FileServer does not permit URL parameters.")
	}

	// Ensure trailing slash for directory listing and index.html
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*" // Match subpaths

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
