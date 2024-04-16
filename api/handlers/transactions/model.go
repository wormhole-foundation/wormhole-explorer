package transactions

import (
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type Scorecards struct {

	// Number of VAAs emitted in the last 24 hours (includes Pyth messages).
	Messages24h string

	// Number of VAAs emitted since the creation of the network (includes Pyth messages).
	TotalMessages string

	// Number of VAAs emitted since the creation of the network (does not include Pyth messages)
	TotalTxCount string

	//Volume transferred since the creation of the network, in USD.
	TotalTxVolume string

	// Total value locked in USD.
	Tvl string

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

// OriginTx represents a origin transaction.
type OriginTx struct {
	TxHash    string        `bson:"nativeTxHash" json:"txHash"`
	From      string        `bson:"from" json:"from"`
	Status    string        `bson:"status" json:"status"`
	Attribute *AttributeDoc `bson:"attribute" json:"attribute"`
}

// AttributeDoc represents a custom attribute for a origin transaction.
type AttributeDoc struct {
	Type  string         `bson:"type" json:"type"`
	Value map[string]any `bson:"value" json:"value"`
}

// DestinationTx represents a destination transaction.
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
	ChainSourceID      string `mapstructure:"emitter_chain" json:"emitter_chain"`
	ChainDestinationID string `mapstructure:"destination_chain" json:"destination_chain"`
	Volume             uint64 `mapstructure:"_value" json:"volume"`
}

type ChainActivityTopResult struct {
	Time               time.Time `json:"from" mapstructure:"_time"`
	To                 string    `json:"to" mapstructure:"to"`
	ChainSourceID      string    `mapstructure:"emitter_chain" json:"emitter_chain"`
	ChainDestinationID string    `mapstructure:"destination_chain" json:"destination_chain,omitempty"`
	Volume             uint64    `mapstructure:"volume" json:"volume"`
	Txs                uint64    `mapstructure:"count" json:"count"`
}

type ChainActivityTimeSpan string

const (
	ChainActivityTs7Days   ChainActivityTimeSpan = "7d"
	ChainActivityTs30Days  ChainActivityTimeSpan = "30d"
	ChainActivityTs90Days  ChainActivityTimeSpan = "90d"
	ChainActivityTs1Year   ChainActivityTimeSpan = "1y"
	ChainActivityTsAllTime ChainActivityTimeSpan = "all-time"
)

// ParseChainActivityTimeSpan parses a string and returns a `ChainActivityTimeSpan`.
func ParseChainActivityTimeSpan(s string) (ChainActivityTimeSpan, error) {

	if s == string(ChainActivityTs7Days) ||
		s == string(ChainActivityTs30Days) ||
		s == string(ChainActivityTs90Days) ||
		s == string(ChainActivityTs1Year) ||
		s == string(ChainActivityTsAllTime) {

		tmp := ChainActivityTimeSpan(s)
		return tmp, nil
	}
	return "", fmt.Errorf("invalid time span: %s", s)
}

type ChainActivityQuery struct {
	TimeSpan   ChainActivityTimeSpan
	IsNotional bool
	AppIDs     []string
}

func (q *ChainActivityQuery) HasAppIDS() bool {
	return len(q.AppIDs) > 0
}

func (q *ChainActivityQuery) GetAppIDs() []string {
	return q.AppIDs
}

// Token represents a token.
type Token struct {
	Symbol      domain.Symbol `json:"symbol"`
	CoingeckoID string        `json:"coingeckoId"`
	Decimals    int64         `json:"decimals"`
}

type TransactionDto struct {
	ID                     string                 `bson:"_id"`
	EmitterChain           sdk.ChainID            `bson:"emitterChain"`
	EmitterAddr            string                 `bson:"emitterAddr"`
	TxHash                 string                 `bson:"txHash"`
	Timestamp              time.Time              `bson:"timestamp"`
	Symbol                 string                 `bson:"symbol"`
	UsdAmount              string                 `bson:"usdAmount"`
	TokenAmount            string                 `bson:"tokenAmount"`
	GlobalTransations      []GlobalTransactionDoc `bson:"globalTransactions"`
	Payload                map[string]interface{} `bson:"payload"`
	StandardizedProperties map[string]interface{} `bson:"standardizedProperties"`
}

type ChainActivityTopsQuery struct {
	SourceChain  sdk.ChainID  `json:"source_chain"`
	TargetChain  sdk.ChainID  `json:"target_chain"`
	From         time.Time    `json:"from"`
	To           time.Time    `json:"to"`
	TimeInterval TimeInterval `json:"time_interval"`
}

type TimeInterval string

const (
	Hour  TimeInterval = "1h"
	Day   TimeInterval = "1d"
	Week  TimeInterval = "1w"
	Month TimeInterval = "1mo"
	Year  TimeInterval = "1y"
)

func (t TimeInterval) IsValid() bool {
	return t == Hour || t == Day || t == Week || t == Month || t == Year
}
