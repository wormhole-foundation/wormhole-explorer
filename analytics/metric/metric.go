package metric

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/internal/metrics"
	wormscanNotionalCache "github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const (
	vaaCountMeasurement       = "vaa_count"
	vaaVolumeMeasurement      = "vaa_volume"
	vaaAllMessagesMeasurement = "vaa_count_all_messages"
)

// Metric definition.
type Metric struct {
	db *mongo.Database
	// transferPrices contains the notional price for each token bridge transfer.
	transferPrices    *mongo.Collection
	influxCli         influxdb2.Client
	apiBucketInfinite api.WriteAPIBlocking
	apiBucket30Days   api.WriteAPIBlocking
	apiBucket24Hours  api.WriteAPIBlocking
	notionalCache     wormscanNotionalCache.NotionalLocalCacheReadable
	metrics           metrics.Metrics
	logger            *zap.Logger
}

// New create a new *Metric.
func New(
	ctx context.Context,
	db *mongo.Database,
	influxCli influxdb2.Client,
	organization string,
	bucketInifite string,
	bucket30Days string,
	bucket24Hours string,
	notionalCache wormscanNotionalCache.NotionalLocalCacheReadable,
	metrics metrics.Metrics,
	logger *zap.Logger,
) (*Metric, error) {

	apiBucketInfinite := influxCli.WriteAPIBlocking(organization, bucketInifite)
	apiBucket30Days := influxCli.WriteAPIBlocking(organization, bucket30Days)
	apiBucket24Hours := influxCli.WriteAPIBlocking(organization, bucket24Hours)
	apiBucket24Hours.EnableBatching()

	m := Metric{
		db:                db,
		transferPrices:    db.Collection("transferPrices"),
		influxCli:         influxCli,
		apiBucketInfinite: apiBucketInfinite,
		apiBucket24Hours:  apiBucket24Hours,
		apiBucket30Days:   apiBucket30Days,
		logger:            logger,
		notionalCache:     notionalCache,
		metrics:           metrics,
	}
	return &m, nil
}

// Push implement MetricPushFunc definition.
func (m *Metric) Push(ctx context.Context, vaa *sdk.VAA) error {

	err1 := m.vaaCountMeasurement(ctx, vaa)

	err2 := m.volumeMeasurement(ctx, vaa)

	err3 := m.vaaCountAllMessagesMeasurement(ctx, vaa)

	err4 := upsertTransferPrices(
		m.logger,
		vaa,
		m.transferPrices,
		func(symbol domain.Symbol, timestamp time.Time) (decimal.Decimal, error) {

			priceData, err := m.notionalCache.Get(symbol)
			if err != nil {
				return decimal.NewFromInt(0), err
			}

			return priceData.NotionalUsd, nil
		},
	)

	//TODO if we had go 1.20, we could just use `errors.Join(err1, err2, err3, ...)` here.
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return fmt.Errorf("err1=%w, err2=%w, err3=%w err4=%w", err1, err2, err3, err4)
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
		// Some VAAs don't generate any data points for this metric (e.g.: PythNet)
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
		m.metrics.IncFailedMeasurement(vaaCountMeasurement)
		return err
	}
	m.metrics.IncSuccessfulMeasurement(vaaCountMeasurement)

	m.logger.Info("generated a data point for the vaa count metric")

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

	// Create a new point
	point := influxdb2.
		NewPointWithMeasurement(vaaAllMessagesMeasurement).
		AddTag("chain_id", strconv.Itoa(int(vaa.EmitterChain))).
		AddField("count", 1).
		SetTime(generateUniqueTimestamp(vaa))

	// Write the point to influx
	err := m.apiBucket24Hours.WritePoint(ctx, point)
	if err != nil {
		m.logger.Error("failed to write metric",
			zap.String("measurement", vaaAllMessagesMeasurement),
			zap.Uint16("chain_id", uint16(vaa.EmitterChain)),
			zap.Error(err),
		)
		m.metrics.IncFailedMeasurement(vaaAllMessagesMeasurement)
		return err
	}
	m.metrics.IncSuccessfulMeasurement(vaaAllMessagesMeasurement)

	return nil
}

