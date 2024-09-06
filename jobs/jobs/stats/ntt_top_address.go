package stats

import (
	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"github.com/wormhole-foundation/wormhole-explorer/common/stats"
	"go.uber.org/zap"
)

type NTTTopAddressJob struct {
	statsRepositorty *stats.AddressRepository
	logger           *zap.Logger
}

// NewNTTTopAddressJob creates a new NttTopAddressJob.
func NewNTTTopAddressJob(influxCli influxdb2.Client, org string, bucketInfiniteRetention string,
	cacheClient cache.Cache, logger *zap.Logger) *NTTTopAddressJob {
	return &NTTTopAddressJob{
		statsRepositorty: stats.NewAddressRepository(influxCli, org, bucketInfiniteRetention, cacheClient, logger),
		logger:           logger,
	}
}

// Run runs the transfer report job.
func (j *NTTTopAddressJob) Run(ctx context.Context) error {

	j.logger.Info("running ntt top address job")

	// Duration in 0 means no expiration
	duration := time.Duration(0)

	err := j.statsRepositorty.LoadNativeTokenTransferTopAddress(ctx, "W", true, duration)
	if err != nil {
		j.logger.Error("failed to get top address by volume", zap.Error(err))
		return err
	}

	err = j.statsRepositorty.LoadNativeTokenTransferTopAddress(ctx, "W", false, duration)
	if err != nil {
		j.logger.Error("failed to get top address by count", zap.Error(err))
		return err
	}

	return nil
}
