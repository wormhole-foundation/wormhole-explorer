package queue

import (
	"context"
	"time"
)

// sqsEvent represents a event data from SQS.
type sqsEvent struct {
	MessageID string `json:"MessageId"`
	Message   string `json:"Message"`
}

// Event represents a event data to be handle.
type Event struct {
	TrackID string       `json:"trackId"`
	Type    string       `json:"type"`
	Source  string       `json:"source"`
	Data    DuplicateVaa `json:"data"`
}

type DuplicateVaa struct {
	VaaID            string     `json:"vaaId"`
	ChainID          uint16     `json:"chainId"`
	Version          uint8      `json:"version"`
	GuardianSetIndex uint32     `json:"guardianSetIndex"`
	Vaa              []byte     `json:"vaas"`
	Digest           string     `json:"digest"`
	ConsistencyLevel uint8      `json:"consistencyLevel"`
	Timestamp        *time.Time `json:"timestamp"`
}

// ConsumerMessage defition.
type ConsumerMessage interface {
	Retry() uint8
	Data() *Event
	Done()
	Failed()
	IsExpired() bool
}

// ConsumeFunc is a function to consume Event.
type ConsumeFunc func(context.Context) <-chan ConsumerMessage
