// package middleare contains all the middleware function to use in the API.
package middleware

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
)

// ExtractPagination parses pagination-related query parameters.
func ExtractPagination(ctx *fiber.Ctx) (*pagination.Pagination, error) {

	// get page number
	pageNumberStr := ctx.Query("page", "0")
	pageNumber, err := strconv.ParseInt(pageNumberStr, 10, 64)
	if err != nil || pageNumber < 0 {
		msg := `parameter 'page' must be a non-negative integer`
		return nil, response.NewInvalidParamError(ctx, msg, err)
	}

	// get page size
	pageSizeStr := ctx.Query("pageSize", "50")
	pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
	if err != nil || pageSize <= 0 {
		msg := `parameter 'pageSize' must be a positive integer`
		return nil, response.NewInvalidParamError(ctx, msg, err)
	}
	skip := pageSize * pageNumber

	// get sort order
	sortOrder := strings.ToUpper(ctx.Query("sortOrder", "DESC"))
	if sortOrder != "ASC" && sortOrder != "DESC" {
		msg := `parameter 'sortOrder' must either be 'ASC' or 'DESC'`
		return nil, response.NewInvalidParamError(ctx, msg, nil)
	}

	// `sortBy` is currently not exposed as a parameter, but could be in the future.
	sortBy := ctx.Query("sortBy", "indexedAt")

	// initialize the result struct and return
	p := &pagination.Pagination{
		Skip:      skip,
		Limit:     pageSize,
		SortOrder: sortOrder,
		SortBy:    sortBy,
	}
	return p, nil
}
