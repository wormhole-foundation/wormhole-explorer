// package middleare contains all the middleware function to use in the API.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/metrics"
)

// ExtractPagination parses pagination-related query parameters.
func OriginMetrics(m metrics.Metrics) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		path := c.Route().Path
		if !IsK8sPath(path) {
			originHeader := strings.ToLower(c.Get(fiber.HeaderOrigin))
			if originHeader == "" {
				originHeader = "unknown"
			}
			m.IncOrigin(originHeader)
		}
		return err
	}
}
