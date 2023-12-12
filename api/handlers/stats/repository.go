package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type Repository struct {
	influxCli              influxdb2.Client
	queryAPI               api.QueryAPI
	bucket24HoursRetention string
	logger                 *zap.Logger
}

func NewRepository(
	client influxdb2.Client,
	org string,
	bucket24HoursRetention string,
	logger *zap.Logger,
) *Repository {

	r := Repository{
		influxCli:              client,
		queryAPI:               client.QueryAPI(org),
		bucket24HoursRetention: bucket24HoursRetention,
		logger:                 logger,
	}
	return &r
}

func (r *Repository) GetSymbolWithAssets(ctx context.Context, timeSpan SymbolWithAssetsTimeSpan) ([]SymbolWithAssetDTO, error) {
	var measurement string
	switch timeSpan {
	case TimeSpan7Days:
		measurement = "assets_by_symbol_7_days_3h_v2"
	case TimeSpan15Days:
		measurement = "assets_by_symbol_15_days_3h_v2"
	case TimeSpan30Days:
		measurement = "assets_by_symbol_30_days_3h_v2"
	default:
		measurement = "assets_by_symbol_7_days_3h_v2"
	}

	query := buildSymbolWithAssets(r.bucket24HoursRetention, time.Now(), measurement)
	result, err := r.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if result.Err() != nil {
		return nil, result.Err()
	}

	// Scan query results
	type Row struct {
		Symbol       string `mapstructure:"symbol"`
		EmitterChain string `mapstructure:"emitter_chain"`
		TokenChain   string `mapstructure:"token_chain"`
		TokenAddress string `mapstructure:"token_address"`
		JsonValue    string `mapstructure:"_value"`
	}

	type TxsVolume struct {
		Txs    decimal.Decimal
		Volume decimal.Decimal
	}

	var rows []Row
	for result.Next() {
		var row Row
		if err := mapstructure.Decode(result.Record().Values(), &row); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

	divisor := decimal.NewFromInt(1_0000_0000)

	// Convert the rows into the response model
	var values []SymbolWithAssetDTO
	for _, row := range rows {

		// parse emitter chain
		emitterChain, err := strconv.ParseUint(row.EmitterChain, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("failed to convert emitter chain field to uint16. %v", err)
		}

		// parse token chain
		tokenChain, err := strconv.ParseUint(row.TokenChain, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("failed to convert token chain field to uint16. %v", err)
		}

		// parse the json value
		var txsVolume TxsVolume
		if err := json.Unmarshal([]byte(row.JsonValue), &txsVolume); err != nil {
			return nil, fmt.Errorf("failed to convert _value to struct. %v", err)
		}

		// append the new item to the response
		value := SymbolWithAssetDTO{
			Symbol:         row.Symbol,
			EmitterChainID: sdk.ChainID(emitterChain),
			TokenChainID:   sdk.ChainID(tokenChain),
			TokenAddress:   row.TokenAddress,
			Volume:         txsVolume.Volume.Div(divisor),
			Txs:            txsVolume.Txs,
		}

		// do not include invalid chain IDs in the response
		if !domain.ChainIdIsValid(value.EmitterChainID) {
			continue
		}

		values = append(values, value)
	}

	return values, nil
}
