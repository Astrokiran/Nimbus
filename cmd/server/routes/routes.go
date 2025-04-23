package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	// Import necessary module packages and their registration functions/components
	"nimbus-service/modules/health"
	"nimbus-service/modules/users"

	customerHandler "nimbus-service/modules/customers/handler"
	customerRepo "nimbus-service/modules/customers/repository"
	customerRouter "nimbus-service/modules/customers/router"
	customerService "nimbus-service/modules/customers/service"

	// Import guide module packages
	guideHandler "nimbus-service/modules/guides/handler"
	guideRepo "nimbus-service/modules/guides/repository"
	guideRouter "nimbus-service/modules/guides/router"
	guideService "nimbus-service/modules/guides/service"

	// Import pricing module packages
	pricingRepo "nimbus-service/modules/pricing/repository"
	pricingService "nimbus-service/modules/pricing/service"
)

// RegisterAllModules initializes dependencies and registers routes for all known modules.
func RegisterAllModules(router *chi.Mux, db *gorm.DB, logger zerolog.Logger) {
	log.Info().Msg("Registering module routes...")

	// --- Health Module ---
	health.RegisterRoutes(router)
	log.Debug().Msg("Registered health routes") // Use Debug for more granular logging

	// --- Users Module ---
	// Assuming users module follows the pattern of needing db and logger injected at registration
	users.RegisterRoutes(router, db, logger)
	log.Debug().Msg("Registered users routes")

	// --- Customers Module ---
	// Initialize Customer Module Dependencies here
	custRepo := customerRepo.NewCustomerRepository(db)
	custService := customerService.NewCustomerService(custRepo)                  // Add validator if needed
	custHandler := customerHandler.NewCustomerHandler(custService /*, logger */) // Pass logger if needed by handler
	// Register Customer Routes
	customerRouter.RegisterCustomerRoutes(router, custHandler)
	log.Debug().Msg("Registered customers routes")

	// --- Pricing Module Dependencies (needed by other modules) ---
	// Initialize Pricing Module Dependencies first as other modules might depend on it
	priceRepo := pricingRepo.NewConsultationPricingRepository(db)
	priceService := pricingService.NewPricingService(priceRepo)
	log.Debug().Msg("Initialized pricing dependencies") // Log initialization

	// --- Guides Module ---
	// Initialize Guide Module Dependencies
	guideRepo := guideRepo.NewGuideRepository(db)
	// Inject the pricing service into the guide service
	guideSvc := guideService.NewGuideService(guideRepo, priceService)
	guideHandler := guideHandler.NewGuideHandler(guideSvc)
	// Register Guide Routes
	guideRouter.RegisterGuideRoutes(router, guideHandler)
	log.Debug().Msg("Registered guides routes")

	// --- Add future modules here ---
	// e.g., someOtherModule.RegisterRoutes(router, db, logger)
	// log.Debug().Msg("Registered someOtherModule routes")

	log.Info().Msg("Finished registering module routes.")
}
