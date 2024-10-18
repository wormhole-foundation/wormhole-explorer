package vaa

import (
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// VaaQuery respresent a query for the vaa mongodb document.
type VaaQuery struct {
	pagination.Pagination
	ids                  []string
	chainId              sdk.ChainID
	emitter              string
	sequence             string
	txHash               string
	appId                string
	toChain              sdk.ChainID
	duplicated           bool
	includeParsedPayload bool
}

// Query create a new VaaQuery with default pagination vaues.
func Query() *VaaQuery {
	p := pagination.Default()
	return &VaaQuery{Pagination: *p}
}

func (q *VaaQuery) SetIDs(ids []string) *VaaQuery {
	q.ids = ids
	return q
}

// SetChain set the chainId field of the VaaQuery struct.
func (q *VaaQuery) SetChain(chainID sdk.ChainID) *VaaQuery {
	q.chainId = chainID
	return q
}

// SetEmitter set the emitter field of the VaaQuery struct.
func (q *VaaQuery) SetEmitter(emitter string) *VaaQuery {
	q.emitter = emitter
	return q
}

// SetSequence set the sequence field of the VaaQuery struct.
func (q *VaaQuery) SetSequence(seq string) *VaaQuery {
	q.sequence = seq
	return q
}

// SetPagination set the pagination field of the VaaQuery struct.
func (q *VaaQuery) SetPagination(p *pagination.Pagination) *VaaQuery {
	q.Pagination = *p
	return q
}

// SetTxHash set the txHash field of the VaaQuery struct.
func (q *VaaQuery) SetTxHash(txHash string) *VaaQuery {
	q.txHash = txHash
	return q
}

func (q *VaaQuery) SetAppId(appId string) *VaaQuery {
	q.appId = appId
	return q
}

func (q *VaaQuery) SetToChain(toChain sdk.ChainID) *VaaQuery {
	q.toChain = toChain
	return q
}

func (q *VaaQuery) IncludeParsedPayload(val bool) *VaaQuery {
	q.includeParsedPayload = val
	return q
}

func (q *VaaQuery) SetDuplicated(val bool) *VaaQuery {
	q.duplicated = val
	return q
}

func (q *VaaQuery) getSortPredicate() bson.E {
	return bson.E{"timestamp", q.GetSortInt()}
}

func (q *VaaQuery) findOptions() *options.FindOptions {

	sort := bson.D{q.getSortPredicate()}

	return options.
		Find().
		SetSort(sort).
		SetLimit(q.Limit).
		SetSkip(q.Skip)
}
