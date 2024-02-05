package queue

import (
	"context"
)

// Message represents a message from a queue.
type Message[T any] interface {
	Data() T
	Done(context.Context)
	Failed()
	IsExpired() bool
}

// Observation represents a signed observation.
type Observation struct {
	Addr      []byte `json:"addr"`
	Hash      []byte `json:"hash"`
	Signature []byte `json:"signature"`
	TxHash    []byte `json:"txHash"`
	MessageID string `json:"messageId"`
}
