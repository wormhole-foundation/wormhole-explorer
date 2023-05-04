package domain

import (
	"fmt"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// TokenMetadata contains information about a token supported by Portal Token Bridge.
type TokenMetadata struct {
	// UnderlyingSymbol is the name that crypto exchanges use to list the underlying asset represented by this token.
	// For example, the underlying symbol of the token "WFTM (wrapped fantom)" is "FTM".
	UnderlyingSymbol string
	Decimals         uint8
}

// GetTokenMetadata returns information about a token identified by the pair (tokenChain, tokenAddr).
func GetTokenMetadata(tokenChain sdk.ChainID, tokenAddr string) (*TokenMetadata, bool) {

	key := fmt.Sprintf("%d-%s", tokenChain, tokenAddr)

	result, ok := tokenMetadata[key]
	if !ok {
		return nil, false
	}

	// The variable `result` is a copy of the value in the map,
	// so we can safely return it without worrying about it being modified.
	return &result, true
}

// tokenMetadata contains information about some of the tokens supported by Portal Token Bridge.
var tokenMetadata = map[string]TokenMetadata{
	// ETH - Ether (Portal)
	//
	// Examples:
	// * https://api.staging.wormscan.io/api/v1/vaas/1/ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5/288088?parsedPayload=true
	"2-0x000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2": {
		UnderlyingSymbol: "ETH",
		Decimals:         8,
	},
	// UST (Wormhole)
	//
	// Examples:
	// * https://api.staging.wormscan.io/api/v1/vaas/2/0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585/111492?parsedPayload=true
	"3-0x0100000000000000000000000000000000000000000000000000000075757364": {
		UnderlyingSymbol: "UST",
		Decimals:         8,
	},
	// Binance-Peg BSC-USD
	//
	// Examples:
	// * https://api.staging.wormscan.io/api/v1/vaas/4/000000000000000000000000b6f6d86a8f9879a9c87f643768d9efc38c1da6e7/242342?parsedPayload=true
	"4-0x00000000000000000000000055d398326f99059ff775485246999027b3197955": {
		UnderlyingSymbol: "BUSD",
		Decimals:         8,
	},
	// WFTM - Wrapped Fantom
	//
	// Examples:
	// * https://api.staging.wormscan.io/api/v1/vaas/10/0000000000000000000000007c9fc5741288cdfdd83ceb07f3ea7e22618d79d2/25144?parsedPayload=true
	"10-0x00000000000000000000000021be370d5312f44cb42ce377bc9b8a0cef1a4c83": {
		UnderlyingSymbol: "FTM",
		Decimals:         8,
	},
}
