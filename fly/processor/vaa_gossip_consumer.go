package processor

import (
	"context"
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/fly/deduplicator"
	"github.com/wormhole-foundation/wormhole-explorer/fly/guardiansets"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type vaaGossipConsumer struct {
	guardianSetHistory *guardiansets.LegacyGuardianSetHistory
	nonPythProcess     VAAPushFunc
	pythProcess        VAAPushFunc
	logger             *zap.Logger
	nonPythDedup       *deduplicator.Deduplicator
	pythDedup          *deduplicator.Deduplicator
	metrics            metrics.Metrics
	repository         storage.Storage
}

// NewVAAGossipConsumer creates a new processor instances.
func NewVAAGossipConsumer(
	guardianSetHistory *guardiansets.LegacyGuardianSetHistory,
	nonPythDedup *deduplicator.Deduplicator,
	pythDedup *deduplicator.Deduplicator,
	nonPythPublish VAAPushFunc,
	pythPublish VAAPushFunc,
	metrics metrics.Metrics,
	repository storage.Storage,
	logger *zap.Logger,
) *vaaGossipConsumer {

	return &vaaGossipConsumer{
		guardianSetHistory: guardianSetHistory,
		nonPythDedup:       nonPythDedup,
		pythDedup:          pythDedup,
		nonPythProcess:     nonPythPublish,
		pythProcess:        pythPublish,
		metrics:            metrics,
		repository:         repository,
		logger:             logger,
	}
}

// Push handles incoming VAAs depending on whether it is a pyth or non pyth.
func (p *vaaGossipConsumer) Push(ctx context.Context, v *vaa.VAA, serializedVaa []byte) error {

	uniqueVaaID := domain.CreateUniqueVaaID(v)
	if err := p.guardianSetHistory.Verify(ctx, v); err != nil {
		p.logger.Error("Received invalid vaa", zap.String("id", uniqueVaaID))
		return err
	}

	key := fmt.Sprintf("vaa:%s", uniqueVaaID)
	var err error
	if vaa.ChainIDPythNet == v.EmitterChain {
		err = p.pythDedup.Apply(ctx, key, func() error {
			p.metrics.IncVaaUnfiltered(v.EmitterChain)
			return p.pythProcess(ctx, v, serializedVaa)
		})
	} else {
		err = p.nonPythDedup.Apply(ctx, key, func() error {
			p.metrics.IncVaaUnfiltered(v.EmitterChain)
			pErr := p.nonPythProcess(ctx, v, serializedVaa)
			if pErr != nil {
				p.logger.Error("Error processing vaa", zap.String("id", uniqueVaaID), zap.Error(err))
				// This is the fallback to store the vaa in the repository.
				pErr = p.repository.UpsertVAA(ctx, v, serializedVaa)
				if pErr != nil {
					p.logger.Error("Error inserting vaa in repository as fallback", zap.String("id", uniqueVaaID), zap.Error(err))
				}
			}
			return pErr
		})

	}

	if err != nil {
		p.logger.Error("Error consuming from Gossip network",
			zap.String("id", uniqueVaaID),
			zap.Error(err))
		return err
	}

	return nil
}
