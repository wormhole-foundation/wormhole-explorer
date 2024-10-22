package stats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"go.uber.org/zap"
)

const nttMedian = "wormscan:ntt-median"

type NTTRepository struct {
	influxCli               influxdb2.Client
	queryAPI                api.QueryAPI
	bucketInfiniteRetention string
	cacheClient             cache.Cache
	logger                  *zap.Logger
}

func NewNTTRepository(influxCli influxdb2.Client, org string, bucketInfiniteRetention string,
	cache cache.Cache, logger *zap.Logger) *NTTRepository {
	return &NTTRepository{
		influxCli:               influxCli,
		queryAPI:                influxCli.QueryAPI(org),
		bucketInfiniteRetention: bucketInfiniteRetention,
		cacheClient:             cache,
		logger:                  logger,
	}
}

func (r *NTTRepository) LoadNativeTokenTransferMedian(ctx context.Context, symbol string, expiration time.Duration) error {
	result, err := r.getNativeTokenTransferMedian(ctx, symbol)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s:%s", nttMedian, symbol)
	cr := cachedResult[decimal.Decimal]{Timestamp: time.Now(), Result: result}
	return r.cacheClient.Set(ctx, key, cr, expiration)
}

func (r *NTTRepository) GetNativeTokenTransferMedian(ctx context.Context, symbol string) (decimal.Decimal, error) {
	key := fmt.Sprintf("%s:%s", nttMedian, symbol)
	result, err := r.cacheClient.Get(ctx, key)
	if err != nil {
		return r.getNativeTokenTransferMedian(ctx, symbol)
	}
	var cached cachedResult[decimal.Decimal]
	err = json.Unmarshal([]byte(result), &cached)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return cached.Result, nil
}

func (r *NTTRepository) GetNativeTokenTransferTokens(ctx context.Context) ([]string, error) {
	queryTemplate := `
	import "influxdata/influxdb/schema"

	schema.measurementTagValues(
    	bucket: "%s",
    	measurement: "ntt_symbol_chain_1d",
    	tag: "symbol")	
	`
	query := fmt.Sprintf(queryTemplate, r.bucketInfiniteRetention)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to query ntt tokens list", zap.Error(err))
		return nil, err
	}
	if result.Err() != nil {
		r.logger.Error("failed to query ntt tokens list", zap.Error(err))
		return nil, result.Err()
	}

	type Row struct {
		Symbol uint64 `mapstructure:"_value"`
	}

	var rows []Row
	for result.Next() {
		var row Row
		if err := mapstructure.Decode(result.Record().Values(), &row); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

}

func (r *NTTRepository) getNativeTokenTransferMedian(ctx context.Context, symbol string) (decimal.Decimal, error) {
	query := buildNTTMedianTransferSize(r.bucketInfiniteRetention, symbol)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to query ntt median transfer size",
			zap.String("symbol", symbol), zap.Error(err))
		return decimal.Decimal{}, err
	}
	if result.Err() != nil {
		r.logger.Error("failed to query ntt median transfer size has errors",
			zap.String("symbol", symbol), zap.Error(err))
		return decimal.Decimal{}, result.Err()
	}
	if !result.Next() {
		r.logger.Error("ntt median transfer size query result has no next",
			zap.String("symbol", symbol))
		return decimal.Decimal{}, errors.New("no result")
	}
	row := struct {
		Value float64 `mapstructure:"_value"`
	}{}
	if err = mapstructure.Decode(result.Record().Values(), &row); err != nil {
		return decimal.Decimal{}, err
	}

	// convert the value to decimal
	value := decimal.NewFromFloat(row.Value)
	return value, nil
}
