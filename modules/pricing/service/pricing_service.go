package service

import (
	"context"
	"errors"
	"fmt"

	"nimbus-service/modules/pricing/models" // Adjust import path if necessary
	"nimbus-service/modules/pricing/repository"

	"gorm.io/gorm"
)

// PricingService defines the interface for pricing business logic.
// This service acts as the primary entry point for other modules interacting with pricing.
type PricingService interface {
	// GetConsultationPricing retrieves the base pricing for a given entity.
	GetConsultationPricing(ctx context.Context, entityType string, entityID uint) (*models.ConsultationPricing, error)

	// GetPricingByEntityIDs retrieves pricing for multiple entities and returns them as a map keyed by EntityID.
	GetPricingByEntityIDs(ctx context.Context, entityType string, entityIDs []uint) (map[uint]*models.ConsultationPricing, error)

	// SetConsultationPricing creates or updates the pricing information for a given entity.
	SetConsultationPricing(ctx context.Context, pricing *models.ConsultationPricing) error

	// --- Future methods ---
	// CalculateEffectivePricing(ctx context.Context, entityType string, entityID uint, userID uint) (*EffectivePricing, error) // Example incorporating offers
}

// Define custom errors for the pricing service
var (
	ErrPricingNotFound     = errors.New("pricing information not found for the specified entity")
	ErrInvalidPricingInput = errors.New("invalid input for pricing information")
)

// pricingServiceImpl implements the PricingService interface.
type pricingServiceImpl struct {
	repo repository.ConsultationPricingRepository
}

// NewPricingService creates a new instance of the pricing service.
func NewPricingService(repo repository.ConsultationPricingRepository) PricingService {
	return &pricingServiceImpl{
		repo: repo,
	}
}

// GetConsultationPricing retrieves the base pricing for a given entity.
func (s *pricingServiceImpl) GetConsultationPricing(ctx context.Context, entityType string, entityID uint) (*models.ConsultationPricing, error) {
	if entityType == "" || entityID == 0 {
		return nil, ErrInvalidPricingInput
	}

	pricing, err := s.repo.GetByEntity(ctx, entityType, entityID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return a service-specific error, or potentially default pricing if desired.
			// For now, return a clear "not found" error.
			// Alternatively, return a default pricing struct: return &models.ConsultationPricing{EntityType: entityType, EntityID: entityID}, nil
			return nil, ErrPricingNotFound
		}
		// Log unexpected errors?
		return nil, fmt.Errorf("failed to get pricing from repository: %w", err)
	}
	return pricing, nil
}

// GetPricingByEntityIDs retrieves pricing for multiple entities and returns them as a map keyed by EntityID.
func (s *pricingServiceImpl) GetPricingByEntityIDs(ctx context.Context, entityType string, entityIDs []uint) (map[uint]*models.ConsultationPricing, error) {
	pricingMap := make(map[uint]*models.ConsultationPricing)
	if len(entityIDs) == 0 {
		return pricingMap, nil
	}

	// Fetch the list from the repository
	pricingList, err := s.repo.GetPricingByEntityIDs(ctx, entityType, entityIDs)
	if err != nil {
		// Log unexpected errors?
		return nil, fmt.Errorf("failed to get pricing list from repository: %w", err)
	}

	// Convert the list to a map for easy lookup
	for _, pricing := range pricingList {
		if pricing != nil { // Add a nil check just in case
			pricingMap[pricing.EntityID] = pricing
		}
	}

	return pricingMap, nil
}

// SetConsultationPricing creates or updates the pricing information for a given entity.
func (s *pricingServiceImpl) SetConsultationPricing(ctx context.Context, pricing *models.ConsultationPricing) error {
	// Basic validation
	if pricing == nil || pricing.EntityType == "" || pricing.EntityID == 0 {
		return ErrInvalidPricingInput
	}
	if pricing.ChatRatePerMin < 0 || pricing.CallRatePerMin < 0 || pricing.VideoCallRatePerMin < 0 {
		return fmt.Errorf("%w: rates cannot be negative", ErrInvalidPricingInput)
	}

	// Use the repository's CreateOrUpdate for simplicity and robustness.
	err := s.repo.CreateOrUpdate(ctx, pricing)
	if err != nil {
		// Log unexpected errors?
		return fmt.Errorf("failed to set pricing in repository: %w", err)
	}
	return nil
}
