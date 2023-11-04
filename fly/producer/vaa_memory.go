package producer

import (
	"context"

	"go.uber.org/zap"
)

// VAAInMemory represents VAA queue in memory.
type VAAInMemory struct {
	logger *zap.Logger
}

// NewVAAInMemory creates a VAA queue in memory instances.
func NewVAAInMemory(logger *zap.Logger) *VAAInMemory {
	m := &VAAInMemory{logger: logger}
	return m
}

// Push pushes a VAAEvent to memory.
func (m *VAAInMemory) Push(context.Context, *Notification) error {
	return nil
}
