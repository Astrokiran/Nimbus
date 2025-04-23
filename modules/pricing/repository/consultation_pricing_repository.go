package repository

import (
	"context"
	"errors"

	"nimbus-service/modules/pricing/models" // Adjust import path if necessary

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ConsultationPricingRepository defines the interface for pricing data operations.
type ConsultationPricingRepository interface {
	GetByEntity(ctx context.Context, entityType string, entityID uint) (*models.ConsultationPricing, error)
	GetPricingByEntityIDs(ctx context.Context, entityType string, entityIDs []uint) ([]*models.ConsultationPricing, error)
	Create(ctx context.Context, pricing *models.ConsultationPricing) error
	Update(ctx context.Context, pricing *models.ConsultationPricing) error
	// CreateOrUpdate creates a new record or updates an existing one based on EntityType and EntityID.
	CreateOrUpdate(ctx context.Context, pricing *models.ConsultationPricing) error
}

// gormConsultationPricingRepository implements the ConsultationPricingRepository interface using GORM.
type gormConsultationPricingRepository struct {
	db *gorm.DB
}

// NewConsultationPricingRepository creates a new instance of the GORM pricing repository.
func NewConsultationPricingRepository(db *gorm.DB) ConsultationPricingRepository {
	return &gormConsultationPricingRepository{db: db}
}

// GetByEntity retrieves pricing for a specific entity type and ID.
func (r *gormConsultationPricingRepository) GetByEntity(ctx context.Context, entityType string, entityID uint) (*models.ConsultationPricing, error) {
	var pricing models.ConsultationPricing
	result := r.db.WithContext(ctx).Where("entity_type = ? AND entity_id = ?", entityType, entityID).First(&pricing)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound // Return specific error for not found
		}
		return nil, result.Error // Return other DB errors
	}
	return &pricing, nil
}

// GetPricingByEntityIDs retrieves pricing for multiple entities of the same type.
func (r *gormConsultationPricingRepository) GetPricingByEntityIDs(ctx context.Context, entityType string, entityIDs []uint) ([]*models.ConsultationPricing, error) {
	var pricingList []*models.ConsultationPricing
	if len(entityIDs) == 0 {
		return pricingList, nil // Return empty slice if no IDs are provided
	}
	result := r.db.WithContext(ctx).Where("entity_type = ? AND entity_id IN ?", entityType, entityIDs).Find(&pricingList)
	if result.Error != nil {
		// Don't treat RecordNotFound as an error here, just return an empty list if nothing is found
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return pricingList, nil
		}
		return nil, result.Error // Return other DB errors
	}
	return pricingList, nil
}

// Create inserts a new pricing record into the database.
func (r *gormConsultationPricingRepository) Create(ctx context.Context, pricing *models.ConsultationPricing) error {
	result := r.db.WithContext(ctx).Create(pricing)
	return result.Error
}

// Update updates an existing pricing record in the database.
// It assumes the record (identified by ID or unique constraint) exists.
func (r *gormConsultationPricingRepository) Update(ctx context.Context, pricing *models.ConsultationPricing) error {
	if pricing.ID == 0 {
		// Attempt to find by entity type/id if ID is not set
		existing, err := r.GetByEntity(ctx, pricing.EntityType, pricing.EntityID)
		if err != nil {
			return err // Not found or other error
		}
		pricing.ID = existing.ID // Set ID for update
	}
	// Use Save to update all fields or Updates for specific fields
	result := r.db.WithContext(ctx).Save(pricing)
	return result.Error
}

// CreateOrUpdate creates a pricing record if it doesn't exist (based on EntityType/EntityID),
// or updates it if it does.
func (r *gormConsultationPricingRepository) CreateOrUpdate(ctx context.Context, pricing *models.ConsultationPricing) error {
	// Use GORM's OnConflict clause for upsert functionality
	// This assumes your database supports it (like PostgreSQL 9.5+)
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "entity_type"}, {Name: "entity_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"chat_rate_per_min", "call_rate_per_min", "video_call_rate_per_min", "updated_at"}),
		// Use DoNothing if you only want to create if not exists: DoNothing: true
	}).Create(pricing)

	return result.Error
}
