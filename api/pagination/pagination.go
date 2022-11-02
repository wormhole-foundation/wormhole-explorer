package pagination

type Pagination struct {
	Offset    int64
	PageSize  int64
	SortOrder string
	SortBy    string
}

func FirstPage() *Pagination {
	return &Pagination{Offset: 0, PageSize: 50}
}

func BuildPagination(page, pageSize int64, sortOrder, sortBy string) *Pagination {
	p := Pagination{}
	p.SetPage(page).SetPageSize(pageSize).SetSortOrder(sortOrder).SetSortBy(sortBy)
	return &p
}

func (p *Pagination) SetPageSize(limit int64) *Pagination {
	p.PageSize = limit
	return p
}

func (p *Pagination) SetOffset(offset int64) *Pagination {
	p.Offset = offset
	return p
}

func (p *Pagination) SetPage(page int64) *Pagination {
	p.Offset = page * p.PageSize
	return p
}

func (p *Pagination) SetSortOrder(order string) *Pagination {
	p.SortOrder = order
	return p
}

func (p *Pagination) SetSortBy(by string) *Pagination {
	p.SortBy = by
	return p
}

func (p *Pagination) GetSortInt() int {
	if p.SortOrder == "ASC" {
		return 1
	}
	return -1
}
