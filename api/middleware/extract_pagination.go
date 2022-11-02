package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/pagination"
	"net/http"
)

func ExtractPagination(c *fiber.Ctx) error {
	if c.Method() != http.MethodGet {
		return c.Next()
	}
	pagination.ExtractPagination(c)
	return c.Next()
}
