package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"ipam-next/models"
)

// permUser returns a user with ID=0 so CheckPermission returns permission
// denied without touching the nil service repository, giving a clean 403.
func permUser() *models.User { return &models.User{ID: 0, Role: "user"} }

func testApp(method, path string, handler fiber.Handler, u *models.User) *fiber.App {
	app := fiber.New()
	if u != nil {
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user", u)
			return c.Next()
		})
	}
	switch method {
	case http.MethodGet:
		app.Get(path, handler)
	case http.MethodPost:
		app.Post(path, handler)
	case http.MethodPut:
		app.Put(path, handler)
	case http.MethodDelete:
		app.Delete(path, handler)
	}
	return app
}

func deviceApp(h *Handler, method, path string, handler fiber.Handler) *fiber.App {
	return testApp(method, path, handler, nil)
}

func deviceAppAs(h *Handler, method, path string, handler fiber.Handler, u *models.User) *fiber.App {
	return testApp(method, path, handler, u)
}

type authRoute struct {
	name        string
	method      string
	routePath   string
	requestPath string
	body        string
	handler     fiber.Handler
}

func assertAuthRequired(t *testing.T, routes []authRoute) {
	t.Helper()

	for _, route := range routes {
		route := route
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()

			cases := []struct {
				name string
				user *models.User
				want int
			}{
				{"no user", nil, fiber.StatusUnauthorized},
				{"no permission", permUser(), fiber.StatusForbidden},
			}

			for _, tc := range cases {
				tc := tc
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					app := testApp(route.method, route.routePath, route.handler, tc.user)
					req := httptest.NewRequest(route.method, route.requestPath, strings.NewReader(route.body))
					if route.body != "" {
						req.Header.Set("Content-Type", "application/json")
					}
					resp, err := app.Test(req)
					assert.NoError(t, err)
					assert.Equal(t, tc.want, resp.StatusCode)
				})
			}
		})
	}
}
