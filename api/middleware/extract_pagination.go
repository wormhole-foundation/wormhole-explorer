// package middleare contains all the middleware function to use in the API.
package middleware

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
)

// ExtractPagination middleware invoke pagination.ExtractPagination.
func ExtractPagination(c *fiber.Ctx) error {
	if c.Method() != http.MethodGet {
		return c.Next()
	}
	extractPagination(c)
	return c.Next()
}

// extractPagination get pagination query params and build a *Pagination.
func extractPagination(ctx *fiber.Ctx) (*pagination.Pagination, error) {

	pageNumberStr := ctx.Query("page", "0")
	pageNumber, err := strconv.ParseInt(pageNumberStr, 10, 64)
	if err != nil {
		return nil, err
	}

	pageSizeStr := ctx.Query("pageSize", "50")
	pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
	if err != nil {
		return nil, err
	}
	skip := pageSize * pageNumber

	sortOrder := ctx.Query("sortOrder", "DESC")
	sortBy := ctx.Query("sortBy", "indexedAt")

	p := pagination.Pagination{
		Skip:      skip,
		Limit:     pageSize,
		SortOrder: sortOrder,
		SortBy:    sortBy,
	}
	ctx.Locals("pagination", p)
	return &p, nil
}

// GetPaginationFromContext get pagination from context.
func GetPaginationFromContext(ctx *fiber.Ctx) *pagination.Pagination {
	p := ctx.Locals("pagination")
	if p == nil {
		return nil
	}
	return p.(*pagination.Pagination)
}
