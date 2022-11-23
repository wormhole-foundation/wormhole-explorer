package vaa

import (
	"time"

	"github.com/certusone/wormhole/node/pkg/vaa"
)

// VaaDoc represent an vaa document.
type VaaDoc struct {
	ID               string      `bson:"_id" json:"id"`
	Version          uint8       `bson:"version" json:"version"`
	EmitterChain     vaa.ChainID `bson:"emitterChain" json:"emitterChain"`
	EmitterAddr      string      `bson:"emitterAddr" json:"emitterAddr"`
	Sequence         uint64      `bson:"sequence" json:"sequence"`
	GuardianSetIndex uint32      `bson:"guardianSetIndex" json:"guardianSetIndex"`
	Vaa              []byte      `bson:"vaas" json:"vaa"`
	Timestamp        *time.Time  `bson:"timestamp" json:"timestamp"`

	UpdatedAt *time.Time `bson:"updatedAt" json:"updatedAt"`
	IndexedAt *time.Time `bson:"indexedAt" json:"indexedAt"`
}

// VaaStats definition.
type VaaStats struct {
	ChainID vaa.ChainID `bson:"_id" json:"chainId"`
	Count   uint        `bson:"count" json:"count"`
}
