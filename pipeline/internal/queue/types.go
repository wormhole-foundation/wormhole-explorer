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

type EventType string

const (
	SourceChainEvent EventType = "source-chain-event"
)

// Event represents a event data to be handled.
type Event struct {
	Source         string
	TrackID        string
	Type           EventType
	ID             string
	VaaId          string
	ChainID        sdk.ChainID
	EmitterAddress string
	Sequence       string
	Timestamp      *time.Time
	Vaa            []byte
	IsVaaSigned    bool
	Attributes     any
	Overwrite      bool
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
