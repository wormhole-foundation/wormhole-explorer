package queue

import (
	"context"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
)

// VAAInMemoryOption represents a VAA queue in memory option function.
type ObservationInMemoryOption func(*ObservationInMemory)

// VAAInMemory represents VAA queue in memory.
type ObservationInMemory struct {
	ch   chan Message[*gossipv1.SignedObservation]
	size int
}

// NewVAAInMemory creates a VAA queue in memory instances.
func NewObservationInMemory(opts ...ObservationInMemoryOption) *ObservationInMemory {
	m := &ObservationInMemory{size: 100}
	for _, opt := range opts {
		opt(m)
	}
	m.ch = make(chan Message[*gossipv1.SignedObservation], m.size)
	return m
}

// WithSize allows to specify an channel size when setting a value.
func ObservationWithSize(v int) ObservationInMemoryOption {
	return func(i *ObservationInMemory) {
		i.size = v
	}
}

// Publish sends the message to a channel.
func (i *ObservationInMemory) Publish(_ context.Context, o *gossipv1.SignedObservation) error {
	i.ch <- &memoryConsumerMessageQueue[*gossipv1.SignedObservation]{
		data: o,
	}
	return nil
}

// Consume returns the channel with the received messages.
func (i *ObservationInMemory) Consume(_ context.Context) <-chan Message[*gossipv1.SignedObservation] {
	return i.ch
}
