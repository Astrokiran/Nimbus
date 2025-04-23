package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"nimbus-service/modules/guides/dto"
	"nimbus-service/modules/guides/models"
	"nimbus-service/modules/guides/repository"
	pricing "nimbus-service/modules/pricing/service"

	"gorm.io/gorm"
)

// GuideService defines the interface for guide business logic
type GuideService interface {
	CreateNewGuide(ctx context.Context, guide *models.Guide) (*models.Guide, error)
	FindGuideByID(ctx context.Context, guideID uint) (*dto.GuideResponse, error)
	GetGuideModelByID(ctx context.Context, guideID uint) (*models.Guide, error)
	FindGuideByPhone(ctx context.Context, areaCode, phoneNumber string) (*models.Guide, error)
	UpdateGuideDetails(ctx context.Context, guide *models.Guide) (*models.Guide, error)
	RemoveGuide(ctx context.Context, guideID uint) error
	GetAllGuides(ctx context.Context) ([]*dto.GuideResponse, error)
}

// Custom errors
var (
	ErrGuideNotFound = errors.New("guide not found")
	ErrInvalidInput  = errors.New("invalid input")
)

// guideService implements the GuideService interface
type guideService struct {
	repo           repository.GuideRepository
	pricingService pricing.PricingService
}

// NewGuideService creates a new instance of the guide service
func NewGuideService(repo repository.GuideRepository, pricingService pricing.PricingService) GuideService {
	return &guideService{
		repo:           repo,
		pricingService: pricingService,
	}
}

// CreateNewGuide handles the business logic for creating a new guide
func (s *guideService) CreateNewGuide(ctx context.Context, guide *models.Guide) (*models.Guide, error) {
	// Basic validation
	if guide.Name == "" || guide.AreaCode == "" || guide.PhoneNumber == "" {
		return nil, ErrInvalidInput
	}

	// Check if guide already exists with the same phone
	existing, err := s.repo.GetGuideByPhone(ctx, guide.AreaCode, guide.PhoneNumber)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check for existing guide: %w", err)
	}
	if existing != nil {
		return nil, errors.New("guide with this phone number already exists")
	}

	// Create the guide
	err = s.repo.CreateGuide(ctx, guide)
	if err != nil {
		return nil, fmt.Errorf("failed to create guide: %w", err)
	}

	return guide, nil
}

// FindGuideByID retrieves guide details and pricing, returning a DTO.
func (s *guideService) FindGuideByID(ctx context.Context, guideID uint) (*dto.GuideResponse, error) {
	if guideID == 0 {
		return nil, ErrInvalidInput
	}

	// Fetch the core guide model
	guide, err := s.repo.GetGuideByID(ctx, guideID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGuideNotFound
		}
		return nil, fmt.Errorf("failed to find guide by ID from repository: %w", err)
	}

	// Now construct the DTO using the model
	guideResponse := &dto.GuideResponse{
		GuideID:     guide.GuideID,
		Name:        guide.Name,
		AreaCode:    guide.AreaCode,
		PhoneNumber: guide.PhoneNumber,
		Gender:      guide.Gender,
		Skills:      guide.Skills,
		Languages:   guide.Languages,
		PhotoURL:    guide.PhotoURL,
		IsActive:    guide.IsActive,
		CreatedAt:   guide.CreatedAt.Format(time.RFC3339), // Consider moving formatting to handler/DTO
		UpdatedAt:   guide.UpdatedAt.Format(time.RFC3339), // Consider moving formatting to handler/DTO
		// Pricing initially nil
	}

	// Fetch the pricing information
	pricingInfo, err := s.pricingService.GetConsultationPricing(ctx, "guide", guide.GuideID)
	if err != nil {
		// If pricing is simply not found, we might not treat it as a fatal error.
		// Log it and return the guide DTO without pricing info.
		if errors.Is(err, pricing.ErrPricingNotFound) {
			// Log this event? e.g., s.logger.Warn().Uint("guideID", guideID).Msg("Pricing info not found for guide")
			return guideResponse, nil // Return DTO without pricing
		} else {
			// For other errors fetching pricing, return the error
			// Log this error?
			return nil, fmt.Errorf("failed to get pricing for guide %d: %w", guideID, err)
		}
	}

	// If pricing info was found, populate the DTO
	guideResponse.Pricing = &dto.ConsultationPricingResponse{
		ChatRatePerMin:      pricingInfo.ChatRatePerMin,
		CallRatePerMin:      pricingInfo.CallRatePerMin,
		VideoCallRatePerMin: pricingInfo.VideoCallRatePerMin,
	}

	return guideResponse, nil
}

