package repositories

import (
	"context"
	"time"
)

type ProtocolActivityRepository interface {
	Get(ctx context.Context, from, to time.Time) (ProtocolActivity[Activity], error)
	ProtocolName() string
}

type ProtocolActivity[T any] struct {
	TotalValueSecure      float64 `json:"total_value_secure"`
	TotalValueTransferred float64 `json:"total_value_transferred"`
	Volume                float64 `json:"volume"`
	TotalMessages         uint64  `json:"total_messages"`
	Activities            []T     `json:"activity"`
}

type Activity struct {
	EmitterChainID     uint64  `json:"emitter_chain_id"`
	DestinationChainID uint64  `json:"destination_chain_id"`
	Txs                uint64  `json:"txs"`
	TotalUSD           float64 `json:"total_usd"`
}
