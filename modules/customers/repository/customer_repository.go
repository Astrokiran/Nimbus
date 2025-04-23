package repository

import (
	"context"
	"errors"

	"nimbus-service/modules/customers/models" // Adjust import path if needed

	"gorm.io/gorm"
)

// CustomerRepository defines the interface for customer data operations.
type CustomerRepository interface {
	CreateCustomer(ctx context.Context, customer *models.Customer) error
	GetCustomerByPhone(ctx context.Context, areaCode, mobileNumber string) (*models.Customer, error)
	// Add other methods like UpdateCustomer, DeleteCustomer etc. as needed
}

// gormCustomerRepository implements the CustomerRepository interface using GORM.
type gormCustomerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository creates a new instance of the GORM customer repository.
func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &gormCustomerRepository{db: db}
}

// CreateCustomer inserts a new customer record into the database.
func (r *gormCustomerRepository) CreateCustomer(ctx context.Context, customer *models.Customer) error {
	result := r.db.WithContext(ctx).Create(customer)
	if result.Error != nil {
		// TODO: Add specific error handling, e.g., for unique constraint violation
		return result.Error
	}
	return nil
}

// GetCustomerByPhone retrieves a customer by their area code and mobile number.
// Returns the customer and nil error if found.
// Returns nil and gorm.ErrRecordNotFound if not found.
// Returns nil and other error on database issues.
func (r *gormCustomerRepository) GetCustomerByPhone(ctx context.Context, areaCode, mobileNumber string) (*models.Customer, error) {
	var customer models.Customer
	result := r.db.WithContext(ctx).Where("area_code = ? AND mobile_number = ?", areaCode, mobileNumber).First(&customer)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound // Explicitly return GORM's not found error
		}
		return nil, result.Error // Return other database errors
	}
	return &customer, nil
}

// --- Implement UpdateCustomer, DeleteCustomer etc. here ---
