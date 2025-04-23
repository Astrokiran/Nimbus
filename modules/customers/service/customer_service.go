package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm" // Import gorm for ErrRecordNotFound

	"nimbus-service/modules/customers/models"
	"nimbus-service/modules/customers/repository"
	// "github.com/go-playground/validator/v10" // Example validator import
)

// CustomerService defines the interface for customer business logic.
type CustomerService interface {
	CreateNewCustomer(ctx context.Context, customer *models.Customer) (*models.Customer, error)
	FindCustomerByPhone(ctx context.Context, areaCode, mobileNumber string) (*models.Customer, error)
	// Add other business logic methods
}

// customerService implements the CustomerService interface.
type customerService struct {
	repo repository.CustomerRepository
	// validate *validator.Validate // Example validator instance
}

// NewCustomerService creates a new instance of the customer service.
func NewCustomerService(repo repository.CustomerRepository /*, validate *validator.Validate*/) CustomerService {
	return &customerService{
		repo: repo,
		// validate: validate,
	}
}

// CreateNewCustomer handles the business logic for creating a new customer.
// It performs validation and calls the repository.
func (s *customerService) CreateNewCustomer(ctx context.Context, customer *models.Customer) (*models.Customer, error) {
	// 1. Validation (using struct tags or custom logic)
	// Example using go-playground/validator (if integrated)
	// if err := s.validate.Struct(customer); err != nil {
	//  return nil, fmt.Errorf("validation failed: %w", err)
	// }
	// Basic manual validation for required fields (can be enhanced)
	if customer.AreaCode == "" || customer.MobileNumber == "" {
		return nil, errors.New("area code and mobile number are required")
	}

	// 2. Check if customer already exists (optional, depends on desired behavior)
	// existing, err := s.repo.GetCustomerByPhone(customer.AreaCode, customer.MobileNumber)
	// if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
	//  return nil, fmt.Errorf("failed to check for existing customer: %w", err)
	// }
	// if existing != nil {
	//  return nil, errors.New("customer with this phone number already exists")
	// }

	// 3. Call repository to create the customer
	err := s.repo.CreateCustomer(ctx, customer)
	if err != nil {
		// TODO: Handle specific errors like unique constraint violation more gracefully
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// The customer object passed in now contains the ID and timestamps from the DB
	return customer, nil
}

// FindCustomerByPhone handles the business logic for retrieving a customer by phone.
func (s *customerService) FindCustomerByPhone(ctx context.Context, areaCode, mobileNumber string) (*models.Customer, error) {
	if areaCode == "" || mobileNumber == "" {
		return nil, errors.New("area code and mobile number are required for lookup")
	}

	customer, err := s.repo.GetCustomerByPhone(ctx, areaCode, mobileNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Return a specific "not found" error that the handler layer can interpret
			return nil, ErrCustomerNotFound // Define this error below or in a shared errors package
		}
		// Log the unexpected error?
		return nil, fmt.Errorf("failed to find customer by phone: %w", err)
	}
	return customer, nil
}

// --- Define custom errors ---
var ErrCustomerNotFound = errors.New("customer not found")

// --- Add other service methods (Update, Delete logic) ---
