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
	"github.com/wormhole-foundation/wormhole-explorer/analytics/cmd/token"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/storage"
	wormscanNotionalCache "github.com/wormhole-foundation/wormhole-explorer/common/client/cache/notional"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

const (
	VaaCountMeasurement       = "vaa_count"
	VaaVolumeMeasurement      = "vaa_volume_v2"
	VaaAllMessagesMeasurement = "vaa_count_all_messages"
)

// Metric definition.
type Metric struct {
	// transferPrices contains the notional price for each token bridge transfer.
	pricesRepository         storage.PricesRepository
	influxCli                influxdb2.Client
	apiBucketInfinite        api.WriteAPIBlocking
	apiBucket30Days          api.WriteAPIBlocking
	apiBucket24Hours         api.WriteAPIBlocking
	notionalCache            wormscanNotionalCache.NotionalLocalCacheReadable
	metrics                  metrics.Metrics
	getTransferredTokenByVaa token.GetTransferredTokenByVaa
	tokenProvider            *domain.TokenProvider
	logger                   *zap.Logger
}

// New create a new *Metric.
func New(
	ctx context.Context,
	pricesRepository storage.PricesRepository,
	influxCli influxdb2.Client,
	organization string,
	bucketInifite string,
	bucket30Days string,
	bucket24Hours string,
	notionalCache wormscanNotionalCache.NotionalLocalCacheReadable,
	metrics metrics.Metrics,
	getTransferredTokenByVaa token.GetTransferredTokenByVaa,
	tokenProvider *domain.TokenProvider,
	logger *zap.Logger,
) (*Metric, error) {

	apiBucketInfinite := influxCli.WriteAPIBlocking(organization, bucketInifite)
	apiBucket30Days := influxCli.WriteAPIBlocking(organization, bucket30Days)
	apiBucket24Hours := influxCli.WriteAPIBlocking(organization, bucket24Hours)
	apiBucket24Hours.EnableBatching()

	m := Metric{
		pricesRepository:         pricesRepository,
		influxCli:                influxCli,
		apiBucketInfinite:        apiBucketInfinite,
		apiBucket24Hours:         apiBucket24Hours,
		apiBucket30Days:          apiBucket30Days,
		logger:                   logger,
		notionalCache:            notionalCache,
		metrics:                  metrics,
		getTransferredTokenByVaa: getTransferredTokenByVaa,
		tokenProvider:            tokenProvider,
	}
	return &m, nil
}

