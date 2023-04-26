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

	// get page number from query params
	var pageNumber *int64
	if param := ctx.Query("page"); param != "" {
		n, err := strconv.ParseInt(param, 10, 64)
		if err != nil || n < 0 {
			msg := `parameter 'page' must be a non-negative integer`
			return nil, response.NewInvalidParamError(ctx, msg, err)
		}
		pageNumber = &n
	}

	// get page size from query params
	var pageSize *int64
	if param := ctx.Query("pageSize"); param != "" {
		n, err := strconv.ParseInt(param, 10, 64)
		if err != nil || n <= 0 {
			msg := `parameter 'pageSize' must be a positive integer`
			return nil, response.NewInvalidParamError(ctx, msg, err)
		}
		pageSize = &n
	}

	// get sort order from query params
	var sortOrder string
	if param := strings.ToUpper(ctx.Query("sortOrder", "DESC")); param != "" {
		if param != "ASC" && param != "DESC" {
			msg := `parameter 'sortOrder' must either be 'ASC' or 'DESC'`
			return nil, response.NewInvalidParamError(ctx, msg, nil)
		}
		sortOrder = param
	}

	// build the result and return
	p := pagination.Default()
	if sortOrder != "" {
		p.SetSortOrder(sortOrder)
	}
	if pageSize != nil {
		p.SetLimit(*pageSize)
	}
	if pageNumber != nil {
		p.SetSkip(p.Limit * *pageNumber)
	}
	return p, nil
}
