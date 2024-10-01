package observations

import (
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// ObservationQuery respresent a query for the observation mongodb document.
type ObservationQuery struct {
	pagination.Pagination
	chainId      vaa.ChainID
	emitter      string
	sequence     string
	guardianAddr string
	hash         []byte
	txHash       *types.TxHash
}

// Query create a new ObservationQuery with default pagination vaues.
func Query() *ObservationQuery {
	page := pagination.Default()
	return &ObservationQuery{Pagination: *page}
}

// SetEmitter set the chainId field of the ObservationQuery struct.
func (q *ObservationQuery) SetChain(chainID vaa.ChainID) *ObservationQuery {
	q.chainId = chainID
	return q
}

// SetEmitter set the emitter field of the ObservationQuery struct.
func (q *ObservationQuery) SetEmitter(emitter string) *ObservationQuery {
	q.emitter = emitter
	return q
}

// SetSequence set the sequence field of the ObservationQuery struct.
func (q *ObservationQuery) SetSequence(seq string) *ObservationQuery {
	q.sequence = seq
	return q
}

// SetGuardianAddr set the guardianAddr field of the ObservationQuery struct.
func (q *ObservationQuery) SetGuardianAddr(guardianAddr string) *ObservationQuery {
	q.guardianAddr = guardianAddr
	return q
}

// SetHash set the hash field of the ObservationQuery struct.
func (q *ObservationQuery) SetHash(hash []byte) *ObservationQuery {
	q.hash = hash
	return q
}

// SetHash set the hash field of the ObservationQuery struct.
func (q *ObservationQuery) SetTxHash(txHash *types.TxHash) *ObservationQuery {
	q.txHash = txHash
	return q
}

// SetPagination set the pagination field of the ObservationQuery struct.
func (q *ObservationQuery) SetPagination(p *pagination.Pagination) *ObservationQuery {
	q.Pagination = *p
	return q
}
