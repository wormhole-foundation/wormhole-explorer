package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// ExtractPagination parses pagination-related query parameters.
func NotSupportedByTestnetEnv(p2pNetwork string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if p2pNetwork == "testnet" {
			return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
				"error": "Not implemented by the testnet environment",
			})
		}
		return c.Next()
	}
}
