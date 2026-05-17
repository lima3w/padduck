package handlers

import (
	"net/http"
	"testing"
)

func TestCustomerRoutes_AuthRequired(t *testing.T) {
	h := &Handler{}
	assertAuthRequired(t, []authRoute{
		{
			name:        "list customers",
			method:      http.MethodGet,
			routePath:   "/customers",
			requestPath: "/customers",
			handler:     h.ListCustomers,
		},
		{
			name:        "get customer",
			method:      http.MethodGet,
			routePath:   "/customers/:id",
			requestPath: "/customers/1",
			handler:     h.GetCustomer,
		},
		{
			name:        "create customer",
			method:      http.MethodPost,
			routePath:   "/customers",
			requestPath: "/customers",
			body:        `{"name":"Acme Corp"}`,
			handler:     h.CreateCustomer,
		},
		{
			name:        "update customer",
			method:      http.MethodPut,
			routePath:   "/customers/:id",
			requestPath: "/customers/1",
			body:        `{"name":"Acme Corp Updated"}`,
			handler:     h.UpdateCustomer,
		},
		{
			name:        "delete customer",
			method:      http.MethodDelete,
			routePath:   "/customers/:id",
			requestPath: "/customers/1",
			handler:     h.DeleteCustomer,
		},
	})
}
