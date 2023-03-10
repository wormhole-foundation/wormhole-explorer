package storage

import (
	"time"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// DestinationTx representa a destination transaction.
type DestinationTx struct {
	ChainID     vaa.ChainID `bson:"chainId"`
	Status      string      `bson:"status"`
	Method      string      `bson:"method"`
	TxHash      string      `bson:"txHash"`
	From        string      `bson:"from"`
	To          string      `bson:"to"`
	BlockNumber string      `bson:"blockNumber"`
	Timestamp   string      `bson:"timestamp"`
	UpdatedAt   *time.Time  `bson:"updatedAt"`
}

type IndexingTimestamps struct {
	IndexedAt time.Time `bson:"indexedAt"`
}

// TransactionUpdate represents a transaction document.
type TransactionUpdate struct {
	ID          string        `bson:"_id"`
	Destination DestinationTx `bson:"destinationTx"`
}

type WatcherBlock struct {
	ID          string    `bson:"_id"`
	BlockNumber int64     `bson:"blockNumber"`
	UpdatedAt   time.Time `bson:"updatedAt"`
}
