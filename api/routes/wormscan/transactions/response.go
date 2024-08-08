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

	// Number of VAAs emitted in the last 24 hours (includes Pyth messages).
	Messages24h string `json:"24h_messages"`

	// Number of VAAs emitted since the creation of the network (includes Pyth messages).
	TotalMessages string `json:"total_messages"`

	// Number of VAAs emitted since the creation of the network (does not include Pyth messages)
	TotalTxCount string `json:"total_tx_count"`

	TotalVolume string `json:"total_volume"`

	// Total value locked in USD.
	Tvl string `json:"tvl"`

	// Volume transferred through the token bridge in the last 24 hours, in USD.
	Volume24h string `json:"24h_volume"`
}

// TopAssetsResponse is the "200 OK" response model for `GET /api/v1/top-assets-by-volume`.
type TopAssetsResponse struct {
	Assets []AssetWithVolume `json:"assets"`
}

type AssetWithVolume struct {
	EmitterChain sdk.ChainID `json:"emitterChain"`
	Symbol       string      `json:"symbol,omitempty"`
	TokenChain   sdk.ChainID `json:"tokenChain"`
	TokenAddress string      `json:"tokenAddress"`
	Volume       string      `json:"volume"`
}

// TopChainPairsResponse is the "200 OK" response model for `GET /api/v1/top-chain-pairs-by-num-transfers`.
type TopChainPairsResponse struct {
	ChainPairs []ChainPair `json:"chainPairs"`
}

type ChainPair struct {
	EmitterChain      sdk.ChainID `json:"emitterChain"`
	DestinationChain  sdk.ChainID `json:"destinationChain"`
	NumberOfTransfers string      `json:"numberOfTransfers"`
}
