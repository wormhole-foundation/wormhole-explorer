package stats

import (
	"context"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/stats"
	"go.uber.org/zap"
)

type NttTopHolderJob struct {
	repository *stats.HolderRepository
	log        *zap.Logger
}

func NewNttTopHolderJob(holderRepository *stats.HolderRepository, log *zap.Logger) *NttTopHolderJob {
	return &NttTopHolderJob{
		repository: holderRepository,
		log:        log,
	}
}

func (j *NttTopHolderJob) Run(ctx context.Context) error {
	j.log.Info("running ntt top holder job")
	duration := 1 * time.Hour

	err := j.repository.LoadNativeTokenTransferTopHolder(ctx, "W", duration)
	if err != nil {
		j.log.Error("failed to get top holder", zap.Error(err))
		return err
	}

	return nil
}
