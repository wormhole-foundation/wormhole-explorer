package grpc

import (
	"github.com/wormhole-foundation/wormhole-explorer/spy/source"
	"go.uber.org/zap"
)

// Publisher represents a signed VAA publisher for subscribing customers.
type Publisher struct {
	svs    *SignedVaaSubscribers
	avs    *AllVaaSubscribers
	logger *zap.Logger
}

// NewPublisher creates a new publisher for subscribing customers.
func NewPublisher(svs *SignedVaaSubscribers, avs *AllVaaSubscribers, logger *zap.Logger) *Publisher {
	return &Publisher{svs: svs, avs: avs, logger: logger}
}

// Publish sends a signed VAA that was stored in the storage.
func (p *Publisher) Publish(e *source.Event) {
	if err := p.svs.HandleVAA(e.Vaas); err != nil {
		p.logger.Error("Failed to publish signed VAA", zap.Error(err))

	}
	if err := p.avs.HandleVAA(e.Vaas); err != nil {
		p.logger.Error("Failed to HandleGossipVAA", zap.Error(err))
	}
}
