package pipeline

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/pipeline/topic"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/watcher"
	"go.uber.org/zap"
)

// Publisher definition.
type Publisher struct {
	logger   *zap.Logger
	pushFunc topic.PushFunc
}

// NewPublisher creates a new publisher for vaa with parse configuration.
func NewPublisher(pushFunc topic.PushFunc, logger *zap.Logger) *Publisher {
	return &Publisher{logger: logger, pushFunc: pushFunc}
}

// Publish sends a Event for the vaa that has parse configuration defined.
func (p *Publisher) Publish(ctx context.Context, e *watcher.Event) {

	// create a Event.
	event := topic.Event{
		ID:               e.ID,
		ChainID:          e.ChainID,
		EmitterAddress:   e.EmitterAddress,
		Sequence:         e.Sequence,
		GuardianSetIndex: e.GuardianSetIndex,
		Vaa:              e.Vaa,
		IndexedAt:        e.IndexedAt,
		Timestamp:        e.Timestamp,
		UpdatedAt:        e.UpdatedAt,
		TxHash:           e.TxHash,
		Version:          e.Version,
		Revision:         e.Revision,
	}

	// push messages to topic.
	err := p.pushFunc(ctx, &event)
	if err != nil {
		p.logger.Error("can not push event to topic", zap.Error(err), zap.String("event", event.ID))
	}
}
