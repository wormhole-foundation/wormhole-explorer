package transactions

import (
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type Scorecards struct {
	// Number of VAAs emitted since the creation of the network (does not include Pyth messages)
	TotalTxCount string

	//Volume transferred since the creation of the network, in USD.
	TotalTxVolume string

	// Number of VAAs emitted in the last 24 hours (does not include Pyth messages).
	TxCount24h string

	// Volume transferred through the token bridge in the last 24 hours, in USD.
	Volume24h string
}

// AssetDTO is used for the return value of the function `GetTopAssets`.
type AssetDTO struct {
	EmitterChain sdk.ChainID
	TokenChain   sdk.ChainID
	TokenAddress string
	Volume       string
}

// ChainPairDTO is used for the return value of the function `GetTopChainPairs`.
type ChainPairDTO struct {
	EmitterChain      sdk.ChainID
	DestinationChain  sdk.ChainID
	NumberOfTransfers string
}

// TopStatisticsTimeSpan is used as an input parameter for the functions `GetTopAssets` and `GetTopChainPairs`.
type TopStatisticsTimeSpan string

const (
	TimeSpan7Days  TopStatisticsTimeSpan = "7d"
	TimeSpan15Days TopStatisticsTimeSpan = "15d"
	TimeSpan30Days TopStatisticsTimeSpan = "30d"
)

// ParseTopStatisticsTimeSpan parses a string and returns a `TopAssetsTimeSpan`.
func ParseTopStatisticsTimeSpan(s string) (*TopStatisticsTimeSpan, error) {

	if s == string(TimeSpan7Days) ||
		s == string(TimeSpan15Days) ||
		s == string(TimeSpan30Days) {

		tmp := TopStatisticsTimeSpan(s)
		return &tmp, nil
	}

	return nil, fmt.Errorf("invalid time span: %s", s)
}

type GlobalTransactionDoc struct {
	ID            string         `bson:"_id" json:"id"`
	OriginTx      *OriginTx      `bson:"originTx" json:"originTx"`
	DestinationTx *DestinationTx `bson:"destinationTx" json:"destinationTx"`
}

// OriginTx representa a origin transaction.
type OriginTx struct {
	ChainID   sdk.ChainID `bson:"chainId" json:"chainId"`
	TxHash    string      `bson:"nativeTxHash" json:"txHash"`
	Timestamp *time.Time  `bson:"timestamp" json:"timestamp"`
	Status    string      `bson:"status" json:"status"`
}

// DestinationTx representa a destination transaction.
type DestinationTx struct {
	ChainID     sdk.ChainID `bson:"chainId" json:"chainId"`
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
	Count uint64    `json:"count" mapstructure:"_value"`
}

type ChainActivityResult struct {
	ChainSourceID      string `mapstructure:"emitter_chain"`
	ChainDestinationID string `mapstructure:"destination_chain"`
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
