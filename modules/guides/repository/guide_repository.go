package repository

import (
	"context"
	"errors"

	"nimbus-service/modules/guides/models"

	"gorm.io/gorm"
)

// GuideRepository defines the interface for guide data operations
type GuideRepository interface {
	CreateGuide(ctx context.Context, guide *models.Guide) error
	GetGuideByID(ctx context.Context, guideID uint) (*models.Guide, error)
	GetGuideByPhone(ctx context.Context, areaCode, phoneNumber string) (*models.Guide, error)
	UpdateGuide(ctx context.Context, guide *models.Guide) error
	DeleteGuide(ctx context.Context, guideID uint) error
	ListGuides(ctx context.Context) ([]*models.Guide, error)
}

// gormGuideRepository implements the GuideRepository interface using GORM
type gormGuideRepository struct {
	db *gorm.DB
}

// NewGuideRepository creates a new instance of the GORM guide repository
func NewGuideRepository(db *gorm.DB) GuideRepository {
	return &gormGuideRepository{db: db}
}

// CreateGuide inserts a new guide record into the database
func (r *gormGuideRepository) CreateGuide(ctx context.Context, guide *models.Guide) error {
	result := r.db.WithContext(ctx).Create(guide)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetGuideByID retrieves a guide by their ID
func (r *gormGuideRepository) GetGuideByID(ctx context.Context, guideID uint) (*models.Guide, error) {
	var guide models.Guide
	result := r.db.WithContext(ctx).Where("guide_id = ?", guideID).First(&guide)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &guide, nil
}

// GetGuideByPhone retrieves a guide by their area code and phone number
func (r *gormGuideRepository) GetGuideByPhone(ctx context.Context, areaCode, phoneNumber string) (*models.Guide, error) {
	var guide models.Guide
	result := r.db.WithContext(ctx).Where("area_code = ? AND phone_number = ?", areaCode, phoneNumber).First(&guide)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &guide, nil
}

// UpdateGuide updates an existing guide record in the database
func (r *gormGuideRepository) UpdateGuide(ctx context.Context, guide *models.Guide) error {
	result := r.db.WithContext(ctx).Save(guide)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteGuide soft-deletes a guide record from the database
func (r *gormGuideRepository) DeleteGuide(ctx context.Context, guideID uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Guide{}, guideID)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// ListGuides retrieves all guides from the database
func (r *gormGuideRepository) ListGuides(ctx context.Context) ([]*models.Guide, error) {
	var guides []*models.Guide
	result := r.db.WithContext(ctx).Find(&guides)
	if result.Error != nil {
		return nil, result.Error
	}
	return guides, nil
}
