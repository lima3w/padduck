package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"padduck/repository"
)

func parseListOptions(c *fiber.Ctx) (page int, limit int, opts repository.ListOptions) {
	page = c.QueryInt("page", 0)
	limit = c.QueryInt("limit", 0)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 25
	}
	opts.Sort = strings.TrimSpace(c.Query("sort"))
	opts.Order = strings.TrimSpace(c.Query("order", c.Query("dir", "asc")))
	opts.Query = strings.TrimSpace(c.Query("q", c.Query("search")))
	opts.Status = strings.TrimSpace(c.Query("status"))
	opts.HideAvailable = c.Query("hide_available") == "true"
	return page, limit, opts
}
