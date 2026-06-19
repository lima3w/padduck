package handlers

import "github.com/gofiber/fiber/v2"

// V2Meta is the pagination/envelope metadata included in every v2 list response.
type V2Meta struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

// V2List wraps a slice payload with the standard v2 list envelope.
func V2List(data any, meta V2Meta) fiber.Map {
	return fiber.Map{"data": data, "meta": meta}
}

// V2Item wraps a single-resource payload in the standard v2 envelope.
func V2Item(data any) fiber.Map {
	return fiber.Map{"data": data}
}

// addDeprecationHeaders adds the standard RFC 8594 deprecation headers pointing
// consumers at the v2 successor endpoint.
func addDeprecationHeaders(c *fiber.Ctx, successorPath string) {
	c.Set("Deprecation", "true")
	c.Set("Link", "<"+successorPath+">; rel=\"successor-version\"")
}
