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
	Sequence         uint64      `bson:"sequence"`
	GuardianSetIndex uint32      `bson:"guardianSetIndex"`
	Vaa              []byte      `bson:"vaas"`
	Timestamp        *time.Time  `bson:"timestamp"`
	UpdatedAt        *time.Time  `bson:"updatedAt"`
}

type ObservationUpdate struct {
	MessageID    string      `bson:"messageId"`
	ChainID      vaa.ChainID `bson:"emitterChain"`
	Emitter      string      `bson:"emitterAddr"`
	Sequence     uint64      `bson:"sequence"`
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
