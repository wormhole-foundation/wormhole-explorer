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
//
// The map is indexed by "<tokenChain>-<tokenAddress>", which you can find on Token Bridge transfer payloads.
var tokenMetadata = map[string]TokenMetadata{
	// SOL (Portal)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/1/ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5/289384?parsedPayload=true
	"1-0x069b8857feab8184fb687f634618c035dac439dc1aeb3b5598a0f00000000001": {
		UnderlyingSymbol: "SOL",
		Decimals:         9,
	},
	// USDCso - USD Coin (Portal from Solana)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/1/ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5/289386?parsedPayload=true
	"1-0xc6fa7af3bedbad3a3d65f36aabc97431b1bbe4c2d2f6e0e47ca60203452f5d61": {
		UnderlyingSymbol: "USDC",
		Decimals:         6,
	},
	// USDTso - Tether USD (Portal from Solana)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/1/ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5/289373?parsedPayload=true
	"1-0xce010e60afedb22717bd63192f54145a3f965a33bb82d2c7029eb2ce1e208264": {
		UnderlyingSymbol: "USDT",
		Decimals:         6,
	},
	// USDCet - USDCoin (Portal from Ethereum)
	//
	// Examples:
	// * 21/ccceeb29348f71bdd22ffef43a2a19c1f5b5e17c5cca5411529120182672ade5/922
	"2-0x000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48": {
		UnderlyingSymbol: "USDC",
		Decimals:         6,
	},
	// ETH - Ether (Portal)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/1/ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5/288088?parsedPayload=true
	"2-0x000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2": {
		UnderlyingSymbol: "ETH",
		Decimals:         8,
	},
	// USDTet - Tether USD (Portal from Ethereum)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/2/0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585/112361?parsedPayload=true
	"2-0x000000000000000000000000dac17f958d2ee523a2206206994597c13d831ec7": {
		UnderlyingSymbol: "USDT",
		Decimals:         6,
	},
	// UST (Wormhole)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/2/0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585/111492?parsedPayload=true
	"3-0x0100000000000000000000000000000000000000000000000000000075757364": {
		UnderlyingSymbol: "UST",
		Decimals:         8,
	},
	// USDTbs - Tether USD (Portal from BSC)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/4/000000000000000000000000b6f6d86a8f9879a9c87f643768d9efc38c1da6e7/242342?parsedPayload=true
	"4-0x00000000000000000000000055d398326f99059ff775485246999027b3197955": {
		UnderlyingSymbol: "USDT",
		Decimals:         18,
	},
	// USDCbs - USD Coin (Portal from BSC)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/4/000000000000000000000000b6f6d86a8f9879a9c87f643768d9efc38c1da6e7/243491?parsedPayload=true
	"4-0x0000000000000000000000008ac76a51cc950d9822d68b83fe1ad97b32cd580d": {
		UnderlyingSymbol: "USDC",
		Decimals:         18,
	},
	// BNB - Binance Coin (Portal)
	//
	// Examples:
	// * 21/ccceeb29348f71bdd22ffef43a2a19c1f5b5e17c5cca5411529120182672ade5/910
	"4-0x000000000000000000000000bb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c": {
		UnderlyingSymbol: "BNB",
		Decimals:         18,
	},
	// BUSDbs - Binance USD (Portal from BSC)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/4/000000000000000000000000b6f6d86a8f9879a9c87f643768d9efc38c1da6e7/243489?parsedPayload=true
	"4-0x000000000000000000000000e9e7cea3dedca5984780bafc599bd69add087d56": {
		UnderlyingSymbol: "BUSD",
		Decimals:         18,
	},
	// USDCpo -	USD Coin (PoS) (Portal from Polygon)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/1/ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5/289376?parsedPayload=true
	"5-0x0000000000000000000000002791bca1f2de4661ed88a30c99a7a9449aa84174": {
		UnderlyingSymbol: "USDC",
		Decimals:         6,
	},
	// USDTpo - Tether USD (PoS) (Portal from Polygon)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/5/0000000000000000000000005a58505a96d1dbf8df91cb21b54419fc36e93fde/100225?parsedPayload=true
	"5-0x000000000000000000000000c2132d05d31c914a87c6611c10748aeb04b58e8f": {
		UnderlyingSymbol: "USDT",
		Decimals:         6,
	},
	// MATICpo - MATIC (Portal from Polygon)
	//
	// Examples:
	// * 21/ccceeb29348f71bdd22ffef43a2a19c1f5b5e17c5cca5411529120182672ade5/913
	"5-0x0000000000000000000000000d500b1d8e8ef31e21c99d1db9a6444d3adf1270": {
		UnderlyingSymbol: "MATIC",
		Decimals:         18,
	},
	// AVAX (Portal)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/6/0000000000000000000000000e082f06ff657d94310cb8ce8b0d9a04541d8052/94692?parsedPayload=true
	"6-0x000000000000000000000000b31f66aa3c1e785363f0875a1b74e27b85fd66c7": {
		UnderlyingSymbol: "AVAX",
		Decimals:         18,
	},
	// USDCav - USD Coin (Portal from Avalanche)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/6/0000000000000000000000000e082f06ff657d94310cb8ce8b0d9a04541d8052/94690?parsedPayload=true
	"6-0x000000000000000000000000b97ef9ef8734c71904d8002f8b6bc66dd9c48a6e": {
		UnderlyingSymbol: "USDC",
		Decimals:         6,
	},
	// WFTM - Wrapped Fantom
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/10/0000000000000000000000007c9fc5741288cdfdd83ceb07f3ea7e22618d79d2/25144?parsedPayload=true
	"10-0x00000000000000000000000021be370d5312f44cb42ce377bc9b8a0cef1a4c83": {
		UnderlyingSymbol: "FTM",
		Decimals:         8,
	},
}
