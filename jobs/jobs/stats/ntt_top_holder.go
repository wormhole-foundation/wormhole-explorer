package stats

import (
	"context"

	"github.com/go-resty/resty/v2"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/stats"
	"go.uber.org/zap"
)

type NTTTopHolderJob struct {
	repository *stats.HolderRepository
	log        *zap.Logger
}

func NewNTTTopHolderJob(c *resty.Client,
	arkhamUrl, arkhamApiKey, solanaUrl string,
	cacheClient cache.Cache,
	tokenProvider *domain.TokenProvider,
	notionalCache notional.NotionalLocalCacheReadable,
	log *zap.Logger) *NTTTopHolderJob {
	return &NTTTopHolderJob{
		repository: stats.NewHolderRepository(c, arkhamUrl, arkhamApiKey, solanaUrl, cacheClient, tokenProvider, notionalCache, log),
		log:        log,
	}
}

func (j *NTTTopHolderJob) Run(ctx context.Context) error {
	j.log.Info("running ntt top holder job")

	// Duration in 0 means no expiration
	err := j.repository.LoadNativeTokenTransferTopHolder(ctx, "W", 0)
	if err != nil {
		j.log.Error("failed to get top holder", zap.Error(err))
		return err
	}

	return nil
}
