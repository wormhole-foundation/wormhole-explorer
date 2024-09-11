package storage

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type PricesRepository interface {
	Upsert(ctx context.Context, p OperationPrice) error
}

type OperationPrice struct {
	Digest        string
	VaaID         string
	ChainID       sdk.ChainID
	TokenChainID  uint16
	TokenAddress  string
	Symbol        string
	CoinGeckoID   string
	TokenUSDPrice decimal.Decimal
	TotalToken    decimal.Decimal
	TotalUSD      decimal.Decimal
	Timestamp     time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type VaaPageQuery struct {
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

type VaaRepository interface {
	FindByVaaID(ctx context.Context, id string) (*Vaa, error)
	FindPage(ctx context.Context, query VaaPageQuery, pagination Pagination) ([]*Vaa, error)
}

type Vaa struct {
	ID    string
	VaaID string
	Vaa   []byte
}
