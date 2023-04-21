package transactions

import (
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type Scorecards struct {
	// Number of VAAs emitted since the creation of the network (does not include Pyth messages)
	TotalTxCount string

	// Number of VAAs emitted in the last 24 hours (does not include Pyth messages).
	TxCount24h string
}

type GlobalTransactionDoc struct {
	ID            string         `bson:"_id" json:"id"`
	OriginTx      *OriginTx      `bson:"originTx" json:"originTx"`
	DestinationTx *DestinationTx `bson:"destinationTx" json:"destinationTx"`
}

// OriginTx representa a origin transaction.
type OriginTx struct {
	ChainID   vaa.ChainID `bson:"chainId" json:"chainId"`
	TxHash    string      `bson:"nativeTxHash" json:"txHash"`
	Status    string      `bson:"status" json:"status"`
	Timestamp *time.Time  `bson:"timestamp" json:"timestamp"`
	From      *string     `bson:"signer" json:"from"`
}

// DestinationTx representa a destination transaction.
type DestinationTx struct {
	ChainID     vaa.ChainID `bson:"chainId" json:"chainId"`
	Status      string      `bson:"status" json:"status"`
	Method      string      `bson:"method" json:"method"`
	TxHash      string      `bson:"txHash" json:"txHash"`
	From        string      `bson:"from" json:"from"`
	To          string      `bson:"to" json:"to"`
	BlockNumber string      `bson:"blockNumber" json:"blockNumber"`
	Timestamp   *time.Time  `bson:"timestamp" json:"timestamp"`
	UpdatedAt   *time.Time  `bson:"updatedAt" json:"updatedAt"`
}

// TransactionUpdate represents a transaction document.
type TransactionUpdate struct {
}

// GlobalTransactionQuery respresent a query for the globalTransactions mongodb document.
type GlobalTransactionQuery struct {
	pagination.Pagination
	id string
}

// Query create a new VaaQuery with default pagination vaues.
func Query() *GlobalTransactionQuery {
	p := pagination.Default()
	return &GlobalTransactionQuery{Pagination: *p}
}

// SetId set the chainId field of the VaaQuery struct.
func (q *GlobalTransactionQuery) SetId(id string) *GlobalTransactionQuery {
	q.id = id
	return q
}

type TransactionCountQuery struct {
	TimeSpan      string
	SampleRate    string
	CumulativeSum bool
}

type TransactionCountResult struct {
	Time  time.Time `json:"time" mapstructure:"_time"`
	Count uint64    `json:"count" mapstructure:"count"`
}

type ChainActivityResult struct {
	ChainSourceID      string `mapstructure:"chain_source_id"`
	ChainDestinationID string `mapstructure:"chain_destination_id"`
	Volume             uint64 `mapstructure:"volume"`
}

type ChainActivityQuery struct {
	Start      *time.Time
	End        *time.Time
	AppIDs     []string
	IsNotional bool
}

func (q *ChainActivityQuery) HasAppIDS() bool {
	return len(q.AppIDs) > 0
}

func (q *ChainActivityQuery) GetAppIDs() []string {
	return q.AppIDs
}

func (q *ChainActivityQuery) GetStart() time.Time {
	if q.Start == nil {
		return time.UnixMilli(0)
	}
	return *q.Start
}

func (q *ChainActivityQuery) GetEnd() time.Time {
	if q.End == nil {
		return time.Now()
	}
	return *q.End
}
