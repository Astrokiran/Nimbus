package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"nimbus-service/modules/guides/dto"
	"nimbus-service/modules/guides/models"
	"nimbus-service/modules/guides/service"

	"github.com/go-chi/chi/v5"
)

// GuideHandler handles HTTP requests for guide operations
type GuideHandler struct {
	service service.GuideService
}

// NewGuideHandler creates a new instance of the guide handler
func NewGuideHandler(s service.GuideService) *GuideHandler {
	return &GuideHandler{
		service: s,
	}
}

// --- DTOs (Data Transfer Objects) ---

type CreateGuideRequest struct {
	Name        string `json:"name"`
	AreaCode    string `json:"area_code"`
	PhoneNumber string `json:"phone_number"`
	Gender      string `json:"gender,omitempty"`
	Skills      string `json:"skills,omitempty"`
	Languages   string `json:"languages,omitempty"`
	PhotoURL    string `json:"photo_url,omitempty"`
	IsActive    bool   `json:"is_active,omitempty"`
}

type UpdateGuideRequest struct {
	Name        string `json:"name,omitempty"`
	AreaCode    string `json:"area_code,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Skills      string `json:"skills,omitempty"`
	Languages   string `json:"languages,omitempty"`
	PhotoURL    string `json:"photo_url,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"` // Pointer to differentiate between not provided and false
}

// ConsultationPricingResponse and GuideResponse are removed from here

// mapGuideToResponse converts a guide model to a response DTO
// Note: This function currently doesn't populate pricing. It will be handled
// within the service layer for GetGuideByID according to Option C.
// This might need adjustments if other handlers call it directly and need pricing.
func mapGuideToResponse(guide *models.Guide) *dto.GuideResponse {
	// Initial response without pricing
	response := &dto.GuideResponse{
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
		Pricing:     nil, // Explicitly nil for now
	}
	// Pricing will be added by the service layer where needed (e.g., FindGuideByID)
	return response
}

// CreateGuide handles POST requests to create a new guide
// Route: POST /guides
func (h *GuideHandler) CreateGuide(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateGuideRequest
	if r.Body == nil {
		http.Error(w, "Missing request body", http.StatusBadRequest)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Map request to model
	guideModel := &models.Guide{
		Name:        req.Name,
		AreaCode:    req.AreaCode,
		PhoneNumber: req.PhoneNumber,
		Gender:      req.Gender,
		Skills:      req.Skills,
		Languages:   req.Languages,
		PhotoURL:    req.PhotoURL,
		IsActive:    req.IsActive,
	}

	createdGuide, err := h.service.CreateNewGuide(ctx, guideModel)
	if err != nil {
		errMsg := err.Error()
		if errors.Is(err, service.ErrInvalidInput) {
			http.Error(w, "Validation Error: "+errMsg, http.StatusBadRequest)
		} else if strings.Contains(errMsg, "already exists") {
			http.Error(w, errMsg, http.StatusConflict)
		} else {
			http.Error(w, "Failed to create guide: "+errMsg, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mapGuideToResponse(createdGuide))
}

// GetGuideByID handles GET requests to retrieve a guide by ID
// Route: GET /guides/{guideID}
func (h *GuideHandler) GetGuideByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	guideIDParam := chi.URLParam(r, "guideID")
	guideID, err := strconv.ParseUint(guideIDParam, 10, 32)
	if err != nil {
		http.Error(w, "Invalid guide ID", http.StatusBadRequest)
		return
	}

	guideResponse, err := h.service.FindGuideByID(ctx, uint(guideID))
	if err != nil {
		if errors.Is(err, service.ErrGuideNotFound) {
			http.Error(w, "Guide not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve guide: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(guideResponse)
}

// GetGuideByPhone handles GET requests to retrieve a guide by phone
// Route: GET /guides/phone/{areaCode}/{phoneNumber}
func (h *GuideHandler) GetGuideByPhone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	areaCode := chi.URLParam(r, "areaCode")
	phoneNumber := chi.URLParam(r, "phoneNumber")

	if areaCode == "" || phoneNumber == "" {
		http.Error(w, "Area code and phone number parameters are required", http.StatusBadRequest)
		return
	}

	guideModel, err := h.service.FindGuideByPhone(ctx, areaCode, phoneNumber)
	if err != nil {
		if errors.Is(err, service.ErrGuideNotFound) {
			http.Error(w, "Guide not found", http.StatusNotFound)
		} else if errors.Is(err, service.ErrInvalidInput) {
			http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to retrieve guide: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapGuideToResponse(guideModel))
}

// UpdateGuide handles PUT requests to update an existing guide
// Route: PUT /guides/{guideID}
func (h *GuideHandler) UpdateGuide(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	guideIDParam := chi.URLParam(r, "guideID")
	guideID, err := strconv.ParseUint(guideIDParam, 10, 32)
	if err != nil {
		http.Error(w, "Invalid guide ID", http.StatusBadRequest)
		return
	}

	var req UpdateGuideRequest
	if r.Body == nil {
		http.Error(w, "Missing request body", http.StatusBadRequest)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Fetch the existing guide *model* using the new service method
	existingGuideModel, err := h.service.GetGuideModelByID(ctx, uint(guideID))
	if err != nil {
		if errors.Is(err, service.ErrGuideNotFound) {
			http.Error(w, "Guide not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve guide for update: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Update only the fields that were provided in the request
	updated := false
	if req.Name != "" && existingGuideModel.Name != req.Name {
		existingGuideModel.Name = req.Name
		updated = true
	}
	if req.AreaCode != "" && existingGuideModel.AreaCode != req.AreaCode {
		existingGuideModel.AreaCode = req.AreaCode
		updated = true
	}
	if req.PhoneNumber != "" && existingGuideModel.PhoneNumber != req.PhoneNumber {
		existingGuideModel.PhoneNumber = req.PhoneNumber
		updated = true
	}
	if req.Gender != "" && existingGuideModel.Gender != req.Gender {
		existingGuideModel.Gender = req.Gender
		updated = true
	}
	if req.Skills != "" && existingGuideModel.Skills != req.Skills {
		existingGuideModel.Skills = req.Skills
		updated = true
	}
	if req.Languages != "" && existingGuideModel.Languages != req.Languages {
		existingGuideModel.Languages = req.Languages
		updated = true
	}
	if req.PhotoURL != "" && existingGuideModel.PhotoURL != req.PhotoURL {
		existingGuideModel.PhotoURL = req.PhotoURL
		updated = true
	}
	if req.IsActive != nil && existingGuideModel.IsActive != *req.IsActive {
		existingGuideModel.IsActive = *req.IsActive
		updated = true
	}

	var updatedGuide *models.Guide
	if updated {
		updatedGuide, err = h.service.UpdateGuideDetails(ctx, existingGuideModel)
		if err != nil {
			errMsg := err.Error()
			if errors.Is(err, service.ErrInvalidInput) {
				http.Error(w, "Validation Error during update: "+errMsg, http.StatusBadRequest)
			} else if errors.Is(err, service.ErrGuideNotFound) {
				http.Error(w, "Guide not found during update", http.StatusNotFound)
			} else if strings.Contains(errMsg, "already exists") {
				http.Error(w, errMsg, http.StatusConflict)
			} else {
				http.Error(w, "Failed to update guide: "+errMsg, http.StatusInternalServerError)
			}
			return
		}
	} else {
		updatedGuide = existingGuideModel
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapGuideToResponse(updatedGuide))
}

// DeleteGuide handles DELETE requests to remove a guide
// Route: DELETE /guides/{guideID}
func (h *GuideHandler) DeleteGuide(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	guideIDParam := chi.URLParam(r, "guideID")
	guideID, err := strconv.ParseUint(guideIDParam, 10, 32)
	if err != nil {
		http.Error(w, "Invalid guide ID", http.StatusBadRequest)
		return
	}

	err = h.service.RemoveGuide(ctx, uint(guideID))
	if err != nil {
		if errors.Is(err, service.ErrGuideNotFound) {
			http.Error(w, "Guide not found", http.StatusNotFound)
		} else if errors.Is(err, service.ErrInvalidInput) {
			http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to delete guide: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListGuides handles GET requests to list all guides
// Route: GET /guides
func (h *GuideHandler) ListGuides(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Service GetAllGuides now returns the DTO slice directly
	guideResponses, err := h.service.GetAllGuides(ctx)
	if err != nil {
		// Log error?
		http.Error(w, "Failed to retrieve guides: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Remove the manual mapping loop as the service already prepared the DTOs
	// guideResponses := make([]*dto.GuideResponse, len(guides))
	// for i, guide := range guides {
	// 	guideResponses[i] = mapGuideToResponse(guide) // mapGuideToResponse might become obsolete or used differently
	// }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(guideResponses) // Encode the DTO slice directly
}
