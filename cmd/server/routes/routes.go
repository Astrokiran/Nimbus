package routes

import (
	"nimbus-service/internal/config"
	"nimbus-service/internal/middleware"

	"github.com/go-chi/chi/v5"
	validator "github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	// Import necessary module packages and their registration functions/components
	"nimbus-service/modules/health"
	"nimbus-service/modules/users"

	// --- Auth Module Imports ---
	authHandler "nimbus-service/modules/auth/handlers"
	authRoutes "nimbus-service/modules/auth/router"
	authService "nimbus-service/modules/auth/services"

	// --- Customer Module Imports ---
	customerHandler "nimbus-service/modules/customers/handler"
	customerRepo "nimbus-service/modules/customers/repository"
	customerRouter "nimbus-service/modules/customers/router"
	customerService "nimbus-service/modules/customers/service"

	// --- Guide Module Imports ---
	guideHandler "nimbus-service/modules/guides/handler"
	guideRepo "nimbus-service/modules/guides/repository"
	guideRouter "nimbus-service/modules/guides/router"
	guideService "nimbus-service/modules/guides/service"

	// --- Pricing Module Imports ---
	pricingRepo "nimbus-service/modules/pricing/repository"
	pricingService "nimbus-service/modules/pricing/service"
)

// RegisterAllModules initializes dependencies and registers routes for all known modules.
func RegisterAllModules(router *chi.Mux, db *gorm.DB, cfg *config.Config, logger zerolog.Logger) {
	log.Info().Msg("Registering module routes...")

	// --- Initialize Shared Dependencies ---
	validate := validator.New()
	// --- Register Routes (Using a base path if desired) ---

	// --- Initialize Auth Module ---
	pwdSvc := authService.NewPasswordService()
	tokenSvc := authService.NewTokenService(cfg, db)
	otpSvc := authService.NewOtpService(cfg)
	logSvc := authService.NewLogService(db)
	authSvc := authService.NewAuthService(db, cfg, tokenSvc, otpSvc, pwdSvc, logSvc, logger)
	authHdlr := authHandler.NewAuthHandlers(authSvc, validate)
	authMW := middleware.AuthMiddleware(tokenSvc, db)
	log.Debug().Msg("Initialized auth dependencies")

	// --- Initialize Other Modules (Order matters if dependent) ---

	// Pricing
	priceRepo := pricingRepo.NewConsultationPricingRepository(db)
	priceService := pricingService.NewPricingService(priceRepo)
	log.Debug().Msg("Initialized pricing dependencies")

	// Customers
	custRepo := customerRepo.NewCustomerRepository(db)
	custService := customerService.NewCustomerService(custRepo)
	custHandler := customerHandler.NewCustomerHandler(custService)
	log.Debug().Msg("Initialized customers dependencies")

	// Guides
	guideRepo := guideRepo.NewGuideRepository(db)
	guideSvc := guideService.NewGuideService(guideRepo, priceService)
	guideHandler := guideHandler.NewGuideHandler(guideSvc)
	log.Debug().Msg("Initialized guides dependencies")

	// Health (usually at the root)
	health.RegisterRoutes(router)
	log.Debug().Msg("Registered health routes")

	// Users (Placeholder - currently at the root)
	users.RegisterRoutes(router, db, logger)
	log.Debug().Msg("Registered users routes")

	// --- Register API v1 Routes ---
	// Public routes (like login/register) should be registered BEFORE applying auth middleware
	router.Group(func(publicRouter chi.Router) {
		publicRouter.Route("/api/v1/auth", func(authRouter chi.Router) {
			authRoutes.RegisterAuthRoutes(authRouter, authHdlr, authMW) // Pass middleware, let auth module decide internally which routes need it
		})
		log.Debug().Msg("Registered public auth routes under /api/v1/auth")
		// Add other public v1 routes here if needed
	})

	// Protected routes - apply AuthMiddleware to this group
	router.Group(func(protectedRouter chi.Router) {
		protectedRouter.Use(authMW) // Apply AuthMiddleware to all routes in this group
		log.Debug().Msg("Applied AuthMiddleware to protected routes group")

		protectedRouter.Route("/api/v1", func(apiV1 chi.Router) {
			// Customers (Now protected by authMW)
			customerRouter.RegisterCustomerRoutes(apiV1, custHandler)
			log.Debug().Msg("Registered customers routes under /api/v1 (protected)")

			// Guides (Now protected by authMW)
			guideRouter.RegisterGuideRoutes(apiV1, guideHandler)
			log.Debug().Msg("Registered guides routes under /api/v1 (protected)")

			// --- Add future PROTECTED module registrations for v1 here ---
		})
	})

	log.Info().Msg("Finished registering module routes.")
}
