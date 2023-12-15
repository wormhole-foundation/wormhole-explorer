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
	TargetChainEvent EventType = "target-chain-event"
)

type SourceChainAttributes struct {
}

type TargetChainAttributes struct {
	Emitter     string
	BlockHeight string
	ChainID     sdk.ChainID
	Status      string
	Method      string
	TxHash      string
	From        string
	To          string
}

type EventAttributes interface {
	*SourceChainAttributes | *TargetChainAttributes
}

// Event represents a event data to be handle.
type Event struct {
	TrackID        string
	Type           EventType
	ID             string
	ChainID        sdk.ChainID
	EmitterAddress string
	Sequence       string
	Timestamp      *time.Time
	TxHash         string
	Attributes     any
}

func GetAttributes[T EventAttributes](e *Event) (T, bool) {
	_, ok := interface{}(e.Attributes).(T)
	if ok {
		return e.Attributes.(T), ok
	}
	return nil, ok
}

// ConsumerMessage defition.
type ConsumerMessage interface {
	Data() *Event
	Done()
	Failed()
	IsExpired() bool
}

// ConsumeFunc is a function to consume Event.
type ConsumeFunc func(context.Context) <-chan ConsumerMessage
