package router

import (
	"nimbus-service/modules/customers/handler" // Adjust import path if needed

	"github.com/go-chi/chi/v5"
)

// RegisterCustomerRoutes sets up the routes for the customer module.
func RegisterCustomerRoutes(router chi.Router, h *handler.CustomerHandler) {
	// Define routes specific to the customer module under a group (e.g., /customers)
	router.Route("/customers", func(r chi.Router) {
		r.Post("/", h.CreateCustomer)                             // POST /customers
		r.Get("/{areaCode}/{mobileNumber}", h.GetCustomerByPhone) // GET /customers/{areaCode}/{mobileNumber}

		// Add routes for Update, Delete, List etc. here
		// Example:
		// r.Put("/{customerID}", h.UpdateCustomer)
		// r.Delete("/{customerID}", h.DeleteCustomer)
		// r.Get("/", h.ListCustomers)
	})
}
