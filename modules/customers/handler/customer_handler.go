package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings" // Added for path splitting placeholder
	"time"    // Added for time parsing/formatting

	"github.com/go-chi/chi/v5" // Import chi
	// "github.com/gorilla/mux" // Example: if using mux router for path parameters

	"nimbus-service/internal/middleware" // Import middleware to access context keys
	"nimbus-service/modules/customers/models"
	"nimbus-service/modules/customers/service"
	// Import logger if available e.g. "nimbus-service/internal/logger"
)

// CustomerHandler handles HTTP requests for customer operations.
type CustomerHandler struct {
	service service.CustomerService
	// logger  zerolog.Logger // Example logger
}

// NewCustomerHandler creates a new instance of the customer handler.
func NewCustomerHandler(s service.CustomerService /*, log zerolog.Logger*/) *CustomerHandler {
	return &CustomerHandler{
		service: s,
		// logger: log.With().Str("component", "CustomerHandler").Logger(), // Contextual logger
	}
}

// --- DTOs (Data Transfer Objects) ---
// Optional: Define specific structs for request/response bodies
// for better separation and control over API schema.

type CreateCustomerRequest struct {
	AreaCode       string `json:"area_code"`
	MobileNumber   string `json:"mobile_number"`
	Name           string `json:"name,omitempty"`
	Gender         string `json:"gender,omitempty"`
	DateOfBirth    string `json:"date_of_birth,omitempty"` // Expect YYYY-MM-DD
	TimeOfBirth    string `json:"time_of_birth,omitempty"` // Expect HH:MM:SS
	PlaceOfBirth   string `json:"place_of_birth,omitempty"`
	CurrentAddress string `json:"current_address,omitempty"`
	City           string `json:"city,omitempty"`
	State          string `json:"state,omitempty"`
	Country        string `json:"country,omitempty"`
	Pincode        string `json:"pincode,omitempty"`
}