// FindGuideByPhone retrieves a guide by their phone details
func (s *guideService) FindGuideByPhone(ctx context.Context, areaCode, phoneNumber string) (*models.Guide, error) {
	if areaCode == "" || phoneNumber == "" {
		return nil, ErrInvalidInput
	}

	guide, err := s.repo.GetGuideByPhone(ctx, areaCode, phoneNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGuideNotFound
		}
		return nil, fmt.Errorf("failed to find guide by phone: %w", err)
	}
	return guide, nil
}

// UpdateGuideDetails updates an existing guide's information
func (s *guideService) UpdateGuideDetails(ctx context.Context, guide *models.Guide) (*models.Guide, error) {
	if guide.GuideID == 0 {
		return nil, ErrInvalidInput
	}

	// Check if guide exists
	existing, err := s.repo.GetGuideByID(ctx, guide.GuideID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGuideNotFound
		}
		return nil, fmt.Errorf("failed to check existing guide: %w", err)
	}

	// If phone number is being changed, check if it would conflict with another guide
	if existing.AreaCode != guide.AreaCode || existing.PhoneNumber != guide.PhoneNumber {
		duplicate, err := s.repo.GetGuideByPhone(ctx, guide.AreaCode, guide.PhoneNumber)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check for duplicate phone: %w", err)
		}
		if duplicate != nil && duplicate.GuideID != guide.GuideID {
			return nil, errors.New("another guide with this phone number already exists")
		}
	}

	// Update the guide
	err = s.repo.UpdateGuide(ctx, guide)
	if err != nil {
		return nil, fmt.Errorf("failed to update guide: %w", err)
	}

	return guide, nil
}

// RemoveGuide handles the removal of a guide
func (s *guideService) RemoveGuide(ctx context.Context, guideID uint) error {
	if guideID == 0 {
		return ErrInvalidInput
	}

	// Check if guide exists
	_, err := s.repo.GetGuideByID(ctx, guideID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGuideNotFound
		}
		return fmt.Errorf("failed to check existing guide: %w", err)
	}

	// Delete the guide
	err = s.repo.DeleteGuide(ctx, guideID)
	if err != nil {
		return fmt.Errorf("failed to delete guide: %w", err)
	}

	return nil
}

// GetAllGuides retrieves all guides with their pricing information.
func (s *guideService) GetAllGuides(ctx context.Context) ([]*dto.GuideResponse, error) {
	// 1. Fetch all guide models
	guides, err := s.repo.ListGuides(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list guides from repository: %w", err)
	}

	if len(guides) == 0 {
		return []*dto.GuideResponse{}, nil // Return empty slice, not nil
	}

	// 2. Collect guide IDs
	guideIDs := make([]uint, 0, len(guides))
	for _, g := range guides {
		guideIDs = append(guideIDs, g.GuideID)
	}

	// 3. Fetch pricing for all collected IDs in one go
	pricingMap, err := s.pricingService.GetPricingByEntityIDs(ctx, "guide", guideIDs)
	if err != nil {
		// Log this error? Non-fatal? For now, treat as error but maybe return partial data?
		// Let's return error for now. Could also log and continue without pricing.
		return nil, fmt.Errorf("failed to get pricing for guides: %w", err)
	}

	// 4. Construct DTOs
	guideResponses := make([]*dto.GuideResponse, 0, len(guides))
	for _, guide := range guides {
		resp := &dto.GuideResponse{
			GuideID:     guide.GuideID,
			Name:        guide.Name,
			AreaCode:    guide.AreaCode,
			PhoneNumber: guide.PhoneNumber,
			Gender:      guide.Gender,
			Skills:      guide.Skills,
			Languages:   guide.Languages,
			PhotoURL:    guide.PhotoURL,
			IsActive:    guide.IsActive,
			CreatedAt:   guide.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   guide.UpdatedAt.Format(time.RFC3339),
		}

		// Check if pricing exists for this guide in the map
		if pricingInfo, ok := pricingMap[guide.GuideID]; ok {
			resp.Pricing = &dto.ConsultationPricingResponse{
				ChatRatePerMin:      pricingInfo.ChatRatePerMin,
				CallRatePerMin:      pricingInfo.CallRatePerMin,
				VideoCallRatePerMin: pricingInfo.VideoCallRatePerMin,
			}
		}
		guideResponses = append(guideResponses, resp)
	}

	return guideResponses, nil
}

// GetGuideModelByID retrieves the guide model by ID
func (s *guideService) GetGuideModelByID(ctx context.Context, guideID uint) (*models.Guide, error) {
	if guideID == 0 {
		return nil, ErrInvalidInput
	}

	guide, err := s.repo.GetGuideByID(ctx, guideID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGuideNotFound
		}
		return nil, fmt.Errorf("failed to get guide model by ID: %w", err)
	}
	return guide, nil
}
