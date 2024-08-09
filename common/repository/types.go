package repository

import (
	"context"
	"time"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// IndexingTimestamps struct.
type IndexingTimestamps struct {
	IndexedAt time.Time `bson:"indexedAt"`
}

// IndexedAt returns a new IndexingTimestamps.
func IndexedAt(t time.Time) IndexingTimestamps {
	return IndexingTimestamps{
		IndexedAt: t,
	}
}

// VaaQuery is a query for VAA.
type VaaQuery struct {
	StartTime      *time.Time
	EndTime        *time.Time
	EmitterChainID *sdk.ChainID
	EmitterAddress *string
	Sequence       *string
}

// Pagination is a pagination for VAA.
type Pagination struct {
	Page     int64
	PageSize int64
	SortAsc  bool
}

// AttestationVaa represent a vaa attestation row in the postgres database.
type AttestationVaa struct {
	ID             string      `db:"id"`
	VaaID          string      `db:"vaa_id"`
	Version        uint8       `db:"version"`
	EmitterChain   sdk.ChainID `db:"emitter_chain_id"`
	EmitterAddress string      `db:"emitter_address"`
	Sequence       uint64      `db:"sequence"`
	GuardianSetIdx uint32      `db:"guardian_set_index"`
	Raw            []byte      `db:"raw"`
	Timestamp      time.Time   `db:"timestamp"`
	Active         bool        `db:"active"`
	IsDuplicated   bool        `db:"is_duplicated"`
	CreatedAt      time.Time   `db:"created_at"`
	UpdatedAt      *time.Time  `db:"updated_at"`
}

type GuardianSetStorager interface {
	FindAll(ctx context.Context) ([]*GuardianSet, error)
	Upsert(ctx context.Context, doc *GuardianSet) error
}
