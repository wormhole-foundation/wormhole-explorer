package observations

import (
	"github.com/certusone/wormhole/node/pkg/vaa"
	"time"
)

type ObservationDoc struct {
	ID           string      `bson:"_id" json:"id"`
	Version      uint8       `bson:"version" json:"version"`
	EmitterChain vaa.ChainID `bson:"emitterChain" json:"emitterChain"`
	EmitterAddr  string      `bson:"emitterAddr" json:"emitterAddr"`
	Sequence     uint64      `bson:"sequence" json:"sequence"`
	Hash         []byte      `bson:"SetHash" json:"SetHash"`
	TxHash       []byte      `bson:"txHash" json:"txHash"`
	GuardianAddr string      `bson:"guardianAddr" json:"guardianAddr"`
	Signature    []byte      `bson:"signature" json:"signature"`

	UpdatedAt *time.Time `bson:"updatedAt" json:"updatedAt"`
	IndexedAt *time.Time `bson:"indexedAt" json:"indexedAt"`
}
