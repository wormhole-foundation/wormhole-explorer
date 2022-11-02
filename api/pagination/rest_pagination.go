package pagination

import (
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func ExtractPagination(ctx *fiber.Ctx) (*Pagination, error) {
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

	sortOrder := ctx.Query("sortOrder", "DESC")
	sortBy := ctx.Query("sortBy", "indexedAt")

	p := BuildPagination(pageNumber, pageSize, sortOrder, sortBy)
	ctx.Locals("pagination", p)
	return p, nil
}

func GetFromContext(ctx *fiber.Ctx) *Pagination {
	p := ctx.Locals("pagination")
	if p == nil {
		return nil
	}
	return p.(*Pagination)
}
