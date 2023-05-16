package metric

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	wormscanNotionalCache "github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Metric definition.
type Metric struct {
	influxCli         influxdb2.Client
	apiBucketInfinite api.WriteAPIBlocking
	apiBucket30Days   api.WriteAPIBlocking
	apiBucket24Hours  api.WriteAPIBlocking
	notionalCache     wormscanNotionalCache.NotionalLocalCacheReadable
	logger            *zap.Logger
}

// New create a new *Metric.
func New(
	ctx context.Context,
	influxCli influxdb2.Client,
	organization string,
	bucketInifite string,
	bucket30Days string,
	bucket24Hours string,
	notionalCache wormscanNotionalCache.NotionalLocalCacheReadable,
	logger *zap.Logger,
) (*Metric, error) {

	apiBucketInfinite := influxCli.WriteAPIBlocking(organization, bucketInifite)
	apiBucket30Days := influxCli.WriteAPIBlocking(organization, bucket30Days)
	apiBucket24Hours := influxCli.WriteAPIBlocking(organization, bucket24Hours)

	m := Metric{
		influxCli:         influxCli,
		apiBucketInfinite: apiBucketInfinite,
		apiBucket24Hours:  apiBucket24Hours,
		apiBucket30Days:   apiBucket30Days,
		logger:            logger,
		notionalCache:     notionalCache,
	}
	return &m, nil
}

// Push implement MetricPushFunc definition.
func (m *Metric) Push(ctx context.Context, vaa *sdk.VAA) error {

	err1 := m.vaaCountMeasurement(ctx, vaa)
	err2 := m.volumeMeasurement(ctx, vaa)
	err3 := m.vaaCountAllMessagesMeasurement(ctx, vaa)

	//TODO if we had go 1.20, we could just use `errors.Join(err1, err2, err3)` here.
	if err1 != nil || err2 != nil || err3 != nil {
		return fmt.Errorf("err1=%w, err2=%w, err3=%w", err1, err2, err3)
	}

	return nil
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
	err := m.apiBucket30Days.WritePoint(ctx, point)
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

// vaaCountAllMessagesMeasurement creates a new point for the `vaa_count_all_messages` measurement.
func (m *Metric) vaaCountAllMessagesMeasurement(ctx context.Context, vaa *sdk.VAA) error {

	const measurement = "vaa_count_all_messages"

	// By the way InfluxDB works, two points with the same timesamp will overwrite each other.
	// Hence, we add a random number of nanoseconds to the timestamp to avoid this.
	randomOffset := rand.Int31() % 1000

	// Create a new point
	point := influxdb2.
		NewPointWithMeasurement(measurement).
		AddTag("chain_id", strconv.Itoa(int(vaa.EmitterChain))).
		AddField("count", 1).
		SetTime(vaa.Timestamp.Add(time.Nanosecond * time.Duration(randomOffset)))

	// Write the point to influx
	err := m.apiBucket24Hours.WritePoint(ctx, point)
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
		AddTag("emitter_chain", fmt.Sprintf("%d", vaa.EmitterChain)).
		// Receiver chain
		AddTag("destination_chain", fmt.Sprintf("%d", payload.TargetChain)).
		// Original mint address
		AddTag("token_address", payload.OriginAddress.String()).
		// Original mint chain
		AddTag("token_chain", fmt.Sprintf("%d", payload.OriginChain)).
		// Amount of tokens transferred, integer, 8 decimals of precision
		AddField("amount", amount.Uint64()).
		// Token price at the time the VAA was processed, integer, 8 decimals of precision
		//
		// TODO: We should use the price at the time the VAA was emitted instead.
		AddField("notional", notionalBigInt.Uint64()).
		// Volume in USD, integer, 8 decimals of precision
		AddField("volume", volume.Uint64()).
		SetTime(vaa.Timestamp)

	// Write the point to influx
	err = m.apiBucketInfinite.WritePoint(ctx, point)
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
