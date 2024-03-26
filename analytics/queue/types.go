package queue

import (
	"context"
	"time"
)

type sqsEvent struct {
	MessageID string `json:"MessageId"`
	Message   string `json:"Message"`
}

// Event represents a event data to be handle.
type Event struct {
	Source         string
	TrackID        string
	ID             string
	ChainID        uint16
	EmitterAddress string
	Sequence       string
	Vaa            []byte
	Timestamp      *time.Time
	VaaIsSigned    bool
}

// ConsumerMessage defition.
type ConsumerMessage interface {
	Retry() uint8
	Data() *Event
	Done()
	Failed()
	IsExpired() bool
}

// ConsumeFunc is a function to consume VAAEvent.
type ConsumeFunc func(context.Context) <-chan ConsumerMessage
