package processor

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/fly/deduplicator"
	"github.com/wormhole-foundation/wormhole-explorer/fly/guardiansets"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type vaaGossipConsumer struct {
	guardianSetHistory *guardiansets.GuardianSetHistory
	nonPythProcess     VAAPushFunc
	pythProcess        VAAPushFunc
	logger             *zap.Logger
	deduplicator       *deduplicator.Deduplicator
	metrics            metrics.Metrics
}

// NewVAAGossipConsumer creates a new processor instances.
func NewVAAGossipConsumer(
	guardianSetHistory *guardiansets.GuardianSetHistory,
	deduplicator *deduplicator.Deduplicator,
	nonPythPublish VAAPushFunc,
	pythPublish VAAPushFunc,
	metrics metrics.Metrics,
	logger *zap.Logger,
) *vaaGossipConsumer {

	return &vaaGossipConsumer{
		guardianSetHistory: guardianSetHistory,
		deduplicator:       deduplicator,
		nonPythProcess:     nonPythPublish,
		pythProcess:        pythPublish,
		metrics:            metrics,
		logger:             logger,
	}
}

// Push handles incoming VAAs depending on whether it is a pyth or non pyth.
func (p *vaaGossipConsumer) Push(ctx context.Context, v *vaa.VAA, serializedVaa []byte) error {

	if err := p.guardianSetHistory.Verify(ctx, v); err != nil {
		p.logger.Error("Received invalid vaa", zap.String("id", v.MessageID()))
		return err
	}

	err := p.deduplicator.Apply(ctx, v.MessageID(), func() error {
		p.metrics.IncVaaUnfiltered(v.EmitterChain)
		if vaa.ChainIDPythNet == v.EmitterChain {
			return p.pythProcess(ctx, v, serializedVaa)
		}
		return p.nonPythProcess(ctx, v, serializedVaa)
	})

	if err != nil {
		p.logger.Error("Error consuming from Gossip network",
			zap.String("id", v.MessageID()),
			zap.Error(err))
		return err
	}

	return nil
}
