package pagination

// Pagination definition.
type Pagination struct {
	Offset    int64
	PageSize  int64
	SortOrder string
	SortBy    string
}

// FirstPage return a *Pagination with default values offset and page size.
func FirstPage() *Pagination {
	return &Pagination{Offset: 0, PageSize: 50}
}

// BuildPagination create a new *Pagination.
func BuildPagination(page, pageSize int64, sortOrder, sortBy string) *Pagination {
	p := Pagination{}
	p.SetPage(page).SetPageSize(pageSize).SetSortOrder(sortOrder).SetSortBy(sortBy)
	return &p
}

// SetPageSize set the PageSize field of the Pagination struct.
func (p *Pagination) SetPageSize(limit int64) *Pagination {
	p.PageSize = limit
	return p
}

// SetOffset set the Offset field of the Pagination struct.
func (p *Pagination) SetOffset(offset int64) *Pagination {
	p.Offset = offset
	return p
}

// SetPage set the Page field of the Pagination struct.
func (p *Pagination) SetPage(page int64) *Pagination {
	p.Offset = page * p.PageSize
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
