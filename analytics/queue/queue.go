package queue

import (
	"context"
	"time"
)

type sqsEvent struct {
	MessageID string `json:"MessageId"`
	Message   string `json:"Message"`
}

// Event represents a vaa event to be handle by the pipeline.
type Event struct {
	ID             string     `json:"id"`
	ChainID        uint16     `json:"emitterChain"`
	EmitterAddress string     `json:"emitterAddr"`
	Sequence       uint64     `json:"sequence"`
	Vaa            []byte     `json:"vaas"`
	Timestamp      *time.Time `json:"timestamp"`
	TxHash         string     `json:"txHash"`
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
