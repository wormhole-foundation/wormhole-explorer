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
	// Raw fee fields for evm
	GasUsed           *string
	EffectiveGasPrice *string
	// Raw fee fields for solana
	Fee *uint64
}

type EventAttributes interface {
	*SourceChainAttributes | *TargetChainAttributes
}

// Event represents a event data to be handled.
type Event struct {
	Source         string
	TrackID        string
	Type           EventType
	VaaID          string // chain/address/sequence
	ID             string // digest
	ChainID        sdk.ChainID
	EmitterAddress string
	Sequence       string
	Timestamp      *time.Time
	TxHash         string
	Vaa            []byte
	IsVaaSigned    bool
	Attributes     any
	Overwrite      bool
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
	Retry() uint8
	Data() *Event
	Done()
	Failed()
	IsExpired() bool
	SentTimestamp() *time.Time
}

// ConsumeFunc is a function to consume Event.
type ConsumeFunc func(context.Context) <-chan ConsumerMessage
