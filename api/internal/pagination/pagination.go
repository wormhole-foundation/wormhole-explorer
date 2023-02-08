package pagination

// Pagination definition.
type Pagination struct {
	Offset    int64
	Limit     int64
	SortOrder string
	SortBy    string
}

// FirstPage return a *Pagination with default values offset and page size.
func FirstPage() *Pagination {
	return &Pagination{Offset: 0, Limit: 50}
}

// BuildPagination create a new *Pagination.
func BuildPagination(offset, limit int64, sortOrder, sortBy string) *Pagination {

	var p Pagination

	p.
		SetPageSize(limit).
		SetOffset(offset).
		SetSortOrder(sortOrder).
		SetSortBy(sortBy)

	return &p
}

// SetPageSize set the PageSize field of the Pagination struct.
func (p *Pagination) SetPageSize(limit int64) *Pagination {
	p.Limit = limit
	return p
}

// SetOffset set the Offset field of the Pagination struct.
func (p *Pagination) SetOffset(offset int64) *Pagination {
	p.Offset = offset
	return p
}

// SetSortOrder set the SortOrder field of the Pagination struct.
func (p *Pagination) SetSortOrder(order string) *Pagination {
	p.SortOrder = order
	return p
}

// SetSortBy set the SortBy field of the Pagination struct.
func (p *Pagination) SetSortBy(by string) *Pagination {
	p.SortBy = by
	return p
}

// GetSortInt mapping to mongodb sort values.
func (p *Pagination) GetSortInt() int {
	if p.SortOrder == "ASC" {
		return 1
	}
	return -1
}
