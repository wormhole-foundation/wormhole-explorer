package storage

import (
	"time"

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
	Timestamp        *time.Time  `bson:"timestamp"`
	UpdatedAt        *time.Time  `bson:"updatedAt"`
}

type ObservationUpdate struct {
	MessageID    string      `bson:"messageId"`
	ChainID      vaa.ChainID `bson:"emitterChain"`
	Emitter      string      `bson:"emitterAddr"`
	Sequence     string      `bson:"sequence"`
	Hash         []byte      `bson:"hash"`
	TxHash       []byte      `bson:"txHash"`
	GuardianAddr string      `bson:"guardianAddr"`
	Signature    []byte      `bson:"signature"`
	UpdatedAt    *time.Time  `bson:"updatedAt"`
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
	ChainId                    uint32                        `bson:"chainid"`
	RemainingAvailableNotional Uint64                        `bson:"remainingavailablenotional"`
	Emitters                   []*ChainGovernorStatusEmitter `bson:"emitters"`
}

type ChainGovernorStatusEmitter struct {
	EmitterAddress    string                            `bson:"emitteraddress"`
	TotalEnqueuedVaas Uint64                            `bson:"totalenqueuedvaas"`
	EnqueuedVaas      []*ChainGovernorStatusEnqueuedVAA `bson:"enqueuedvaas"`
}

type ChainGovernorStatusEnqueuedVAA struct {
	Sequence      string `bson:"sequence"`
	ReleaseTime   uint32 `bson:"releasetime"`
	NotionalValue Uint64 `bson:"notionalvalue"`
	TxHash        string `bson:"txhash"`
}

type ChainGovernorConfigUpdate struct {
	NodeName  string
	Counter   int64
	Timestamp int64
	Chains    []*ChainGovernorConfigChain
	Tokens    []*ChainGovernorConfigToken
}

type ChainGovernorConfigChain struct {
	ChainId            uint32
	NotionalLimit      Uint64
	BigTransactionSize Uint64
}

type ChainGovernorConfigToken struct {
	OriginChainId uint32
	OriginAddress string
	Price         float32
}
