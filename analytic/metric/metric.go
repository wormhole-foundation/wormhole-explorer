package metric

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
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
	apiBucket24Hours.EnableBatching()

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

	const flushTimeout = 5 * time.Second

	// wait a bounded amount of time for all buckets to flush
	ctx, cancelFunc := context.WithTimeout(context.Background(), flushTimeout)
	m.apiBucket24Hours.Flush(ctx)
	m.apiBucket30Days.Flush(ctx)
	m.apiBucketInfinite.Flush(ctx)
	cancelFunc()

	m.influxCli.Close()
}

// vaaCountMeasurement creates a new point for the `vaa_count` measurement.
func (m *Metric) vaaCountMeasurement(ctx context.Context, vaa *sdk.VAA) error {

	// Create a new point
	point, err := MakePointForVaaCount(vaa)
	if err != nil {
		return fmt.Errorf("failed to generate data point for vaa count measurement: %w", err)
	}
	if point == nil {
		// Some VAAs don't generate any data points for this metric (i.e.: PythNet)
		return nil
	}

	// Write the point to influx
	err = m.apiBucket30Days.WritePoint(ctx, point)
	if err != nil {
		m.logger.Error("failed to write metric",
			zap.String("measurement", point.Name()),
			zap.Uint16("chain_id", uint16(vaa.EmitterChain)),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// vaaCountAllMessagesMeasurement creates a new point for the `vaa_count_all_messages` measurement.
func (m *Metric) vaaCountAllMessagesMeasurement(ctx context.Context, vaa *sdk.VAA) error {

	// Quite often we get VAAs that are older than 24 hours.
	// We do not want to generate metrics for those, and moreover influxDB
	// returns an error when we try to do so.
	if time.Since(vaa.Timestamp) > time.Hour*24 {
		m.logger.Debug("vaa is older than 24 hours, skipping",
			zap.Time("timestamp", vaa.Timestamp),
			zap.String("vaaId", vaa.UniqueID()),
		)
		return nil
	}

	const measurement = "vaa_count_all_messages"

	// By the way InfluxDB works, two points with the same timesamp will overwrite each other.
	// Most VAA timestamps only have millisecond resolution, so it is possible that two VAAs
	// will have the same timestamp.
	//
	// Hence, we add a deterministic number of nanoseconds to the timestamp to avoid collisions.
	pseudorandomOffset := vaa.Sequence % 1000

	// Create a new point
	point := influxdb2.
		NewPointWithMeasurement(measurement).
		AddTag("chain_id", strconv.Itoa(int(vaa.EmitterChain))).
		AddField("count", 1).
		SetTime(vaa.Timestamp.Add(time.Nanosecond * time.Duration(pseudorandomOffset)))

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

	// Generate a data point for the volume metric
	p := MakePointForVaaVolumeParams{
		Logger: m.logger,
		Vaa:    vaa,
		TokenPriceFunc: func(symbol domain.Symbol, timestamp time.Time) (float64, error) {

			priceData, err := m.notionalCache.Get(symbol)
			if err != nil {
				return 0, err
			}

			return priceData.NotionalUsd, nil
		},
	}
	point, err := MakePointForVaaVolume(&p)
	if err != nil {
		return err
	}
	if point == nil {
		return nil
	}

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

// MakePointForVaaCount generates a data point for the VAA count measurement.
func MakePointForVaaCount(vaa *sdk.VAA) (*write.Point, error) {

	// Do not generate this metric for PythNet VAAs
	if vaa.EmitterChain == sdk.ChainIDPythNet {
		return nil, nil
	}

	const measurement = "vaa_count"

	// Create a new point
	point := influxdb2.
		NewPointWithMeasurement(measurement).
		AddTag("chain_id", strconv.Itoa(int(vaa.EmitterChain))).
		AddField("count", 1).
		SetTime(vaa.Timestamp.Add(time.Nanosecond * time.Duration(vaa.Sequence)))

	return point, nil
}

// MakePointForVaaVolumeParams contains input parameters for the function `MakePointForVaaVolume`
type MakePointForVaaVolumeParams struct {

	// Vaa is the VAA for which we want to compute the volume metric
	Vaa *sdk.VAA

	// TokenPriceFunc returns the price of the given token at the specified timestamp.
	TokenPriceFunc func(symbol domain.Symbol, timestamp time.Time) (float64, error)

	// Logger is an optional parameter, in case the caller wants additional visibility.
	Logger *zap.Logger
}

// MakePointForVaaVolume builds the InfluxDB volume metric for a given VAA
func MakePointForVaaVolume(params *MakePointForVaaVolumeParams) (*write.Point, error) {

	// Do not generate this metric for PythNet VAAs
	if params.Vaa.EmitterChain == sdk.ChainIDPythNet {
		return nil, nil
	}

	const measurement = "vaa_volume"

	// Decode the VAA payload
	payload, err := sdk.DecodeTransferPayloadHdr(params.Vaa.Payload)
	if err != nil {
		return nil, nil
	}

	// Get the token metadata
	//
	// This is complementary data about the token that is not present in the VAA itself.
	tokenMeta, ok := domain.GetTokenByAddress(payload.OriginChain, payload.OriginAddress.String())
	if !ok {
		if params.Logger != nil {
			params.Logger.Debug("found no token metadata for VAA",
				zap.String("vaaId", params.Vaa.MessageID()),
				zap.String("tokenAddress", payload.OriginAddress.String()),
				zap.Uint16("tokenChain", uint16(payload.OriginChain)),
			)
		}
		return nil, nil
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
	notionalUSD, err := params.TokenPriceFunc(tokenMeta.UnderlyingSymbol, params.Vaa.Timestamp)
	if err != nil {
		if params.Logger != nil {
			params.Logger.Warn("failed to obtain notional for this token",
				zap.String("vaaId", params.Vaa.MessageID()),
				zap.String("tokenAddress", payload.OriginAddress.String()),
				zap.Uint16("tokenChain", uint16(payload.OriginChain)),
				zap.Any("tokenMetadata", tokenMeta),
				zap.Error(err),
			)
		}
		return nil, nil
	}

	// Convert the notional value to an integer with an implicit precision of 8 decimals
	notionalBigInt, err := floatToBigInt(notionalUSD)
	if err != nil {
		return nil, fmt.Errorf("failed to convert notional to big integer: %w", err)
	}

	// Calculate the volume, with an implicit precision of 8 decimals
	var volume big.Int
	volume.Mul(amount, notionalBigInt)
	volume.Div(&volume, big.NewInt(1e8))

	if params.Logger != nil {
		params.Logger.Info("Generated data point for volume metric",
			zap.String("vaaId", params.Vaa.MessageID()),
			zap.String("amount", amount.String()),
			zap.String("notional", notionalBigInt.String()),
			zap.String("volume", volume.String()),
			zap.String("underlyingSymbol", tokenMeta.UnderlyingSymbol.String()),
		)
	}

	// Create a data point with volume-related fields
	//
	// We're converting big integers to int64 because influxdb doesn't support bigint/numeric types.
	point := influxdb2.NewPointWithMeasurement(measurement).
		// This is always set to the portal token bridge app ID, but we may have other apps in the future
		AddTag("app_id", domain.AppIdPortalTokenBridge).
		AddTag("emitter_chain", fmt.Sprintf("%d", params.Vaa.EmitterChain)).
		// Receiver chain
		AddTag("destination_chain", fmt.Sprintf("%d", payload.TargetChain)).
		// Original mint address
		AddTag("token_address", payload.OriginAddress.String()).
		// Original mint chain
		AddTag("token_chain", fmt.Sprintf("%d", payload.OriginChain)).
		// Amount of tokens transferred, integer, 8 decimals of precision
		AddField("amount", amount.Uint64()).
		// Token price at the time the VAA was emitted, integer, 8 decimals of precision
		AddField("notional", notionalBigInt.Uint64()).
		// Volume in USD, integer, 8 decimals of precision
		AddField("volume", volume.Uint64()).
		AddField("symbol", tokenMeta.UnderlyingSymbol.String()).
		SetTime(params.Vaa.Timestamp)

	return point, nil
}
