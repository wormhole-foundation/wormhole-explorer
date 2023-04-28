package pagination

// Pagination definition.
type Pagination struct {
	Skip      int64
	Limit     int64
	SortOrder string
}

// Default returns a `*Pagination` with default values.
func Default() *Pagination {

	p := &Pagination{
		Skip:      0,
		Limit:     50,
		SortOrder: "DESC",
	}

	return p
}

func (p *Pagination) SetSkip(skip int64) *Pagination {
	p.Skip = skip
	return p
}

func (p *Pagination) SetLimit(limit int64) *Pagination {
	p.Limit = limit
	return p
}

func (p *Pagination) SetSortOrder(sortOrder string) *Pagination {
	p.SortOrder = sortOrder
	return p
}

// GetSortInt mapping to mongodb sort values.
func (p *Pagination) GetSortInt() int {
	if p.SortOrder == "ASC" {
		return 1
	}
	return -1
}
