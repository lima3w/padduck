package handlers

import (
	"net/http"
	"testing"
)

func TestAutonomousSystemRoutes_AuthRequired(t *testing.T) {
	h := minHandler()
	assertAuthRequired(t, []authRoute{
		{
			name:        "list autonomous systems",
			method:      http.MethodGet,
			routePath:   "/autonomous-systems",
			requestPath: "/autonomous-systems",
			handler:     h.ListAutonomousSystems,
		},
		{
			name:        "get autonomous system",
			method:      http.MethodGet,
			routePath:   "/autonomous-systems/:id",
			requestPath: "/autonomous-systems/1",
			handler:     h.GetAutonomousSystem,
		},
		{
			name:        "create autonomous system",
			method:      http.MethodPost,
			routePath:   "/autonomous-systems",
			requestPath: "/autonomous-systems",
			body:        `{"asn":65000,"name":"Test AS"}`,
			handler:     h.CreateAutonomousSystem,
		},
		{
			name:        "update autonomous system",
			method:      http.MethodPut,
			routePath:   "/autonomous-systems/:id",
			requestPath: "/autonomous-systems/1",
			body:        `{"asn":65000,"name":"Updated AS"}`,
			handler:     h.UpdateAutonomousSystem,
		},
		{
			name:        "delete autonomous system",
			method:      http.MethodDelete,
			routePath:   "/autonomous-systems/:id",
			requestPath: "/autonomous-systems/1",
			handler:     h.DeleteAutonomousSystem,
		},
	})
}
