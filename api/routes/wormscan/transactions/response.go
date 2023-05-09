package transactions

import (
	"github.com/shopspring/decimal"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type Tx struct {
	Chain        int             `json:"chain"`
	Volume       decimal.Decimal `json:"volume"`
	Percentage   float64         `json:"percentage"`
	Destinations []Destination   `json:"destinations"`
}

type Destination struct {
	Chain      int             `json:"chain"`
	Volume     decimal.Decimal `json:"volume"`
	Percentage float64         `json:"percentage"`
}

// ChainActivity represent a cross chain activity.
type ChainActivity struct {
	Txs []Tx `json:"txs"`
}

// ScorecardsResponse is the response model for the endpoint `GET /api/v1/scorecards`.
type ScorecardsResponse struct {
	//TODO: we don't have the data for these fields yet, uncomment as the data becomes available.

	// Number of VAAs emitted in the last 24 hours (includes Pyth messages).
	//Messages24h  string `json:"24h_messages"`

	// Number of VAAs emitted since the creation of the network (does not include Pyth messages)
	TotalTxCount string `json:"total_tx_count,omitempty"`

	//TotalVolume  string `json:"total_volume"`

	//TVL          string `json:"tvl"`

	// Number of VAAs emitted in the last 24 hours (does not include Pyth messages).
	TxCount24h string `json:"24h_tx_count"`

	// Volume transferred through the token bridge in the last 24 hours, in USD.
	Volume24h string `json:"24h_volume"`
}

// TopAssetsByVolumeResponse is the "200 OK" response model for `GET /api/v1/top-assets-by-volume`.
type TopAssetsByVolumeResponse struct {
	Assets []AssetWithVolume `json:"assets"`
}

type AssetWithVolume struct {
	EmitterChain sdk.ChainID `json:"emitterChain`
	Symbol       string      `json:"symbol"`
	Volume       string      `json:"volume"`
}
