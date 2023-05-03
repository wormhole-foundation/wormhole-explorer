package metric

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/wormhole-foundation/wormhole-explorer/analytic/config"
	wormscanCache "github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	wormscanNotionalCache "github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Metric definition.
type Metric struct {
	influxCli     influxdb2.Client
	writeApi      api.WriteAPIBlocking
	logger        *zap.Logger
	notionalCache wormscanNotionalCache.NotionalLocalCacheReadable
}

// New create a new *Metric.
func New(
	influxCli influxdb2.Client,
	cfg *config.Configuration,
	logger *zap.Logger,
) *Metric {

	writeAPI := influxCli.WriteAPIBlocking(cfg.InfluxOrganization, cfg.InfluxBucket)

	//FIXME ctx
	_, notionalCache := newCache(context.Background(), cfg, logger)

	m := Metric{
		influxCli:     influxCli,
		writeApi:      writeAPI,
		logger:        logger,
		notionalCache: notionalCache,
	}

	return &m
}

func newCache(ctx context.Context, cfg *config.Configuration, logger *zap.Logger) (wormscanCache.CacheReadable, wormscanNotionalCache.NotionalLocalCacheReadable) {

	// use a distributed cache and for notional a pubsub to sync local cache.
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.CacheURL})

	// get cache client
	//FIXME check return value
	cacheClient, _ := wormscanCache.NewCacheClient(redisClient, true /*enabled*/, logger)

	// get notional cache client and init load to local cache
	//FIXME check return value
	notionalCache, _ := wormscanNotionalCache.NewNotionalCache(ctx, redisClient, cfg.CacheChannel, logger)
	notionalCache.Init(ctx)

	return cacheClient, notionalCache
}

// Push implement MetricPushFunc definition.
func (m *Metric) Push(ctx context.Context, vaa *sdk.VAA) error {
	return m.vaaCountMeasurement(ctx, vaa)
}

// Close influx client.
func (m *Metric) Close() {
	m.influxCli.Close()
}

// vaaCountMeasurement handle the push of metric point for measurement vaa_count.
func (m *Metric) vaaCountMeasurement(ctx context.Context, vaa *sdk.VAA) error {

	// Create a new point for the `vaa_count` measurement.
	{
		const measurement = "vaa_count"

		point := influxdb2.
			NewPointWithMeasurement(measurement).
			AddTag("chain_id", strconv.Itoa(int(vaa.EmitterChain))).
			AddField("count", 1).
			SetTime(vaa.Timestamp.Add(time.Nanosecond * time.Duration(vaa.Sequence)))

		// Write the point to influx
		err := m.writeApi.WritePoint(ctx, point)
		if err != nil {
			m.logger.Error("failed to write metric",
				zap.String("measurement", measurement),
				zap.Uint16("chain_id", uint16(vaa.EmitterChain)),
				zap.Error(err),
			)
			return err
		}
	}

	// Create a new point for the `vaa_volume` measurement.
	{
		const measurement = "vaa_volume"

		// Decode the VAA payload
		//
		// If the VAA didn't come from the portal token bridge, we just skip it.
		payload, err := sdk.DecodeTransferPayloadHdr(vaa.Payload)
		if err != nil {
			return nil
		}

		// Get the token metadata from a static, in-memory dictionary
		//
		// This dictionary contains complementary data about the token that is not present in the VAA itself.
		tokenMeta, ok := domain.GetTokenMetadata(payload.OriginChain, "0x"+payload.OriginAddress.String())
		if !ok {
			m.logger.Warn("found no token metadata for VAA",
				zap.String("originAddress", payload.OriginAddress.String()),
				zap.Uint16("originChain", uint16(payload.OriginChain)),
			)
			return nil
		}

		// Normalize the amount to 8 decimals
		amount := payload.Amount
		if tokenMeta.Decimals < 8 {

			// factor = 10 ^ (8 - tokenMeta.Decimals)
			var factor big.Int
			factor.Exp(big.NewInt(10), big.NewInt(int64(8-tokenMeta.Decimals)), nil)

			amount = amount.Mul(amount, &factor)
		}

		// Try to obtain the token notional value from the cache
		notional, err := m.notionalCache.Get(tokenMeta.UnderlyingSymbol)
		if err != nil {
			return nil
		}

		// Create a data point with volume-related fields
		point := influxdb2.NewPointWithMeasurement(measurement).
			AddTag("chain_source_id", fmt.Sprintf("%d", payload.OriginChain)).
			AddTag("chain_destination_id", fmt.Sprintf("%d", payload.TargetChain)).
			AddTag("app_id", domain.AppIdPortalTokenBridge).
			AddTag("symbol", tokenMeta.UnderlyingSymbol).
			AddField("amount", amount).
			AddField("notional", notional.NotionalUsd).
			SetTime(vaa.Timestamp)

		// Write the point to influx
		err = m.writeApi.WritePoint(ctx, point)
		if err != nil {
			return err
		}
	}

	return nil
}
