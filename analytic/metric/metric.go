package metric

import (
	"context"
	"fmt"
	"math"
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
	ctx context.Context,
	influxCli influxdb2.Client,
	cfg *config.Configuration,
	logger *zap.Logger,
) (*Metric, error) {

	writeAPI := influxCli.WriteAPIBlocking(cfg.InfluxOrganization, cfg.InfluxBucket)

	_, notionalCache, err := newCache(ctx, cfg, logger)
	if err != nil {
		return nil, err
	}

	m := Metric{
		influxCli:     influxCli,
		writeApi:      writeAPI,
		logger:        logger,
		notionalCache: notionalCache,
	}
	return &m, nil
}

func newCache(
	ctx context.Context,
	cfg *config.Configuration,
	logger *zap.Logger,
) (wormscanCache.CacheReadable, wormscanNotionalCache.NotionalLocalCacheReadable, error) {

	// use a distributed cache and for notional a pubsub to sync local cache.
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.CacheURL})

	// get cache client
	cacheClient, err := wormscanCache.NewCacheClient(redisClient, true /*enabled*/, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cache client: %w", err)
	}

	// get notional cache client and init load to local cache
	notionalCache, err := wormscanNotionalCache.NewNotionalCache(ctx, redisClient, cfg.CacheChannel, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create notional cache client: %w", err)
	}
	notionalCache.Init(ctx)

	return cacheClient, notionalCache, nil
}

// Push implement MetricPushFunc definition.
func (m *Metric) Push(ctx context.Context, vaa *sdk.VAA) error {

	err1 := m.vaaCountMeasurement(ctx, vaa)
	err2 := m.volumeMeasurement(ctx, vaa)

	//TODO if we had go 1.20, we could just use `errors.Join(err1, err2)` here.
	return fmt.Errorf("err1=%w, err2=%w", err1, err2)
}

// Close influx client.
func (m *Metric) Close() {
	m.influxCli.Close()
}

// vaaCountMeasurement creates a new point for the `vaa_count` measurement.
func (m *Metric) vaaCountMeasurement(ctx context.Context, vaa *sdk.VAA) error {

	const measurement = "vaa_count"

	// Create a new point
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

	return nil
}

// volumeMeasurement creates a new point for the `vaa_volume` measurement.
func (m *Metric) volumeMeasurement(ctx context.Context, vaa *sdk.VAA) error {

	const measurement = "vaa_volume"

	// Decode the VAA payload
	//
	// If the VAA didn't come from the portal token bridge, we just skip it.
	payload, err := sdk.DecodeTransferPayloadHdr(vaa.Payload)
	if err != nil {
		return nil
	}

	// Get the token metadata
	//
	// This is complementary data about the token that is not present in the VAA itself.
	tokenMeta, ok := domain.GetTokenByAddress(payload.OriginChain, payload.OriginAddress.String())
	if !ok {
		m.logger.Debug("found no token metadata for VAA",
			zap.String("vaaId", vaa.MessageID()),
			zap.String("tokenAddress", payload.OriginAddress.String()),
			zap.Uint16("tokenChain", uint16(payload.OriginChain)),
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
		m.logger.Warn("failed to obtain notional for this token",
			zap.String("vaaId", vaa.MessageID()),
			zap.String("tokenAddress", payload.OriginAddress.String()),
			zap.Uint16("tokenChain", uint16(payload.OriginChain)),
			zap.Any("tokenMetadata", tokenMeta),
			zap.Error(err),
		)
		return nil
	}

	// Convert the notional value to an integer with an implicit precision of 8 decimals
	notionalBigInt, err := floatToBigInt(notional.NotionalUsd)
	if err != nil {
		return nil
	}

	// Calculate the volume, with an implicit precision of 8 decimals
	var volume big.Int
	volume.Mul(amount, notionalBigInt)
	volume.Div(&volume, big.NewInt(1e8))

	m.logger.Info("Pushing volume metrics",
		zap.String("vaaId", vaa.MessageID()),
		zap.String("amount", amount.String()),
		zap.String("notional", notionalBigInt.String()),
		zap.String("volume", volume.String()),
		zap.String("underlyingSymbol", tokenMeta.UnderlyingSymbol.String()),
	)

	// Create a data point with volume-related fields
	//
	// We're converting big integers to int64 because influxdb doesn't support bigint/numeric types.
	point := influxdb2.NewPointWithMeasurement(measurement).
		// This is always set to the portal token bridge app ID, but we may have other apps in the future
		AddTag("app_id", domain.AppIdPortalTokenBridge).
		// Receiver address
		AddTag("destination_address", payload.TargetAddress.String()).
		// Receiver chain
		AddTag("destination_chain", fmt.Sprintf("%d", payload.TargetChain)).
		// Original mint address
		AddTag("token_address", payload.OriginAddress.String()).
		// Original mint chain
		AddTag("token_chain", fmt.Sprintf("%d", payload.OriginChain)).
		// Amount of tokens transferred, integer, 8 decimals of precision
		AddField("amount", amount.Int64()).
		// Token price at the time the VAA was processed, integer, 8 decimals of precision
		//
		// TODO: We should use the price at the time the VAA was emitted instead.
		AddField("notional", notionalBigInt.Int64()).
		// Volume in USD, integer, 8 decimals of precision
		AddField("volume", volume.Int64()).
		SetTime(vaa.Timestamp)

	// Write the point to influx
	err = m.writeApi.WritePoint(ctx, point)
	if err != nil {
		return err
	}

	return nil
}

// toInt converts a float64 into a big.Int with 8 decimals of implicit precision.
//
// If we ever upgrade the notional cache to store prices as big integers,
// this gnarly function won't be needed anymore.
func floatToBigInt(f float64) (*big.Int, error) {

	integral, frac := math.Modf(f)

	strIntegral := strconv.FormatFloat(integral, 'f', 0, 64)
	strFrac := fmt.Sprintf("%.8f", frac)[2:]

	i, err := strconv.ParseInt(strIntegral+strFrac, 10, 64)
	if err != nil {
		return nil, err
	}

	return big.NewInt(i), nil
}
