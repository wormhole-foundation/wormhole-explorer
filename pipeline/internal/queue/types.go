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

// Event represents a event data to be handled.
type Event struct {
	Source           string
	TrackID          string
	ID               string      `json:"id"`
	VaaID            string      `json:"vaaId"`
	EmitterChainID   sdk.ChainID `json:"emitterChain"`
	EmitterAddress   string      `json:"emitterAddress"`
	Sequence         uint64      `json:"sequence"`
	GuardianSetIndex uint32      `json:"guardianSetIndex"`
	Timestamp        time.Time   `json:"timestamp"`
	Vaa              []byte      `json:"vaa"`
	TxHash           string      `json:"txHash"`
	Version          int         `json:"version"`
}

// ConsumerMessage definition.
type ConsumerMessage interface {
	Retry() uint8
	Data() *Event
	Done()
	Failed()
	IsExpired() bool
	SentTimestamp() *time.Time
}

// ConsumeFunc is a function to consume Event.
type ConsumeFunc func(context.Context) <-chan ConsumerMessage