// Push implement MetricPushFunc definition.
func (m *Metric) Push(ctx context.Context, params *Params) error {

	var err1, err2, err3, err4 error

	isVaaSigned := params.VaaIsSigned

	if isVaaSigned {
		err1 = m.vaaCountMeasurement(ctx, params)

		err2 = m.vaaCountAllMessagesMeasurement(ctx, params)
	}

	if params.Vaa.EmitterChain != sdk.ChainIDPythNet {

		transferredToken, err := m.getTransferredTokenByVaa(ctx, params.Vaa)
		if err != nil {
			if !token.IsUnknownTokenErr(err) {
				m.logger.Error("Failed to obtain transferred token for this VAA",
					zap.String("trackId", params.TrackID),
					zap.String("vaaId", params.Vaa.MessageID()),
					zap.Error(err))
				return err
			}
		}

		if transferredToken != nil {

			if isVaaSigned {
				err3 = m.volumeMeasurement(ctx, params, transferredToken.Clone())
			}

			err4 = UpsertTransferPrices(
				ctx,
				m.logger,
				params.Vaa,
				m.pricesRepository,
				func(tokenID, _ string, timestamp time.Time) (decimal.Decimal, error) {

					priceData, err := m.notionalCache.Get(tokenID)
					if err != nil {
						return decimal.NewFromInt(0), err
					}
					return priceData.NotionalUsd, nil
				},
				transferredToken.Clone(),
				m.tokenProvider,
				params.Source,
				params.TrackID,
			)

		} else {
			m.logger.Warn("Cannot obtain transferred token for this VAA",
				zap.Error(err),
				zap.String("trackId", params.TrackID),
				zap.String("vaaId", params.Vaa.MessageID()),
			)
		}
	}

	//TODO if we had go 1.20, we could just use `errors.Join(err1, err2, err3, ...)` here.
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return fmt.Errorf("err1=%w, err2=%w, err3=%w err4=%w", err1, err2, err3, err4)
	}

	if params.Vaa.EmitterChain != sdk.ChainIDPythNet {
		m.logger.Info("Transaction processed successfully",
			zap.String("trackId", params.TrackID),
			zap.Bool("isVaaSigned", isVaaSigned),
			zap.String("vaaId", params.Vaa.MessageID()))
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
func (m *Metric) vaaCountMeasurement(ctx context.Context, p *Params) error {

	// Create a new point
	point, err := MakePointForVaaCount(p.Vaa)
	if err != nil {
		return fmt.Errorf("failed to generate data point for vaa count measurement: %w", err)
	}
	if point == nil {
		// Some VAAs don't generate any data points for this metric (e.g.: PythNet)
		return nil
	}

	// Ignore vaa older than 30 days
	thirtyDaysBefore := time.Now().AddDate(0, 0, -30)
	if p.Vaa.Timestamp.Before(thirtyDaysBefore) {
		return nil
	}

	// Write the point to influx
	err = m.apiBucket30Days.WritePoint(ctx, point)
	if err != nil {
		m.logger.Error("Failed to write metric",
			zap.String("measurement", point.Name()),
			zap.Uint16("chain_id", uint16(p.Vaa.EmitterChain)),
			zap.Error(err),
		)
		m.metrics.IncFailedMeasurement(VaaCountMeasurement)
		return err
	}
	m.metrics.IncSuccessfulMeasurement(VaaCountMeasurement)

	return nil
}

// vaaCountAllMessagesMeasurement creates a new point for the `vaa_count_all_messages` measurement.
func (m *Metric) vaaCountAllMessagesMeasurement(ctx context.Context, params *Params) error {

	// Quite often we get VAAs that are older than 24 hours.
	// We do not want to generate metrics for those, and moreover influxDB
	// returns an error when we try to do so.
	if time.Since(params.Vaa.Timestamp) > time.Hour*24 {
		m.logger.Debug("vaa is older than 24 hours, skipping",
			zap.String("trackId", params.TrackID),
			zap.Time("timestamp", params.Vaa.Timestamp),
			zap.String("vaaId", params.Vaa.UniqueID()),
		)
		return nil
	}

	// Create a new point
	point := influxdb2.
		NewPointWithMeasurement(VaaAllMessagesMeasurement).
		AddTag("chain_id", strconv.Itoa(int(params.Vaa.EmitterChain))).
		AddField("count", 1).
		SetTime(generateUniqueTimestamp(params.Vaa))

	// Write the point to influx
	err := m.apiBucket24Hours.WritePoint(ctx, point)
	if err != nil {
		m.logger.Error("Failed to write metric",
			zap.String("measurement", VaaAllMessagesMeasurement),
			zap.Uint16("chain_id", uint16(params.Vaa.EmitterChain)),
			zap.Error(err),
		)
		m.metrics.IncFailedMeasurement(VaaAllMessagesMeasurement)
		return err
	}
	m.metrics.IncSuccessfulMeasurement(VaaAllMessagesMeasurement)

	return nil
}

// volumeMeasurement creates a new point for the `vaa_volume_v2` measurement.
func (m *Metric) volumeMeasurement(ctx context.Context, params *Params, token *token.TransferredToken) error {

	// Generate a data point for the volume metric
	p := MakePointForVaaVolumeParams{
		Logger: m.logger,
		Vaa:    params.Vaa,
		TokenPriceFunc: func(tokenID string, timestamp time.Time) (decimal.Decimal, error) {

			priceData, err := m.notionalCache.Get(tokenID)
			if err != nil {
				return decimal.NewFromInt(0), err
			}

			return priceData.NotionalUsd, nil
		},
		Metrics:          m.metrics,
		TransferredToken: token,
		TokenProvider:    m.tokenProvider,
	}
	point, err := MakePointForVaaVolume(&p)
	if err != nil {
		return err
	}
	if point == nil {
		// Some VAAs don't generate any data points for this metric (e.g.: PythNet, non-token-bridge VAAs)
		return nil
	}

	vaaVolumeV3point := m.MakePointVaaVolumeV3(point, params, token)

	// Write the point to influx
	err = m.apiBucketInfinite.WritePoint(ctx, point, vaaVolumeV3point)
	if err != nil {
		m.metrics.IncFailedMeasurement(VaaVolumeMeasurement)
		return err
	}
	m.logger.Debug("Wrote a data point for the volume metric",
		zap.String("vaaId", params.Vaa.MessageID()),
		zap.String("trackId", params.TrackID),
		zap.String("measurement", point.Name()),
		zap.Any("tags", point.TagList()),
		zap.Any("fields", point.FieldList()),
	)
	m.metrics.IncSuccessfulMeasurement(VaaVolumeMeasurement)

	return nil
}

func (m *Metric) MakePointVaaVolumeV3(vaaVolumeV2Point *write.Point, params *Params, transferredToken *token.TransferredToken) *write.Point {

	point := influxdb2.NewPointWithMeasurement("vaa_volume_v3")

	point.SetTime(vaaVolumeV2Point.Time())

	for _, field := range vaaVolumeV2Point.FieldList() {
		point.AddField(field.Key, field.Value)
	}

	for _, tag := range vaaVolumeV2Point.TagList() {
		if tag.Key != "app_id" {
			point.AddTag(tag.Key, tag.Value)
		}
	}

	point.AddTag("version", "v5")

	for i, appID := range transferredToken.AppIDs {
		point.AddTag(fmt.Sprintf("app_id_%d", i+1), appID)
	}

	// fill with none app_id_2/3 depending on the number of appIDs to ensure that all data points contain the 3 tags.
	for i := len(transferredToken.AppIDs); i < 3; i++ {
		point.AddTag(fmt.Sprintf("app_id_%d", i+1), "none")
	}

	point.AddTag("size", strconv.Itoa(len(transferredToken.AppIDs)))

	if transferredToken.FromAddress != "" {
		var fromAddr string
		fromAddrHex, err := domain.DecodeNativeAddressToHex(transferredToken.FromChain, transferredToken.FromAddress)
		if err != nil {
			m.logger.Error("Failed to decode native fromAddress to hex", zap.String("trackId", params.TrackID), zap.String("vaaId", params.Vaa.MessageID()), zap.String("nativeFromAddress", transferredToken.FromAddress), zap.Uint16("fromChain", uint16(transferredToken.FromChain)))
			fromAddr = transferredToken.FromAddress
		} else {
			fromAddr = fromAddrHex
		}
		point.AddField("from_address", fromAddr)
	}

	if transferredToken.ToAddress != "" {
		var toAddr string
		toAddrHex, err := domain.DecodeNativeAddressToHex(transferredToken.ToChain, transferredToken.ToAddress)
		if err != nil {
			m.logger.Error("Failed to decode native toAddress to hex", zap.String("trackId", params.TrackID), zap.String("vaaId", params.Vaa.MessageID()), zap.String("nativeToAddress", transferredToken.ToAddress), zap.Uint16("toChain", uint16(transferredToken.ToChain)))
			toAddr = transferredToken.ToAddress
		} else {
			toAddr = toAddrHex
		}
		point.AddField("to_address", toAddr)
	}

	if len(transferredToken.AppIDs) > 3 {
		m.logger.Warn("Too many appIDs.",
			zap.String("vaaId", params.Vaa.MessageID()),
			zap.String("trackId", params.TrackID),
			zap.String("appIDs", fmt.Sprintf("%v", transferredToken.AppIDs)))
	}

	return point
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
		NewPointWithMeasurement(VaaCountMeasurement).
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
	TokenPriceFunc func(tokenID string, timestamp time.Time) (decimal.Decimal, error)

	// Logger is an optional parameter, in case the caller wants additional visibility.
	Logger *zap.Logger

	// Metrics is in case the caller wants additional visibility.
	Metrics metrics.Metrics

	// TransferredToken is the token that was transferred in the VAA.
	TransferredToken *token.TransferredToken

	// TokenProvider is used to obtain token metadata.
	TokenProvider *domain.TokenProvider
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
			params.Logger.Warn("Emitter chain is unset",
				zap.String("vaaId", params.Vaa.MessageID()),
				zap.Uint16("emitterChain", uint16(params.Vaa.EmitterChain)),
			)
		}
		return nil, nil
	}

	// Do not generate this metric when the TransferredToken is undefined
	if params.TransferredToken == nil {
		if params.Logger != nil {
			params.Logger.Warn("Transferred token is undefined",
				zap.String("vaaId", params.Vaa.MessageID()),
			)
		}
		return nil, nil
	}

	// Create a data point
	point := influxdb2.NewPointWithMeasurement(VaaVolumeMeasurement).
		// This is always set to the portal token bridge app ID, but we may have other apps in the future
		AddTag("app_id", params.TransferredToken.AppId).
		AddTag("emitter_chain", fmt.Sprintf("%d", params.Vaa.EmitterChain)).
		// Receiver chain
		AddTag("destination_chain", fmt.Sprintf("%d", params.TransferredToken.ToChain)).
		// Original mint address
		AddTag("token_address", params.TransferredToken.TokenAddress.String()).
		// Original mint chain
		AddTag("token_chain", fmt.Sprintf("%d", params.TransferredToken.TokenChain)).
		// Measurement version
		AddTag("version", "v2").
		SetTime(params.Vaa.Timestamp)

	// Get the token metadata
	//
	// This is complementary data about the token that is not present in the VAA itself.
	tokenMeta, ok := params.TokenProvider.GetTokenByAddress(params.TransferredToken.TokenChain, params.TransferredToken.TokenAddress.String())
	if !ok {
		params.Metrics.IncMissingToken(params.TransferredToken.TokenChain.String(), params.TransferredToken.TokenAddress.String())
		// We don't have metadata for this token, so we can't compute the volume-related fields
		// (i.e.: amount, notional, volume, symbol, etc.)
		//
		// InfluxDB will reject data points that don't have any fields, so we need to
		// add a dummy field.
		//
		// Moreover, many flux queries depend on the existence of the `volume` field,
		// and would break if we had measurements without it.
		params.Logger.Warn("Cannot obtain this token",
			zap.String("vaaId", params.Vaa.MessageID()),
			zap.String("tokenAddress", params.TransferredToken.TokenAddress.String()),
			zap.Uint16("tokenChain", uint16(params.TransferredToken.TokenChain)),
			zap.Any("tokenMetadata", tokenMeta),
		)
		point.AddField("volume", uint64(0))
		return point, nil
	}
	params.Metrics.IncFoundToken(params.TransferredToken.TokenChain.String(), params.TransferredToken.TokenAddress.String())

	// Normalize the amount to 8 decimals
	amount := params.TransferredToken.Amount
	if tokenMeta.Decimals < 8 {

		// factor = 10 ^ (8 - tokenMeta.Decimals)
		var factor big.Int
		factor.Exp(big.NewInt(10), big.NewInt(int64(8-tokenMeta.Decimals)), nil)

		amount = amount.Mul(amount, &factor)
	}

	// Try to obtain the token notional value from the cache
	notionalUSD, err := params.TokenPriceFunc(tokenMeta.GetTokenID(), params.Vaa.Timestamp)
	if err != nil {
		params.Metrics.IncMissingNotional(tokenMeta.Symbol.String())
		if params.Logger != nil {
			params.Logger.Warn("Failed to obtain notional for this token",
				zap.String("vaaId", params.Vaa.MessageID()),
				zap.String("tokenAddress", params.TransferredToken.TokenAddress.String()),
				zap.Uint16("tokenChain", uint16(params.TransferredToken.TokenChain)),
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
