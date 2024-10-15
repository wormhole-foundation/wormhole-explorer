// Package governor handle the request of governor data from governor endpoint defined in the api.
package governor

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/api/internal/mongo"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// GovConfigPage represent a governor configuration.
type GovConfig struct {
	ID        string              `bson:"_id" json:"id"`
	CreatedAt *time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt *time.Time          `bson:"updatedAt" json:"updatedAt"`
	NodeName  string              `bson:"nodename" json:"nodeName"`
	Counter   int                 `bson:"counter" json:"counter"`
	Chains    []*GovConfigChains  `bson:"chains" json:"chains"`
	Tokens    []*GovConfigfTokens `bson:"tokens" json:"tokens"`
}

type GovConfigChains struct {
	ChainID            vaa.ChainID  `bson:"chainid" json:"chainId"`
	NotionalLimit      mongo.Uint64 `bson:"notionallimit" json:"notionalLimit"`
	BigTransactionSize mongo.Uint64 `bson:"bigtransactionsize" json:"bigTransactionSize"`
}

type GovConfigfTokens struct {
	OriginChainID int     `bson:"originchainid" json:"originChainId"`
	OriginAddress string  `bson:"originaddress" json:"originAddress"`
	Price         float32 `bson:"price" json:"price"`
}

// GovStatusPage represent a governor status.
type GovStatus struct {
	ID        string             `bson:"_id" json:"id"`
	CreatedAt *time.Time         `bson:"createdAt" json:"createdAt"`
	UpdatedAt *time.Time         `bson:"updatedAt" json:"updatedAt"`
	NodeName  string             `bson:"nodename" json:"nodeName"`
	Chains    []*GovStatusChains `bson:"chains" json:"chains"`
}

type GovStatusChains struct {
	ChainID                    vaa.ChainID              `bson:"chainid" json:"chainId"`
	RemainingAvailableNotional mongo.Uint64             `bson:"remainingavailablenotional" json:"remainingAvailableNotional"`
	Emitters                   []*GovStatusChainEmitter `bson:"emitters" json:"emitters"`
}

type GovStatusChainEmitter struct {
	EmitterAddress    string       `bson:"emitteraddress" json:"emitterAddress"`
	TotalEnqueuedVaas mongo.Uint64 `bson:"totalenqueuedvaas" json:"totalEnqueuedVaas"`
	EnqueuedVass      interface{}  `bson:"enqueuedvaas" json:"enqueuedVaas"`
}

// NotionalLimit represent the notional limit value and maximun tranasction size for a chainID.
type NotionalLimit struct {
	ChainID            vaa.ChainID `json:"chainId"`
	NotionalLimit      uint64      `json:"notionalLimit"`
	MaxTransactionSize uint64      `json:"maxTransactionSize"`
}

type notionalLimitMongo struct {
	ChainID            vaa.ChainID   `bson:"chainid"`
	NotionalLimit      *mongo.Uint64 `bson:"notionallimit"`
	MaxTransactionSize *mongo.Uint64 `bson:"maxtransactionsize"`
}

type notionalLimitSQL struct {
	ChainID            vaa.ChainID `db:"chainid"`
	NotionalLimit      string      `db:"notionallimit"`
	MaxTransactionSize string      `db:"maxtransactionsize"`
}

// NotionalLimitDetail represent a notional limit value
type NotionalLimitDetail struct {
	ID                 string      `json:"id"`
	ChainID            vaa.ChainID `json:"chainId"`
	NodeName           string      `json:"nodeName"`
	NotionalLimit      uint64      `json:"notionalLimit"`
	MaxTransactionSize uint64      `json:"maxTransactionSize"`
	CreatedAt          *time.Time  `json:"createdAt"`
	UpdatedAt          *time.Time  `json:"updatedAt"`
}

type notionalLimitDetailSQL struct {
	ID                 string      `db:"id"`
	ChainID            vaa.ChainID `db:"chainid"`
	NodeName           string      `db:"guardian_name"`
	NotionalLimit      string      `db:"notionallimit"`
	MaxTransactionSize string      `db:"maxtransactionsize"`
	CreatedAt          *time.Time  `db:"created_at"`
	UpdatedAt          *time.Time  `db:"updated_at"`
}

type notionalLimitDetailMongo struct {
	ID                string        `bson:"_id"`
	ChainID           vaa.ChainID   `bson:"chainId"`
	NodeName          string        `bson:"nodename"`
	NotionalLimit     *mongo.Uint64 `bson:"notionalLimit"`
	MaxTrasactionSize *mongo.Uint64 `bson:"maxTransactionSize"`
	CreatedAt         *time.Time    `bson:"createdAt"`
	UpdatedAt         *time.Time    `bson:"updatedAt"`
}

// NotionalAvailable represent the available notional for chainID.
type NotionalAvailable struct {
	ChainID           vaa.ChainID `json:"chainId"`
	AvailableNotional uint64      `json:"availableNotional"`
}