// volumeMeasurement creates a new point for the `vaa_volume` measurement.
func (m *Metric) volumeMeasurement(ctx context.Context, vaa *sdk.VAA) error {

	// Generate a data point for the volume metric
	p := MakePointForVaaVolumeParams{
		Logger: m.logger,
		Vaa:    vaa,
		TokenPriceFunc: func(symbol domain.Symbol, timestamp time.Time) (decimal.Decimal, error) {

			priceData, err := m.notionalCache.Get(symbol)
			if err != nil {
				return decimal.NewFromInt(0), err
			}

			return priceData.NotionalUsd, nil
		},
		Metrics: m.metrics,
	}
	point, err := MakePointForVaaVolume(&p)
	if err != nil {
		return err
	}
	if point == nil {
		// Some VAAs don't generate any data points for this metric (e.g.: PythNet, non-token-bridge VAAs)
		return nil
	}

	// Write the point to influx
	err = m.apiBucketInfinite.WritePoint(ctx, point)
	if err != nil {
		m.metrics.IncFailedMeasurement(vaaVolumeMeasurement)
		return err
	}
	m.logger.Info("Wrote a data point for the volume metric",
		zap.String("vaaId", vaa.MessageID()),
		zap.String("measurement", point.Name()),
		zap.Any("tags", point.TagList()),
		zap.Any("fields", point.FieldList()),
	)
	m.metrics.IncSuccessfulMeasurement(vaaVolumeMeasurement)

	return nil
}

// MakePointForVaaCount generates a data point for the VAA count measurement.
//
// Some VAAs will not generate a measurement, so the caller must always check
// whether the returned point is nil.
func MakePointForVaaCount(vaa *sdk.VAA) (*write.Point, error) {

	// Do not generate this metric for PythNet VAAs
	if vaa.EmitterChain == sdk.ChainIDPythNet {
		return nil, nil
	}

	// Create a new point
	point := influxdb2.
		NewPointWithMeasurement(vaaCountMeasurement).
		AddTag("chain_id", strconv.Itoa(int(vaa.EmitterChain))).
		AddField("count", 1).
		SetTime(generateUniqueTimestamp(vaa))

	return point, nil
}

// MakePointForVaaVolumeParams contains input parameters for the function `MakePointForVaaVolume`
type MakePointForVaaVolumeParams struct {

	// Vaa is the VAA for which we want to compute the volume metric
	Vaa *sdk.VAA

	// TokenPriceFunc returns the price of the given token at the specified timestamp.
	TokenPriceFunc func(symbol domain.Symbol, timestamp time.Time) (decimal.Decimal, error)

	// Logger is an optional parameter, in case the caller wants additional visibility.
	Logger *zap.Logger

	// Metrics is in case the caller wants additional visibility.
	Metrics metrics.Metrics
}

