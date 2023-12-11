package stats

import (
	"github.com/shopspring/decimal"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type TopSymbolResult struct {
	Symbol string          `json:"symbol"`
	Volume decimal.Decimal `json:"volume"`
	Txs    decimal.Decimal `json:"txs"`
	Tokens []TokenResult   `json:"tokens"`
}

type TokenResult struct {
	EmitterChainID sdk.ChainID     `json:"emitter_chain"`
	TokenChainID   sdk.ChainID     `json:"token_chain"`
	TokenAddress   string          `json:"token_address"`
	Volume         decimal.Decimal `json:"volume"`
	Txs            decimal.Decimal `json:"txs"`
}

type TopSymbolByVolumeResult struct {
	Symbols []*TopSymbolResult `json:"symbols"`
}