type notionalAvailableSQL struct {
	ChainID           vaa.ChainID `db:"chainid"`
	AvailableNotional string      `db:"availablenotional"`
}

type notionalAvailableMongo struct {
	ChainID           vaa.ChainID   `bson:"chainid"`
	AvailableNotional *mongo.Uint64 `bson:"availableNotional"`
}

// NotionalAvailableDetail represent a notional available value.
type NotionalAvailableDetail struct {
	ID                string      `json:"id"`
	ChainID           vaa.ChainID `json:"chainId"`
	NodeName          string      `json:"nodeName"`
	NotionalAvailable uint64      `json:"availableNotional"`
	CreatedAt         *time.Time  `json:"createdAt"`
	UpdatedAt         *time.Time  `json:"updatedAt"`
}

type notionalAvailableDetailSQL struct {
	ID                string      `db:"id"`
	ChainID           vaa.ChainID `db:"chainid"`
	NodeName          string      `db:"guardian_name"`
	NotionalAvailable string      `db:"availablenotional"`
	CreatedAt         *time.Time  `db:"created_at"`
	UpdatedAt         *time.Time  `db:"updated_at"`
}

type notionalAvailableDetailMongo struct {
	ID                string        `bson:"_id"`
	ChainID           vaa.ChainID   `bson:"chainId"`
	NodeName          string        `bson:"nodeName"`
	NotionalAvailable *mongo.Uint64 `bson:"availableNotional"`
	CreatedAt         *time.Time    `bson:"createdAt"`
	UpdatedAt         *time.Time    `bson:"updatedAt"`
}

type Emitter struct {
	Address           string        `bson:"emitteraddress" json:"emitterAddress"`
	TotalEnqueuedVaas uint64        `bson:"totalenqueuedvaas" json:"totalEnqueuedVaas"`
	EnqueuedVaas      []EnqueuedVAA `bson:"enqueuedvaas" json:"enqueuedVaas"`
}

type emitterMongo struct {
	Address           string             `bson:"emitteraddress" json:"emitterAddress"`
	TotalEnqueuedVaas *mongo.Uint64      `bson:"totalenqueuedvaas" json:"totalEnqueuedVaas"`
	EnqueuedVaas      []enqueuedVAAMongo `bson:"enqueuedvaas" json:"enqueuedVaas"`
}

type enqueuedVAAMongo struct {
	Sequence    string        `bson:"sequence" json:"sequence"`
	ReleaseTime *time.Time    `bson:"releasetime" json:"releaseTime"`
	Notional    *mongo.Uint64 `bson:"notionalvalue" json:"notionalValue"`
	TxHash      string        `bson:"txhash" json:"txHash"`
}

type emitterSQL struct {
	Address           string           `json:"emitteraddress"`
	TotalEnqueuedVaas float64          `json:"totalenqueuedvaas"`
	EnqueuedVaas      []enqueuedVAASQL `json:"enqueuedvaas"`
}

type enqueuedVAASQL struct {
	Sequence    string     `json:"sequence"`
	ReleaseTime *time.Time `json:"releasetime"`
	Notional    uint64     `json:"notionalvalue"`
	TxHash      string     `json:"txhash"`
}

// EnqueuedVAA definition.
type EnqueuedVAA struct {
	Sequence    string     `bson:"sequence" json:"sequence"`
	ReleaseTime *time.Time `bson:"releasetime" json:"releaseTime"`
	Notional    uint64     `bson:"notionalvalue" json:"notionalValue"`
	TxHash      string     `bson:"txhash" json:"txHash"`
}

// MaxNotionalAvailableRecord definition.
type MaxNotionalAvailableRecord struct {
	ID                string      `json:"id"`
	ChainID           vaa.ChainID `json:"chainId"`
	NodeName          string      `json:"nodeName"`
	NotionalAvailable uint64      `json:"availableNotional"`
	Emitters          []*Emitter  `json:"emitters"`
	CreatedAt         *time.Time  `json:"createdAt"`
	UpdatedAt         *time.Time  `json:"updatedAt"`
}

type maxNotionalAvailableRecordSQL struct {
	ID                string      `db:"id"`
	ChainID           vaa.ChainID `db:"chainid"`
	NodeName          string      `db:"guardian_name"`
	NotionalAvailable string      `db:"availablenotional"`
	Emitters          string      `db:"emitters"`
	CreatedAt         *time.Time  `db:"created_at"`
	UpdatedAt         *time.Time  `db:"updated_at"`
}

type maxNotionalAvailableRecordMongo struct {
	ID                string          `bson:"_id"`
	ChainID           vaa.ChainID     `bson:"chainId"`
	NodeName          string          `bson:"nodeName"`
	NotionalAvailable *mongo.Uint64   `bson:"availableNotional"`
	Emitters          []*emitterMongo `bson:"emitters"`
	CreatedAt         *time.Time      `bson:"createdAt"`
	UpdatedAt         *time.Time      `bson:"updatedAt"`
}

