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

	p := &Pagination{
		Skip:      0,
		Limit:     50,
		SortOrder: "DESC",
		SortBy:    "indexedAt",
	}

	return p
}

// GetSortInt mapping to mongodb sort values.
func (p *Pagination) GetSortInt() int {
	if p.SortOrder == "ASC" {
		return 1
	}
	return -1
}
