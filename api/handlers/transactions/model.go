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

	// Volume transferred through the token bridge in the last 24 hours, in USD.
	Volume24h string

	// Volume transferred in the last 7 days, in USD.
	Volume7d string

	// Volume transferred in the last 30 days, in USD.
	Volume30d string
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

type ApplicationActivityTotalsResult struct {
	From   time.Time `json:"from" mapstructure:"_time"`
	To     time.Time `json:"to" mapstructure:"to"`
	Volume float64   `mapstructure:"total_value_transferred" json:"total_value_transferred"`
	Txs    uint64    `mapstructure:"total_messages" json:"total_messages"`
	AppID  string    `mapstructure:"app_id" json:"app_id"`
}

type ApplicationActivityResult struct {
	From   time.Time `mapstructure:"_time" json:"from"`
	To     time.Time `mapstructure:"to" json:"to"`
	Volume float64   `mapstructure:"total_value_transferred" json:"total_value_transferred"`
	Txs    uint64    `mapstructure:"total_messages" json:"total_messages"`
	AppID1 string    `mapstructure:"app_id_1" json:"app_id_1"`
	AppID2 string    `mapstructure:"app_id_2" json:"app_id_2"`
	AppID3 string    `mapstructure:"app_id_3" json:"app_id_3"`
}

type ChainActivityTopResults []ChainActivityTopResult

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
	SourceChains []sdk.ChainID `json:"source_chain"`
	TargetChains []sdk.ChainID `json:"target_chain"`
	AppId        string        `json:"app_id"`
	From         time.Time     `json:"from"`
	To           time.Time     `json:"to"`
	Timespan     Timespan      `json:"timespan"`
}

type ApplicationActivityQuery struct {
	AppId          string
	ExclusiveAppID bool
	From           time.Time
	To             time.Time
	Timespan       Timespan
}

type Timespan string

const (
	Hour  Timespan = "1h"
	Day   Timespan = "1d"
	Month Timespan = "1mo"
	Year  Timespan = "1y"
)

func (t Timespan) IsValid() bool {
	return t == Hour || t == Day || t == Month || t == Year
}

type TokenSymbolActivityQuery struct {
	From         time.Time
	To           time.Time
	TokenSymbols []string
	SourceChains []sdk.ChainID
	TargetChains []sdk.ChainID
	Timespan     Timespan
}

type TokenVolume struct {
	Symbol string  `json:"symbol"`
	Volume float64 `json:"volume"`
}

type TokenSymbolActivityResult struct {
	Symbol              string      `mapstructure:"symbol" json:"symbol,omitempty"`
	From                time.Time   `mapstructure:"_time" json:"from"`
	To                  time.Time   `mapstructure:"to" json:"to"`
	Volume              float64     `mapstructure:"volume" json:"total_value_transferred"`
	Txs                 uint64      `mapstructure:"txs" json:"total_messages"`
	EmitterChainStr     string      `mapstructure:"emitter_chain"`
	DestinationChainStr string      `mapstructure:"destination_chain"`
	EmitterChain        sdk.ChainID `json:"emitter_chain"`
	DestinationChain    sdk.ChainID `json:"destination_chain"`
}

// StatsOverview Mayan overview stats
type StatsOverview struct {
	Last24h StatsData `json:"last24h"`
	AllTime StatsData `json:"allTime"`
}

type StatsData struct {
	Volume        uint64 `json:"volume"`
	ToSolCount    uint64 `json:"toSolCount"`
	FromSolCount  uint64 `json:"fromSolCount"`
	Swaps         uint64 `json:"swaps"`
	ActiveTraders uint64 `json:"activeTraders"`
}
