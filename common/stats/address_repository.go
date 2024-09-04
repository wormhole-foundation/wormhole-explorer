package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"go.uber.org/zap"
)

const nttTopAddress = "wormscan:ntt-top-address"

type AddressRepository struct {
	influxCli               influxdb2.Client
	queryAPI                api.QueryAPI
	bucketInfiniteRetention string
	cacheClient             cache.Cache
	logger                  *zap.Logger
}

// NewAddressRepository creates a new instance of AddressRepository
func NewAddressRepository(influxCli influxdb2.Client, org string, bucketInfiniteRetention string,
	cache cache.Cache, logger *zap.Logger) *AddressRepository {
	return &AddressRepository{
		influxCli:               influxCli,
		queryAPI:                influxCli.QueryAPI(org),
		bucketInfiniteRetention: bucketInfiniteRetention,
		cacheClient:             cache,
		logger:                  logger,
	}
}

func (r *AddressRepository) LoadNativeTokenTransferTopAddress(ctx context.Context, symbol string, isNotional bool, expiration time.Duration) error {
	result, err := r.getNativeTokenTransferTopAddress(ctx, symbol, isNotional)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%s:%s:%t", nttTopAddress, symbol, isNotional)
	cr := cachedResult[[]NativeTokenTransferTopAddress]{Timestamp: time.Now(), Result: result}
	return r.cacheClient.Set(ctx, key, cr, expiration)
}

func (r *AddressRepository) GetNativeTokenTransferTopAddress(ctx context.Context, symbol string, isNotional bool) ([]NativeTokenTransferTopAddress, error) {
	key := fmt.Sprintf("%s:%s:%t", nttTopAddress, symbol, isNotional)
	result, err := r.cacheClient.Get(ctx, key)
	if err != nil {
		return r.getNativeTokenTransferTopAddress(ctx, symbol, isNotional)
	}
	var cached cachedResult[[]NativeTokenTransferTopAddress]
	err = json.Unmarshal([]byte(result), &cached)
	if err != nil {
		return nil, err
	}
	return cached.Result, nil
}

func (r *AddressRepository) getNativeTokenTransferTopAddress(ctx context.Context, symbol string, isNotional bool) ([]NativeTokenTransferTopAddress, error) {
	query := buildNTTTopAddress(r.bucketInfiniteRetention, symbol, isNotional, time.Now())
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		r.logger.Error("failed to query ntt top address", zap.Error(err))
		return nil, err
	}
	if result.Err() != nil {
		r.logger.Error("failed to query ntt top address has errors", zap.Error(err))
		return nil, result.Err()
	}

	type Row struct {
		FromAddress string `mapstructure:"from_address"`
		Value       uint64 `mapstructure:"_value"`
	}

	var rows []Row
	for result.Next() {
		var row Row
		if err := mapstructure.Decode(result.Record().Values(), &row); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

	var values []NativeTokenTransferTopAddress
	for _, row := range rows {

		rowValue := decimal.NewFromUint64(row.Value)

		if isNotional {
			rowValue = rowValue.Div(decimal.NewFromInt(1_0000_0000))
		}

		value := NativeTokenTransferTopAddress{
			FromAddress: row.FromAddress,
			Value:       rowValue,
		}
		values = append(values, value)
	}

	return values, nil
}
