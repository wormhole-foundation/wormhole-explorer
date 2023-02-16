package topic

import (
	"context"

	"go.uber.org/zap"
)

// VAAInMemoryOption represents a VAA queue in memory option function.
type VAAInMemoryOption func(*VAAInMemory)

// VAAInMemory represents VAA queue in memory.
type VAAInMemory struct {
	logger *zap.Logger
}

// NewVAAInMemory creates a VAA queue in memory instances.
func NewVAAInMemory(logger *zap.Logger) *VAAInMemory {
	m := &VAAInMemory{logger: logger}
	return m
}

// Publish sends the message to a channel.
func (i *VAAInMemory) Publish(_ context.Context, message *Event) error {

	return nil
}
