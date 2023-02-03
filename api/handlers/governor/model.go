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
	ChainID           vaa.ChainID   `bson:"chainid" json:"chainId"`
	NotionalLimit     *mongo.Uint64 `bson:"notionalLimit" json:"notionalLimit"`
	MaxTrasactionSize *mongo.Uint64 `bson:"maxTransactionSize" json:"maxTransactionSize"`
}

// NotionalLimitDetail represent a notional limit value
type NotionalLimitDetail struct {
	ID                string        `bson:"_id" json:"id"`
	ChainID           vaa.ChainID   `bson:"chainId" json:"chainId"`
	NodeName          string        `bson:"nodename" json:"nodeName"`
	NotionalLimit     *mongo.Uint64 `bson:"notionalLimit" json:"notionalLimit"`
	MaxTrasactionSize *mongo.Uint64 `bson:"maxTransactionSize" json:"maxTransactionSize"`
	CreatedAt         *time.Time    `bson:"createdAt" json:"createdAt"`
	UpdatedAt         *time.Time    `bson:"updatedAt" json:"updatedAt"`
}

// NotionalAvailable represent the available notional for chainID.
type NotionalAvailable struct {
	ChainID           vaa.ChainID   `bson:"chainid" json:"chainId"`
	AvailableNotional *mongo.Uint64 `bson:"availableNotional" json:"availableNotional"`
}

// NotionalAvailableDetail represent a notional available value.
type NotionalAvailableDetail struct {
	ID                string        `bson:"_id" json:"id"`
	ChainID           vaa.ChainID   `bson:"chainId" json:"chainId"`
	NodeName          string        `bson:"nodeName" json:"nodeName"`
	NotionalAvailable *mongo.Uint64 `bson:"availableNotional" json:"availableNotional"`
	CreatedAt         *time.Time    `bson:"createdAt" json:"createdAt"`
	UpdatedAt         *time.Time    `bson:"updatedAt" json:"updatedAt"`
}

type Emitter struct {
	Address           string       `bson:"emitteraddress" json:"emitterAddress"`
	TotalEnqueuedVaas mongo.Uint64 `bson:"totalenqueuedvaas" json:"totalEnqueuedVaas"`
	EnqueuedVaas      *int         `bson:"enqueuedvaas" json:"enqueuedVaas"`
}

// MaxNotionalAvailableRecord definition.
type MaxNotionalAvailableRecord struct {
	ID                string        `bson:"_id" json:"id"`
	ChainID           vaa.ChainID   `bson:"chainId" json:"chainId"`
	NodeName          string        `bson:"nodeName" json:"nodeName"`
	NotionalAvailable *mongo.Uint64 `bson:"availableNotional" json:"availableNotional"`
	Emitters          []Emitter     `bson:"emitters" json:"emitters"`
	CreatedAt         *time.Time    `bson:"createdAt" json:"createdAt"`
	UpdatedAt         *time.Time    `bson:"updatedAt" json:"updatedAt"`
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
	ChainID        vaa.ChainID `bson:"chainid" json:"chainId"`
	EmitterAddress string      `bson:"emitterAddress" json:"emitterAddress"`
	Sequence       string      `bson:"sequence" json:"sequence"`
	NotionalValue  int64       `bson:"notionalValue" json:"notionalValue"`
	TxHash         string      `bson:"txHash" json:"txHash"`
	ReleaseTime    int64       `bson:"releaseTime" json:"releaseTime"`
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
