package dto

// ConsultationPricingResponse holds the pricing details for the API response.
// Using int64 assuming currency is stored in smallest unit (e.g., cents).
type ConsultationPricingResponse struct {
	ChatRatePerMin      int64 `json:"chat_rate_per_min"`
	CallRatePerMin      int64 `json:"call_rate_per_min"`
	VideoCallRatePerMin int64 `json:"video_call_rate_per_min"`
}

// GuideResponse is the DTO for guide details including pricing.
type GuideResponse struct {
	GuideID     uint                         `json:"guide_id"`
	Name        string                       `json:"name"`
	AreaCode    string                       `json:"area_code"`
	PhoneNumber string                       `json:"phone_number"`
	Gender      string                       `json:"gender,omitempty"`
	Skills      string                       `json:"skills,omitempty"`
	Languages   string                       `json:"languages,omitempty"`
	PhotoURL    string                       `json:"photo_url,omitempty"`
	IsActive    bool                         `json:"is_active"`
	CreatedAt   string                       `json:"created_at"` // Consider using time.Time if consumers prefer
	UpdatedAt   string                       `json:"updated_at"` // Consider using time.Time if consumers prefer
	Pricing     *ConsultationPricingResponse `json:"pricing,omitempty"`
}
