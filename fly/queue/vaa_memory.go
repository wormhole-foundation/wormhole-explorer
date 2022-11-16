package queue

import (
	"context"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type VAAInMemoryOption func(*VAAInMemory)

type VAAInMemory struct {
	ch   chan *Message
	size int
}

func NewVAAInMemory(opts ...VAAInMemoryOption) *VAAInMemory {
	m := &VAAInMemory{size: 100}
	for _, opt := range opts {
		opt(m)
	}
	m.ch = make(chan *Message, m.size)
	return m
}

func WithSize(v int) VAAInMemoryOption {
	return func(i *VAAInMemory) {
		i.size = v
	}
}

func (i *VAAInMemory) Publish(_ context.Context, v *vaa.VAA, data []byte) error {
	i.ch <- &Message{
		Data: data,
		Ack:  func() {},
	}
	return nil
}

func (i *VAAInMemory) Consume(_ context.Context) <-chan *Message {
	return i.ch
}
