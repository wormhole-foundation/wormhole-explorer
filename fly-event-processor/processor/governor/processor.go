package governor

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/storage"
	"go.uber.org/zap"
)

type Processor struct {
	repository *storage.Repository
	logger     *zap.Logger
	metrics    metrics.Metrics
}

func NewProcessor(repository *storage.Repository, logger *zap.Logger, metrics metrics.Metrics) *Processor {
	return &Processor{
		repository: repository,
		logger:     logger,
		metrics:    metrics,
	}
}

func (p *Processor) Process(ctx context.Context, params *Params) error {
	// logger := p.logger.With(
	// 	zap.String("trackId", params.TrackID),
	// )

	return nil
}
