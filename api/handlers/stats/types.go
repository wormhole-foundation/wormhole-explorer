package stats

import (
	"fmt"

	"github.com/shopspring/decimal"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// SymbolWithAssetsTimeSpan is used as an input parameter for the functions `GetTopAssets` and `GetTopChainPairs`.
type SymbolWithAssetsTimeSpan string
type TopCorridorsTimeSpan string

const (
	TimeSpan7Days  SymbolWithAssetsTimeSpan = "7d"
	TimeSpan15Days SymbolWithAssetsTimeSpan = "15d"
	TimeSpan30Days SymbolWithAssetsTimeSpan = "30d"

	TimeSpan2DaysTopCorridors TopCorridorsTimeSpan = "2d"
	TimeSpan7DaysTopCorridors TopCorridorsTimeSpan = "7d"
)

// ParseSymbolsWithAssetsTimeSpan parses a string and returns a `SymbolsWithAssetsTimeSpan`.
func ParseSymbolsWithAssetsTimeSpan(s string) (*SymbolWithAssetsTimeSpan, error) {

	if s == string(TimeSpan7Days) ||
		s == string(TimeSpan15Days) ||
		s == string(TimeSpan30Days) {

		tmp := SymbolWithAssetsTimeSpan(s)
		return &tmp, nil
	}

	return nil, fmt.Errorf("invalid time span: %s", s)
}

type SymbolWithAssetDTO struct {
	Symbol         string
	EmitterChainID sdk.ChainID
	TokenChainID   sdk.ChainID
	TokenAddress   string
	Volume         decimal.Decimal
	Txs            decimal.Decimal
}

func ParseTopCorridorsTimeSpan(s string) (*TopCorridorsTimeSpan, error) {
	if s == string(TimeSpan2DaysTopCorridors) ||
		s == string(TimeSpan7DaysTopCorridors) {

		tmp := TopCorridorsTimeSpan(s)
		return &tmp, nil
	}

	return nil, fmt.Errorf("invalid time span: %s", s)
}

type TopCorridorsDTO struct {
	EmitterChainID     sdk.ChainID
	DestinationChainID sdk.ChainID
	TokenChainID       sdk.ChainID
	TokenAddress       string
	Txs                uint64
}
