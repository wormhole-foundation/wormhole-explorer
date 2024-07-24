package storage

import (
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type IndexingTimestamps struct {
	IndexedAt time.Time `bson:"indexedAt"`
}

type VaaUpdate struct {
	ID               string      `bson:"_id"`
	Version          uint8       `bson:"version"`
	EmitterChain     vaa.ChainID `bson:"emitterChain"`
	EmitterAddr      string      `bson:"emitterAddr"`
	Sequence         string      `bson:"sequence"`
	GuardianSetIndex uint32      `bson:"guardianSetIndex"`
	Vaa              []byte      `bson:"vaas"`
	TxHash           string      `bson:"txHash,omitempty"`
	OriginTxHash     *string     `bson:"_originTxHash,omitempty"` //this is temporary field for fix enconding txHash
	Timestamp        *time.Time  `bson:"timestamp"`
	UpdatedAt        *time.Time  `bson:"updatedAt"`
	Digest           string      `bson:"digest"`
	IsDuplicated     bool        `bson:"isDuplicated"`
}

// ToMap returns a map representation of the VaaUpdate.
func (v *VaaUpdate) ToMap() map[string]string {
	return map[string]string{
		"id":               v.ID,
		"version":          fmt.Sprint(v.Version),
		"emitterChain":     v.EmitterChain.String(),
		"emitterAddr":      v.EmitterAddr,
		"sequence":         v.Sequence,
		"guardianSetIndex": fmt.Sprint(v.GuardianSetIndex),
		"txHash":           v.TxHash,
		"timestamp":        v.Timestamp.String(),
	}
}

type DuplicateVaaUpdate struct {
	ID               string      `bson:"_id"`
	VaaID            string      `bson:"vaaId"`
	Version          uint8       `bson:"version"`
	EmitterChain     vaa.ChainID `bson:"emitterChain"`
	EmitterAddr      string      `bson:"emitterAddr"`
	Sequence         string      `bson:"sequence"`
	GuardianSetIndex uint32      `bson:"guardianSetIndex"`
	Vaa              []byte      `bson:"vaas"`
	Digest           string      `bson:"digest"`
	ConsistencyLevel uint8       `bson:"consistencyLevel"`
	TxHash           string      `bson:"txHash,omitempty"`
	Timestamp        *time.Time  `bson:"timestamp"`
	UpdatedAt        *time.Time  `bson:"updatedAt"`
}

// ToMap returns a map representation of the VaaUpdate.
func (v *DuplicateVaaUpdate) ToMap() map[string]string {
	return map[string]string{
		"id":               v.ID,
		"vaaId":            v.VaaID,
		"version":          fmt.Sprint(v.Version),
		"emitterChain":     v.EmitterChain.String(),
		"emitterAddr":      v.EmitterAddr,
		"sequence":         v.Sequence,
		"guardianSetIndex": fmt.Sprint(v.GuardianSetIndex),
		"txHash":           v.TxHash,
		"timestamp":        v.Timestamp.String(),
		"consistencyLevel": fmt.Sprint(v.ConsistencyLevel),
		"digest":           v.Digest,
	}
}

type ObservationUpdate struct {
	MessageID    string      `bson:"messageId"`
	ChainID      vaa.ChainID `bson:"emitterChain"`
	Emitter      string      `bson:"emitterAddr"`
	Sequence     string      `bson:"sequence"`
	Hash         []byte      `bson:"hash"`
	TxHash       []byte      `bson:"txHash"`
	NativeTxHash string      `bson:"nativeTxHash"`
	GuardianAddr string      `bson:"guardianAddr"`
	Signature    []byte      `bson:"signature"`
	UpdatedAt    *time.Time  `bson:"updatedAt"`
}

func (v *ObservationUpdate) ToMap() map[string]string {
	txHash, _ := domain.EncodeTrxHashByChainID(v.ChainID, v.TxHash)
	return map[string]string{
		"messageId":    v.MessageID,
		"emitterChain": v.ChainID.String(),
		"emitterAddr":  v.Emitter,
		"sequence":     v.Sequence,
		"txHash":       txHash,
		"guardianAddr": v.GuardianAddr,
	}
}

func indexedAt(t time.Time) IndexingTimestamps {
	return IndexingTimestamps{
		IndexedAt: t,
	}
}

// MongoStatus represent a mongo server status.
type MongoStatus struct {
	Ok          int32             `bson:"ok"`
	Host        string            `bson:"host"`
	Version     string            `bson:"version"`
	Process     string            `bson:"process"`
	Pid         int32             `bson:"pid"`
	Uptime      int32             `bson:"uptime"`
	Connections *MongoConnections `bson:"connections"`
}

// MongoConnections represents a mongo server connection.
type MongoConnections struct {
	Current      int32 `bson:"current"`
	Available    int32 `bson:"available"`
	TotalCreated int32 `bson:"totalCreated"`
}

type GovernorStatusUpdate struct {
	NodeName  string                      `bson:"nodename"`
	Counter   int64                       `bson:"counter"`
	Timestamp int64                       `bson:"timestamp"`
	Chains    []*ChainGovernorStatusChain `bson:"chains"`
}

type ChainGovernorStatusChain struct {
	ChainId                    uint32                        `bson:"chainid" json:"chainId"`
	RemainingAvailableNotional Uint64                        `bson:"remainingavailablenotional" json:"remainingAvailableNotional"`
	Emitters                   []*ChainGovernorStatusEmitter `bson:"emitters" json:"emitters"`
}

type ChainGovernorStatusEmitter struct {
	EmitterAddress    string                            `bson:"emitteraddress" json:"emitterAddress"`
	TotalEnqueuedVaas Uint64                            `bson:"totalenqueuedvaas" json:"totalEnqueuedVaas"`
	EnqueuedVaas      []*ChainGovernorStatusEnqueuedVAA `bson:"enqueuedvaas" json:"enqueuedVaas"`
}

type ChainGovernorStatusEnqueuedVAA struct {
	Sequence      string `bson:"sequence" json:"sequence"`
	ReleaseTime   uint32 `bson:"releasetime" json:"releaseTime"`
	NotionalValue Uint64 `bson:"notionalvalue" json:"notionalValue"`
	TxHash        string `bson:"txhash" json:"txHash"`
}

type ChainGovernorConfigUpdate struct {
	NodeName  string                      `json:"nodeName"`
	Counter   int64                       `json:"counter"`
	Timestamp int64                       `json:"timestamp"`
	Chains    []*ChainGovernorConfigChain `json:"chains"`
	Tokens    []*ChainGovernorConfigToken `json:"tokens"`
}

type ChainGovernorConfigChain struct {
	ChainId            uint32 `json:"chainId"`
	NotionalLimit      Uint64 `json:"notionalLimit"`
	BigTransactionSize Uint64 `json:"bigTransactionSize"`
}

type ChainGovernorConfigToken struct {
	OriginChainId uint32  `json:"originChainId"`
	OriginAddress string  `json:"originAddress"`
	Price         float32 `json:"price"`
}
