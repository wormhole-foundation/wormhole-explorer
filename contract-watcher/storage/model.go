package storage

import (
	"time"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// RedeemedUpdate representa a redeemed document.
type RedeemedUpdate struct {
	ID           string      `bson:"_id"`
	Chain        string      `bson:"chain"`
	EmitterChain vaa.ChainID `bson:"emitterChain"`
	EmitterAddr  string      `bson:"emitterAddr"`
	Sequence     string      `bson:"sequence"`
	Method       string      `bson:"method"`
	Status       string      `bson:"status"`
	TxHash       string      `bson:"txHash"`
	From         string      `bson:"from"`
	To           string      `bson:"to"`
	BlockNumber  string      `bson:"blockNumber"`
	VaaTimestamp *time.Time  `bson:"vaaTimestamp"`
	OriginTxHash string      `bson:"originTxHash"`
	Timestamp    *time.Time  `bson:"timestamp"`
	UpdatedAt    *time.Time  `bson:"updatedAt"`
}

type IndexingTimestamps struct {
	IndexedAt time.Time `bson:"indexedAt"`
}

// TransactionUpdate represents a transaction document.
type TransactionUpdate struct {
}