// MakePointForVaaVolume builds the InfluxDB volume metric for a given VAA
//
// Some VAAs will not generate a measurement, so the caller must always check
// whether the returned point is nil.
func MakePointForVaaVolume(params *MakePointForVaaVolumeParams) (*write.Point, error) {

	// Do not generate this metric for PythNet VAAs
	if params.Vaa.EmitterChain == sdk.ChainIDPythNet {
		return nil, nil
	}

	// Do not generate this metric when the emitter chain is unset
	if params.Vaa.EmitterChain.String() == sdk.ChainIDUnset.String() {
		if params.Logger != nil {
			params.Logger.Warn("emitter chain is unset",
				zap.String("vaaId", params.Vaa.MessageID()),
				zap.Uint16("emitterChain", uint16(params.Vaa.EmitterChain)),
			)
		}
		return nil, nil
	}

	// Decode the VAA payload
	payload, err := sdk.DecodeTransferPayloadHdr(params.Vaa.Payload)
	if err != nil {
		return nil, nil
	}

	// Do not generate this metric when the target chain is unset
	if payload.TargetChain.String() == sdk.ChainIDUnset.String() {
		if params.Logger != nil {
			params.Logger.Warn("target chain is unset",
				zap.String("vaaId", params.Vaa.MessageID()),
				zap.Uint16("targetChain", uint16(payload.TargetChain)),
			)
		}
		return nil, nil
	}

	// Create a data point
	point := influxdb2.NewPointWithMeasurement(vaaVolumeMeasurement).
		// This is always set to the portal token bridge app ID, but we may have other apps in the future
		AddTag("app_id", domain.AppIdPortalTokenBridge).
		AddTag("emitter_chain", fmt.Sprintf("%d", params.Vaa.EmitterChain)).
		// Receiver chain
		AddTag("destination_chain", fmt.Sprintf("%d", payload.TargetChain)).
		// Original mint address
		AddTag("token_address", payload.OriginAddress.String()).
		// Original mint chain
		AddTag("token_chain", fmt.Sprintf("%d", payload.OriginChain)).
		SetTime(params.Vaa.Timestamp)

	// Get the token metadata
	//
	// This is complementary data about the token that is not present in the VAA itself.
	tokenMeta, ok := domain.GetTokenByAddress(payload.OriginChain, payload.OriginAddress.String())
	if !ok {
		params.Metrics.IncMissingToken(payload.OriginChain.String(), payload.OriginAddress.String())
		// We don't have metadata for this token, so we can't compute the volume-related fields
		// (i.e.: amount, notional, volume, symbol, etc.)
		//
		// InfluxDB will reject data points that don't have any fields, so we need to
		// add a dummy field.
		//
		// Moreover, many flux queries depend on the existence of the `volume` field,
		// and would break if we had measurements without it.
		point.AddField("volume", uint64(0))
		return point, nil
	}
	params.Metrics.IncFoundToken(payload.OriginChain.String(), payload.OriginAddress.String())

	// Normalize the amount to 8 decimals
	amount := payload.Amount
	if tokenMeta.Decimals < 8 {

		// factor = 10 ^ (8 - tokenMeta.Decimals)
		var factor big.Int
		factor.Exp(big.NewInt(10), big.NewInt(int64(8-tokenMeta.Decimals)), nil)

		amount = amount.Mul(amount, &factor)
	}

	// Try to obtain the token notional value from the cache
	notionalUSD, err := params.TokenPriceFunc(tokenMeta.Symbol, params.Vaa.Timestamp)
	if err != nil {
		params.Metrics.IncMissingNotional(tokenMeta.Symbol.String())
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
	params.Metrics.IncFoundNotional(tokenMeta.Symbol.String())

	// Convert the notional value to an integer with an implicit precision of 8 decimals
	notionalBigInt := notionalUSD.
		Truncate(8).
		Mul(decimal.NewFromInt(1e8)).
		BigInt()

	// Calculate the volume, with an implicit precision of 8 decimals
	var volume big.Int
	volume.Mul(amount, notionalBigInt)
	volume.Div(&volume, big.NewInt(1e8))

	// Add volume-related fields to the data point.
	//
	// We're converting big integers to int64 because influxdb doesn't support bigint/numeric types.
	point.
		AddField("symbol", tokenMeta.Symbol.String()).
		// Amount of tokens transferred, integer, 8 decimals of precision
		AddField("amount", amount.Uint64()).
		// Token price at the time the VAA was emitted, integer, 8 decimals of precision
		AddField("notional", notionalBigInt.Uint64()).
		// Volume in USD, integer, 8 decimals of precision
		AddField("volume", volume.Uint64()).
		SetTime(generateUniqueTimestamp(params.Vaa))

	return point, nil
}

// generateUniqueTimestamp generates a unique timestamp for each VAA.
//
// Most VAA timestamps only have millisecond resolution, so it is possible that two VAAs
// will have the same timestamp.
// By the way InfluxDB works, two points with the same timesamp will overwrite each other.
//
// Hence, we are forced to generate a deterministic unique timestamp for each VAA.
func generateUniqueTimestamp(vaa *sdk.VAA) time.Time {

	// We're adding 1 a nanosecond offset per sequence.
	// Then, we're taking the modulo of 10^6 to ensure that the offset
	// will always be lower than one millisecond.
	//
	// We could also hash the chain, emitter and seq fields,
	// but the current approach is good enough for the time being.
	offset := time.Duration(vaa.Sequence % 1_000_000)

	return vaa.Timestamp.Add(time.Nanosecond * offset)
}
