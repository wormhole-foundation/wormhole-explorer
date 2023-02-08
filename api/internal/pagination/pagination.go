package pagination

// Pagination definition.
type Pagination struct {
	Skip      int64
	Limit     int64
	SortOrder string
	SortBy    string
}

// Default returns a `*Pagination` with default values.
func Default() *Pagination {
	return &Pagination{Skip: 0, Limit: 50}
}

// New creates a `*Pagination`.
func New(skip, limit int64, sortOrder, sortBy string) *Pagination {

	var p Pagination

	p.
		SetPageSize(limit).
		SetSkip(skip).
		SetSortOrder(sortOrder).
		SetSortBy(sortBy)

	return &p
}

// SetPageSize set the PageSize field of the Pagination struct.
func (p *Pagination) SetPageSize(limit int64) *Pagination {
	p.Limit = limit
	return p
}

// SetSkip sets the `Skip` field of the `Pagination` struct.
func (p *Pagination) SetSkip(skip int64) *Pagination {
	p.Skip = skip
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
