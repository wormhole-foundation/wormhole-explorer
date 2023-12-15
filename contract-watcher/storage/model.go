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
	Timestamp   *time.Time  `bson:"timestamp"`
	UpdatedAt   *time.Time  `bson:"updatedAt"`
}

// TransactionUpdate represents a transaction document.
type TransactionUpdate struct {
	ID          string        `bson:"_id"`
	Destination DestinationTx `bson:"destinationTx"`
}

func (t *TransactionUpdate) ToMap() map[string]string {
	return map[string]string{
		"id":                      t.ID,
		"destination.chainId":     t.Destination.ChainID.String(),
		"destination.status":      t.Destination.Status,
		"destination.method":      t.Destination.Method,
		"destination.txHash":      t.Destination.TxHash,
		"destination.from":        t.Destination.From,
		"destination.to":          t.Destination.To,
		"destination.blockNumber": t.Destination.BlockNumber,
	}
}

type WatcherBlock struct {
	ID          string    `bson:"_id"`
	BlockNumber int64     `bson:"blockNumber"`
	UpdatedAt   time.Time `bson:"updatedAt"`
}
