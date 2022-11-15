package governor

import (
	"time"

	"github.com/certusone/wormhole/node/pkg/vaa"
)

// GovConfigPage definition.
type GovConfig struct {
	ID        string              `bson:"_id" json:"id"`
	CreatedAt *time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt *time.Time          `bson:"updatedAt" json:"updatedAt"`
	NodeName  string              `bson:"nodename" json:"nodename"`
	Counter   int                 `bson:"counter" json:"counter"`
	Chains    []*GovConfigChains  `bson:"chains" json:"chains"`
	Tokens    []*GovConfigfTokens `bson:"tokens" json:"tokens"`
}

type GovConfigChains struct {
	ChainID            vaa.ChainID `bson:"chainid" json:"chainid"`
	NotionalLimit      int64       `bson:"notionallimit" json:"notionallimit"`
	BigTransactionSize int64       `bson:"bigtransactionsize" json:"bigtransactionsize"`
}

type GovConfigfTokens struct {
	OriginChainID int     `bson:"originchainid" json:"originchainid"`
	OriginAddress string  `bson:"originaddress" json:"originaddress"`
	Price         float64 `bson:"price" json:"price"`
}

// GovStatusPage definition.
type GovStatus struct {
	ID        string             `bson:"_id" json:"id"`
	CreatedAt *time.Time         `bson:"createdAt" json:"createdAt"`
	UpdatedAt *time.Time         `bson:"updatedAt" json:"updatedAt"`
	NodeName  string             `bson:"nodename" json:"nodename"`
	Chains    []*GovStatusChains `bson:"chains" json:"chains"`
}

type GovStatusChains struct {
	ChainID                    vaa.ChainID              `bson:"chainid" json:"chainid"`
	RemainingAvailableNotional int64                    `bson:"remainingavailablenotional" json:"remainingavailablenotional"`
	Emitters                   []*GovStatusChainEmitter `bson:"emitters" json:"emitters"`
}

type GovStatusChainEmitter struct {
	EmitterAddress    string      `bson:"emitteraddress" json:"emitteraddress"`
	TotalEnqueuedVaas int         `bson:"totalenqueuedvaas" json:"totalenqueuedvaas"`
	EnqueuedVass      interface{} `bson:"enqueuedvaas" json:"enqueuedvaas"`
}

// NotionalLimit definition.
type NotionalLimit struct {
	ChainID           vaa.ChainID `bson:"chainid" json:"chainid"`
	NotionalLimit     *int64      `bson:"notionalLimit" json:"notionalLimit"`
	MaxTrasactionSize *int64      `bson:"maxTransactionSize" json:"maxTransactionSize"`
}

// NotionalLimitDetail definition.
type NotionalLimitDetail struct {
	ID                string      `bson:"_id" json:"id"`
	ChainID           vaa.ChainID `bson:"chainid" json:"chainid"`
	NodeName          string      `bson:"nodename" json:"nodename"`
	NotionalLimit     *int64      `bson:"notionalLimit" json:"notionalLimit"`
	MaxTrasactionSize *int64      `bson:"maxTransactionSize" json:"maxTransactionSize"`
	CreatedAt         *time.Time  `bson:"createdAt" json:"createdAt"`
	UpdatedAt         *time.Time  `bson:"updatedAt" json:"updatedAt"`
}

type NotionalAvailable struct {
	ChainID           vaa.ChainID `bson:"chainid" json:"chainId"`
	AvailableNotional *int64      `bson:"availableNotional" json:"availableNotional"`
}

type NotionalAvailableDetail struct {
	ID                string      `bson:"_id" json:"id"`
	ChainID           vaa.ChainID `bson:"chainId" json:"chainId"`
	NodeName          string      `bson:"nodeName" json:"nodeName"`
	NotionalAvailable *int64      `bson:"availableNotional" json:"availableNotional"`
	CreatedAt         *time.Time  `bson:"createdAt" json:"createdAt"`
	UpdatedAt         *time.Time  `bson:"updatedAt" json:"updatedAt"`
}

type Emitter struct {
	Address           string `bson:"emitteraddress" json:"emitterAddress"`
	TotalEnqueuedVaas int    `bson:"totalenqueuedvaas" json:"totalEnqueuedVaas"`
	EnqueuedVaas      *int   `bson:"enqueuedvaas" json:"enqueuedVaas"`
}

// MaxNotionalAvailableRecord definition.
type MaxNotionalAvailableRecord struct {
	ID                string      `bson:"_id" json:"id"`
	ChainID           vaa.ChainID `bson:"chainId" json:"chainId"`
	NodeName          string      `bson:"nodeName" json:"nodeName"`
	NotionalAvailable *int64      `bson:"availableNotional" json:"availableNotional"`
	Emitters          []Emitter   `bson:"emitters" json:"emitters"`
	CreatedAt         *time.Time  `bson:"createdAt" json:"createdAt"`
	UpdatedAt         *time.Time  `bson:"updatedAt" json:"updatedAt"`
}

// EnqueuedVaa definition.
type EnqueuedVaa struct {
	ChainID        vaa.ChainID `bson:"chainId" json:"chainId"`
	EmitterAddress string      `bson:"emitterAddress" json:"emitterAddress"`
	Sequence       int64       `bson:"sequence" json:"sequence"`
	NotionalValue  int64       `bson:"notionalValue" json:"notionalValue"`
	TxHash         string      `bson:"txHash" json:"txHash"`
}

type EnqueuedVaas struct {
	ChainID     vaa.ChainID    `bson:"chainid" json:"chainId"`
	EnqueuedVaa []*EnqueuedVaa `bson:"enqueuedVaas" json:"enqueuedVaas"`
}

// EnqueuedVaaDetail definition.
type EnqueuedVaaDetail struct {
	ChainID        vaa.ChainID `bson:"chainid" json:"chainid"`
	EmitterAddress string      `bson:"emitterAddress" json:"emitterAddress"`
	Sequence       int64       `bson:"sequence" json:"sequence"`
	NotionalValue  int64       `bson:"notionalValue" json:"notionalValue"`
	TxHash         string      `bson:"txHash" json:"txHash"`
	ReleaseTime    int64       `bson:"releaseTime" json:"releaseTime"`
}

// GovernorLimit definition.
type GovernorLimit struct {
	ChainID            vaa.ChainID `bson:"chainId" json:"chainId"`
	AvailableNotional  int64       `bson:"availableNotional" json:"availableNotional"`
	NotionalLimit      int64       `bson:"notionalLimit" json:"notionalLimit"`
	MaxTransactionSize int64       `bson:"maxTransactionSize" json:"maxTransactionSize"`
}
