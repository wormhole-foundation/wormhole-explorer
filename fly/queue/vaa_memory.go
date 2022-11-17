package queue

import (
	"context"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// VAAInMemoryOption represents a VAA queue in memory option function.
type VAAInMemoryOption func(*VAAInMemory)

// VAAInMemory represents VAA queue in memory.
type VAAInMemory struct {
	ch   chan *Message
	size int
}

// NewVAAInMemory creates a VAA queue in memory instances.
func NewVAAInMemory(opts ...VAAInMemoryOption) *VAAInMemory {
	m := &VAAInMemory{size: 100}
	for _, opt := range opts {
		opt(m)
	}
	m.ch = make(chan *Message, m.size)
	return m
}

// WithSize allows to specify an channel size when setting a value.
func WithSize(v int) VAAInMemoryOption {
	return func(i *VAAInMemory) {
		i.size = v
	}
}

// Publish sends the message to a channel.
func (i *VAAInMemory) Publish(_ context.Context, v *vaa.VAA, data []byte) error {
	i.ch <- &Message{
		Data:      data,
		Ack:       func() {},
		IsExpired: func() bool { return false },
	}
	return nil
}

// Consume returns the channel with the received messages.
func (i *VAAInMemory) Consume(_ context.Context) <-chan *Message {
	return i.ch
}
