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

// Event represents a vaa data to be handle by the pipeline.
type Event struct {
	ID             string
	ChainID        sdk.ChainID
	EmitterAddress string
	Sequence       string
	Vaa            []byte
	Timestamp      *time.Time
	TxHash         string
	TrackID        string
}

// ConsumerMessage defition.
type ConsumerMessage interface {
	Data() *Event
	Done()
	Failed()
	IsExpired() bool
}

// VAAConsumeFunc is a function to consume VAAEvent.
type VAAConsumeFunc func(context.Context) <-chan ConsumerMessage
