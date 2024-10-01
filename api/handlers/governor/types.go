package governor

import (
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
)

// GovernorQuery respresent a query for the governors mongodb documents.
type GovernorQuery struct {
	pagination.Pagination
	id *types.Address
}

// NewGovernorQuery creates a new `*GovernorQuery` with default pagination values.
func NewGovernorQuery() *GovernorQuery {
	p := pagination.Default()
	return &GovernorQuery{Pagination: *p}
}

// SetID sets the `id` field of the GovernorQuery struct.
func (q *GovernorQuery) SetID(id *types.Address) *GovernorQuery {

	// Make a deep copy to avoid aliasing bugs
	q.id = id.Copy()

	return q
}

// SetPagination set the pagination field of the GovernorQuery struct.
func (q *GovernorQuery) SetPagination(p *pagination.Pagination) *GovernorQuery {
	q.Pagination = *p
	return q
}
