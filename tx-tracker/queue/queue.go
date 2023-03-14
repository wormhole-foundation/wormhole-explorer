package queue

import (
	"context"
	"time"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type sqsEvent struct {
	MessageID string `json:"MessageId"`
	Message   string `json:"Message"`
}

// VaaEvent represents a vaa data to be handle by the pipeline.
type VaaEvent struct {
	ID               string      `json:"id"`
	ChainID          sdk.ChainID `json:"emitterChain"`
	EmitterAddress   string      `json:"emitterAddr"`
	Sequence         string      `json:"sequence"`
	GuardianSetIndex uint32      `json:"guardianSetIndex"`
	Vaa              []byte      `json:"vaas"`
	IndexedAt        time.Time   `json:"indexedAt"`
	Timestamp        *time.Time  `json:"timestamp"`
	UpdatedAt        *time.Time  `json:"updatedAt"`
	TxHash           string      `json:"txHash"`
	Version          uint16      `json:"version"`
	Revision         uint16      `json:"revision"`
}

// ConsumerMessage defition.
type ConsumerMessage interface {
	Data() *VaaEvent
	Done()
	Failed()
	IsExpired() bool
}

// VAAConsumeFunc is a function to consume VAAEvent.
type VAAConsumeFunc func(context.Context) <-chan ConsumerMessage
