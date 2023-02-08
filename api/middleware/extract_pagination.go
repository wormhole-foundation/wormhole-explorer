// package middleare contains all the middleware function to use in the API.
package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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

	// get page number
	pageNumberStr := ctx.Query("page", "0")
	pageNumber, err := strconv.ParseInt(pageNumberStr, 10, 64)
	if err != nil {
		return nil, err
	}
	if pageNumber < 0 {
		return nil, errors.New(`parameter "page" must be a non-negative integer`)
	}

	// get page size
	pageSizeStr := ctx.Query("pageSize", "50")
	pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
	if err != nil {
		return nil, err
	}
	if pageSize <= 0 {
		return nil, errors.New(`parameter "pageSize" must be a positive integer`)
	}
	skip := pageSize * pageNumber

	// get sort order
	sortOrder := strings.ToUpper(ctx.Query("sortOrder", "DESC"))
	if sortOrder != "ASC" && sortOrder != "DESC" {
		return nil, errors.New(`parameter "sortOrder" must either be "ASC" or "DESC"`)
	}

	// `sortBy` is currently not exposed as a parameter, but could be in the future.
	sortBy := ctx.Query("sortBy", "indexedAt")

	p := &pagination.Pagination{
		Skip:      skip,
		Limit:     pageSize,
		SortOrder: sortOrder,
		SortBy:    sortBy,
	}
	ctx.Locals("pagination", p)
	return p, nil
}

// GetPaginationFromContext get pagination from context.
func GetPaginationFromContext(ctx *fiber.Ctx) *pagination.Pagination {
	p := ctx.Locals("pagination")
	if p == nil {
		return nil
	}
	return p.(*pagination.Pagination)
}
