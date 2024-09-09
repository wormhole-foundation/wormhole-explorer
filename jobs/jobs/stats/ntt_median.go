package stats

import (
	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"github.com/wormhole-foundation/wormhole-explorer/common/stats"
	"go.uber.org/zap"
)

type NTTMedian struct {
	nttRepository *stats.NTTRepository
	logger        *zap.Logger
}

// NewNTTMedian creates a new NTTMedian.
func NewNTTMedian(influxCli influxdb2.Client, org string, bucketInfiniteRetention string,
	cacheClient cache.Cache, logger *zap.Logger) *NTTMedian {
	return &NTTMedian{
		nttRepository: stats.NewNTTRepository(influxCli, org, bucketInfiniteRetention, cacheClient, logger),
		logger:        logger,
	}
}

// Run runs the ntt median job.
func (j *NTTMedian) Run(ctx context.Context) error {

	j.logger.Info("running ntt median job")

	// Duration in 0 means no expiration
	duration := time.Duration(0)
	err := j.nttRepository.LoadNativeTokenTransferMedian(ctx, "W", duration)
	if err != nil {
		j.logger.Error("failed to get ntt median", zap.Error(err))
		return err
	}

	return nil
}
