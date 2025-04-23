package models

import (
	"time"

	"gorm.io/gorm"
)

// ConsultationPricing stores the pricing rates for various consultation types
// linked to a specific entity (e.g., guide).
// Using int64 for rates assumes storing the value in the smallest currency unit (e.g., cents).
type ConsultationPricing struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// EntityType indicates the type of entity this pricing belongs to (e.g., "guide")
	EntityType string `gorm:"type:varchar(50);uniqueIndex:idx_entity_pricing,priority:1;not null" json:"entity_type"`

	// EntityID is the ID of the specific entity (e.g., the GuideID)
	EntityID uint `gorm:"uniqueIndex:idx_entity_pricing,priority:2;not null" json:"entity_id"`

	ChatRatePerMin      int64 `gorm:"type:bigint;default:0" json:"chat_rate_per_min"`
	CallRatePerMin      int64 `gorm:"type:bigint;default:0" json:"call_rate_per_min"`
	VideoCallRatePerMin int64 `gorm:"type:bigint;default:0" json:"video_call_rate_per_min"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Use gorm.DeletedAt for soft deletes, exclude from json by default
}

// TableName specifies the table name for GORM
func (ConsultationPricing) TableName() string {
	return "consultation_pricing"
}