// EnqueuedVaa definition.
type EnqueuedVaa struct {
	ChainID        vaa.ChainID `bson:"chainId" json:"chainId"`
	EmitterAddress string      `bson:"emitterAddress" json:"emitterAddress"`
	Sequence       string      `bson:"sequence" json:"sequence"`
	NotionalValue  int64       `bson:"notionalValue" json:"notionalValue"`
	TxHash         string      `bson:"txHash" json:"txHash"`
}

// MarshalJSON interface implementation.
func (v *EnqueuedVaa) MarshalJSON() ([]byte, error) {
	sequence, err := strconv.ParseUint(v.Sequence, 10, 64)
	if err != nil {
		return []byte{}, err
	}

	type Alias EnqueuedVaa
	return json.Marshal(&struct {
		Sequence uint64 `json:"sequence"`
		*Alias
	}{
		Sequence: sequence,
		Alias:    (*Alias)(v),
	})
}

// EnqueuedVaas definition.
type EnqueuedVaas struct {
	ChainID     vaa.ChainID    `bson:"chainid" json:"chainId"`
	EnqueuedVaa []*EnqueuedVaa `bson:"enqueuedVaas" json:"enqueuedVaas"`
}

// EnqueuedVaaDetail definition.
type EnqueuedVaaDetail struct {
	ChainID        vaa.ChainID `bson:"chainid" json:"chainId" db:"chain_id"`
	EmitterAddress string      `bson:"emitterAddress" json:"emitterAddress" db:"emitter_address"`
	Sequence       string      `bson:"sequence" json:"sequence" db:"sequence"`
	NotionalValue  int64       `bson:"notionalValue" json:"notionalValue" db:"notional_value"`
	TxHash         string      `bson:"txHash" json:"txHash" db:"tx_hash"`
	ReleaseTime    int64       `bson:"releaseTime" json:"releaseTime" db:"release_time"`
}

// MarshalJSON interface implementation.
func (v *EnqueuedVaaDetail) MarshalJSON() ([]byte, error) {
	sequence, err := strconv.ParseUint(v.Sequence, 10, 64)
	if err != nil {
		return []byte{}, err
	}

	type Alias EnqueuedVaaDetail
	return json.Marshal(&struct {
		Sequence uint64 `json:"sequence"`
		*Alias
	}{
		Sequence: sequence,
		Alias:    (*Alias)(v),
	})
}

// GovernorLimit definition.
type GovernorLimit struct {
	ChainID            vaa.ChainID  `bson:"chainId" json:"chainId"`
	AvailableNotional  mongo.Uint64 `bson:"availableNotional" json:"availableNotional"`
	NotionalLimit      mongo.Uint64 `bson:"notionalLimit" json:"notionalLimit"`
	MaxTransactionSize mongo.Uint64 `bson:"maxTransactionSize" json:"maxTransactionSize"`
}

// AvailableNotionalByChain definition.
// This is the structure that is used in guardian api grpc api version.
type AvailableNotionalByChain struct {
	ChainID            vaa.ChainID `db:"chainId" json:"chainId"`
	AvailableNotional  uint64      `db:"availableNotional" json:"remainingAvailableNotional"`
	NotionalLimit      uint64      `db:"notionalLimit" json:"notionalLimit"`
	MaxTransactionSize uint64      `db:"maxTransactionSize" json:"bigTransactionSize"`
}

type availableNotionalByChainMongo struct {
	ChainID            vaa.ChainID  `bson:"chainId" json:"chainId"`
	AvailableNotional  mongo.Uint64 `bson:"availableNotional" json:"remainingAvailableNotional"`
	NotionalLimit      mongo.Uint64 `bson:"notionalLimit" json:"notionalLimit"`
	MaxTransactionSize mongo.Uint64 `bson:"maxTransactionSize" json:"bigTransactionSize"`
}

// TokenList definition
type TokenList struct {
	OriginChainID vaa.ChainID `bson:"originchainid" json:"originChainId"`
	OriginAddress string      `bson:"originaddress" json:"originAddress"`
	Price         float32     `bson:"price" json:"price"`
}

// EnqueuedVaaItem definition
type EnqueuedVaaItem struct {
	EmitterChain   vaa.ChainID  `bson:"chainid" json:"emitterChain"`
	EmitterAddress string       `bson:"emitteraddress" json:"emitterAddress"`
	Sequence       string       `bson:"sequence" json:"sequence"`
	ReleaseTime    int64        `bson:"releasetime" json:"releaseTime"`
	NotionalValue  mongo.Uint64 `bson:"notionalvalue" json:"notionalValue"`
	TxHash         string       `bson:"txhash" json:"txHash"`
}