type CustomerResponse struct {
	// ID             uint   `json:"id"`
	CustomerID     uint   `json:"customer_id"`
	AreaCode       string `json:"area_code"`
	MobileNumber   string `json:"mobile_number"`
	Name           string `json:"name,omitempty"`
	Gender         string `json:"gender,omitempty"`
	DateOfBirth    string `json:"date_of_birth,omitempty"` // Format as YYYY-MM-DD
	TimeOfBirth    string `json:"time_of_birth,omitempty"` // Format as HH:MM:SS
	PlaceOfBirth   string `json:"place_of_birth,omitempty"`
	CurrentAddress string `json:"current_address,omitempty"`
	City           string `json:"city,omitempty"`
	State          string `json:"state,omitempty"`
	Country        string `json:"country,omitempty"`
	Pincode        string `json:"pincode,omitempty"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// Helper to map model to response DTO
func mapCustomerToResponse(customer *models.Customer) *CustomerResponse {
	dob := ""
	if customer.DateOfBirth != nil {
		dob = customer.DateOfBirth.Format("2006-01-02")
	}
	tob := ""
	if customer.TimeOfBirth != nil {
		// Assuming TimeOfBirth is stored as string "HH:MM:SS" pointer in model based on previous definition
		tob = *customer.TimeOfBirth
	}

	return &CustomerResponse{
		// ID:             customer.ID,
		CustomerID:     customer.CustomerID,
		AreaCode:       customer.AreaCode,
		MobileNumber:   customer.MobileNumber,
		Name:           customer.Name,
		Gender:         customer.Gender,
		DateOfBirth:    dob,
		TimeOfBirth:    tob,
		PlaceOfBirth:   customer.PlaceOfBirth,
		CurrentAddress: customer.CurrentAddress,
		City:           customer.City,
		State:          customer.State,
		Country:        customer.Country,
		Pincode:        customer.Pincode,
		CreatedAt:      customer.CreatedAt.Format(time.RFC3339), // Use RFC3339 for standard timestamp format
		UpdatedAt:      customer.UpdatedAt.Format(time.RFC3339), // Use RFC3339 for standard timestamp format
	}
}

// CreateCustomer handles POST requests to create a new customer.
// Expected Route: POST /customers
func (h *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateCustomerRequest
	// Ensure request body is not nil
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

	// --- Data Conversion/Mapping from Request DTO to Model ---
	var dob *time.Time
	if req.DateOfBirth != "" {
		parsedDate, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err != nil {
			http.Error(w, "Invalid date_of_birth format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		dob = &parsedDate
	}
	var tob *string
	if req.TimeOfBirth != "" {
		// Basic validation for HH:MM:SS format
		if _, err := time.Parse("15:04:05", req.TimeOfBirth); err != nil {
			http.Error(w, "Invalid time_of_birth format, use HH:MM:SS", http.StatusBadRequest)
			return
		}
		tob = &req.TimeOfBirth
	}

	// --- Retrieve User from context using helper (Updated) ---
	authenticatedUser := middleware.GetUserFromContext(ctx)
	if authenticatedUser == nil || authenticatedUser.ID == 0 {
		// Log this error - indicates an issue with auth middleware or request context
		// h.logger.Error().Msg("Could not retrieve valid User from context")
		http.Error(w, "Internal Server Error: Unable to identify authenticated user", http.StatusInternalServerError)
		return
	}

	customerModel := &models.Customer{
		UserID:         authenticatedUser.ID, // Assign the retrieved UserID
		AreaCode:       req.AreaCode,
		MobileNumber:   req.MobileNumber,
		Name:           req.Name,
		Gender:         req.Gender,
		DateOfBirth:    dob,
		TimeOfBirth:    tob, // Store as string pointer in model
		PlaceOfBirth:   req.PlaceOfBirth,
		CurrentAddress: req.CurrentAddress,
		City:           req.City,
		State:          req.State,
		Country:        req.Country,
		Pincode:        req.Pincode,
		// Auth related fields like OtpSecret are usually handled by the service/repo
	}

	createdCustomer, err := h.service.CreateNewCustomer(ctx, customerModel)
	if err != nil {
		// Example error mapping (can be more sophisticated)
		errMsg := err.Error()
		// h.logger.Warn().Err(err).Msg("Service error during customer creation") // Example logging
		if strings.Contains(errMsg, "required") {
			http.Error(w, "Validation Error: "+errMsg, http.StatusBadRequest)
		} else if strings.Contains(errMsg, "duplicate key value violates unique constraint") || strings.Contains(errMsg, "already exists") {
			// Catch potential unique constraint errors from DB or explicit checks in service
			http.Error(w, "Customer with this phone number already exists", http.StatusConflict) // 409 Conflict
		} else {
			// Log internal errors for debugging
			// h.logger.Error().Err(err).Msg("Internal server error during customer creation")
			http.Error(w, "Failed to create customer due to an internal error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mapCustomerToResponse(createdCustomer))
}

// GetCustomerByPhone handles GET requests to retrieve a customer by phone.
// Expected Route: GET /customers/{areaCode}/{mobileNumber}
// IMPORTANT: Parameter extraction below is a PLACEHOLDER and MUST be replaced
// with logic specific to the HTTP router being used (e.g., Gin, Echo, Chi, Mux).
func (h *CustomerHandler) GetCustomerByPhone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// --- Router-Specific Parameter Extraction (Using Chi) ---
	areaCode := chi.URLParam(r, "areaCode")
	mobileNumber := chi.URLParam(r, "mobileNumber")
	// --- End Chi Parameter Extraction ---

	if areaCode == "" || mobileNumber == "" {
		http.Error(w, "Area code and mobile number path parameters are required and cannot be empty", http.StatusBadRequest)
		return
	}

	customer, err := h.service.FindCustomerByPhone(ctx, areaCode, mobileNumber)
	if err != nil {
		if errors.Is(err, service.ErrCustomerNotFound) {
			http.Error(w, "Customer not found", http.StatusNotFound)
		} else {
			// Log internal errors for debugging
			// h.logger.Error().Err(err).Str("areaCode", areaCode).Str("mobileNumber", mobileNumber).Msg("Failed to get customer by phone")
			http.Error(w, "Failed to retrieve customer due to an internal error", http.StatusInternalServerError) // Avoid leaking specific errors
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mapCustomerToResponse(customer))
}

// --- Add handlers for Update, Delete etc. with ctx ---
