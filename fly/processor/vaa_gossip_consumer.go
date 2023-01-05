package processor

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/fly/deduplicator"

	"github.com/certusone/wormhole/node/pkg/common"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type vaaGossipConsumer struct {
	gst            *common.GuardianSetState
	nonPythProcess VAAPushFunc
	pythProcess    VAAPushFunc
	logger         *zap.Logger
	deduplicator   *deduplicator.Deduplicator
}

// NewVAAGossipConsumer creates a new processor instances.
func NewVAAGossipConsumer(
	gst *common.GuardianSetState,
	deduplicator *deduplicator.Deduplicator,
	nonPythPublish VAAPushFunc,
	pythPublish VAAPushFunc,
	logger *zap.Logger) *vaaGossipConsumer {
	return &vaaGossipConsumer{
		gst:            gst,
		deduplicator:   deduplicator,
		nonPythProcess: nonPythPublish,
		pythProcess:    pythPublish,
		logger:         logger,
	}
}

// Push handles incoming VAAs depending on whether it is a pyth or non pyth.
func (p *vaaGossipConsumer) Push(ctx context.Context, v *vaa.VAA, serializedVaa []byte) error {
	if err := v.Verify(p.gst.Get().Keys); err != nil {
		p.logger.Error("Received invalid vaa", zap.String("id", v.MessageID()))
		return err
	}

	err := p.deduplicator.Apply(ctx, v.MessageID(), func() error {
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
