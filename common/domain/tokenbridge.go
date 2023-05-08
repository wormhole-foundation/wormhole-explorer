package domain

import (
	"fmt"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// Symbol identifies a publicly traded token (i.e. "ETH" for Ethereum, "ALGO" for Algorand, etc.)
type Symbol string

func (s Symbol) String() string {
	return string(s)
}

// TokenMetadata contains information about a token supported by Portal Token Bridge.
type TokenMetadata struct {
	// UnderlyingSymbol is the name that crypto exchanges use to list the underlying asset represented by this token.
	// For example, the underlying symbol of the token "WFTM (wrapped fantom)" is "FTM".
	UnderlyingSymbol Symbol
	Decimals         uint8
	CoingeckoID      string
	TokenChain       sdk.ChainID
	TokenAddress     string
}

var (
	tokenMetadataByContractID  = make(map[string]*TokenMetadata, len(tokenMetadata))
	tokenMetadataByCoingeckoID = make(map[string]*TokenMetadata, len(tokenMetadata))
)

func init() {

	for i := range tokenMetadata {

		// populate the map `tokenMetadataByCoingeckoID`
		coingeckoID := tokenMetadata[i].CoingeckoID
		if coingeckoID != "" {
			tokenMetadataByCoingeckoID[coingeckoID] = &tokenMetadata[i]
		}

		// populate the map `tokenMetadataByContractID`
		contractID := makeContractID(tokenMetadata[i].TokenChain, tokenMetadata[i].TokenAddress)
		if contractID != "" {
			tokenMetadataByContractID[contractID] = &tokenMetadata[i]
		}
	}
}

func makeContractID(tokenChain sdk.ChainID, tokenAddress string) string {
	return fmt.Sprintf("%d-%s", tokenChain, tokenAddress)
}

// GetAllCoingeckoIDs returns a list of all coingecko IDs that exist in the database.
func GetAllCoingeckoIDs() []string {

	// use a map to remove duplicates
	uniqueIDs := make(map[string]bool, len(tokenMetadata))
	for i := range tokenMetadata {
		uniqueIDs[tokenMetadata[i].CoingeckoID] = true
	}

	// collect keys into a slice
	ids := make([]string, 0, len(uniqueIDs))
	for k := range uniqueIDs {
		ids = append(ids, k)
	}

	return ids
}

// GetTokenByCoingeckoID returns information about a token identified by its coingecko ID.
//
// The caller must not modify the `*TokenMetadata` returned.
func GetTokenByCoingeckoID(coingeckoID string) (*TokenMetadata, bool) {

	result, ok := tokenMetadataByCoingeckoID[coingeckoID]
	if !ok {
		return nil, false
	}

	return result, true
}

// GetTokenByAddress returns information about a token identified by its original mint address.
//
// The caller must not modify the `*TokenMetadata` returned.
func GetTokenByAddress(tokenChain sdk.ChainID, tokenAddress string) (*TokenMetadata, bool) {

	key := makeContractID(tokenChain, tokenAddress)

	result, ok := tokenMetadataByContractID[key]
	if !ok {
		return nil, false
	}

	return result, true
}

// tokenMetadata contains information about the most relevant tokens supported by the Token Bridge.
var tokenMetadata = []TokenMetadata{
	// SOL (Portal)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/1/ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5/289384?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDSolana,
		TokenAddress:     "069b8857feab8184fb687f634618c035dac439dc1aeb3b5598a0f00000000001",
		UnderlyingSymbol: "SOL",
		Decimals:         9,
		CoingeckoID:      "solana",
	},
	// DUST Protocol
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/1/ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5/289670?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDSolana,
		TokenAddress:     "b953b5f8dd5457a2a0f0d41903409785b9d84d4045614faa4f505ee132dcd769",
		UnderlyingSymbol: "DUST",
		Decimals:         9,
		CoingeckoID:      "dust-protocol",
	},
	// USDCso - USD Coin (Portal from Solana)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/1/ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5/289386?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDSolana,
		TokenAddress:     "c6fa7af3bedbad3a3d65f36aabc97431b1bbe4c2d2f6e0e47ca60203452f5d61",
		UnderlyingSymbol: "USDC",
		Decimals:         6,
		CoingeckoID:      "usd-coin",
	},
	// USDTso - Tether USD (Portal from Solana)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/1/ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5/289373?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDSolana,
		TokenAddress:     "ce010e60afedb22717bd63192f54145a3f965a33bb82d2c7029eb2ce1e208264",
		UnderlyingSymbol: "USDT",
		Decimals:         6,
		CoingeckoID:      "tether",
	},
	{
		// BRZ - Brazilian Digital
		//
		// Examples:
		// * https://api.wormscan.io/api/v1/vaas/1/ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5/289681?parsedPayload=true
		TokenChain:       sdk.ChainIDSolana,
		TokenAddress:     "dd40a2f6f423e4c3990a83eac3d9d9c1fe625b36cbc5e4a6d553544552a867ee",
		UnderlyingSymbol: "BRZ",
		Decimals:         4,
		CoingeckoID:      "brz",
	},
	// xHashtag
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/6/0000000000000000000000000e082f06ff657d94310cb8ce8b0d9a04541d8052/94898?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDSolana,
		TokenAddress:     "45a5161476cc9df6ef8583b581a3111b4416ebcea65f4eca5bb961124c3399df",
		UnderlyingSymbol: "XTAG",
		Decimals:         6,
		CoingeckoID:      "xhashtag",
	},
	// Zebec Protocol
	//
	// Examples
	// * https://api.wormscan.io/api/v1/vaas/1/ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5/290406?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDSolana,
		TokenAddress:     "0e167d0db0259fb83bca338947ce42fe2c34b803285c7e99b26874bd83bac0a8",
		UnderlyingSymbol: "ZBC",
		Decimals:         8,
		CoingeckoID:      "zebec-protocol",
	},
	// USDCet - USDCoin (Portal from Ethereum)
	//
	// Examples:
	// * 21/ccceeb29348f71bdd22ffef43a2a19c1f5b5e17c5cca5411529120182672ade5/922
	{
		TokenChain:       sdk.ChainIDEthereum,
		TokenAddress:     "000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
		UnderlyingSymbol: "USDC",
		Decimals:         6,
		CoingeckoID:      "usd-coin",
	},
	// ETH - Ether (Portal)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/1/ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5/288088?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDEthereum,
		TokenAddress:     "000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
		UnderlyingSymbol: "ETH",
		Decimals:         8,
		CoingeckoID:      "ethereum",
	},
	// USDTet - Tether USD (Portal from Ethereum)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/2/0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585/112361?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDEthereum,
		TokenAddress:     "000000000000000000000000dac17f958d2ee523a2206206994597c13d831ec7",
		UnderlyingSymbol: "USDT",
		Decimals:         6,
		CoingeckoID:      "tether",
	},
	// XCAD
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/2/0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585/112812?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDEthereum,
		TokenAddress:     "0000000000000000000000007659ce147d0e714454073a5dd7003544234b6aa0",
		UnderlyingSymbol: "XCAD",
		Decimals:         9,
		CoingeckoID:      "xcad-network",
	},
	// ETHIX
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/2/0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585/108181?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDEthereum,
		TokenAddress:     "000000000000000000000000fd09911130e6930bf87f2b0554c44f400bd80d3e",
		UnderlyingSymbol: "ETHIX",
		Decimals:         8,
		CoingeckoID:      "ethichub",
	},
	// SDEX
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/2/0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585/112815?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDEthereum,
		TokenAddress:     "0000000000000000000000005de8ab7e27f6e7a1fff3e5b337584aa43961beef",
		UnderlyingSymbol: "SDEX",
		Decimals:         18,
		CoingeckoID:      "smardex",
	},
	// ORION - Orion Money
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/3/0000000000000000000000007cf7b764e38a0a5e967972c1df77d432510564e2/258681?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDEthereum,
		TokenAddress:     "000000000000000000000000727f064a78dc734d33eec18d5370aef32ffd46e4",
		UnderlyingSymbol: "ORION",
		Decimals:         18,
		CoingeckoID:      "orion-money",
	},
	// SUSHI
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/2/0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585/112824?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDEthereum,
		TokenAddress:     "0000000000000000000000006b3595068778dd592e39a122f4f5a5cf09c90fe2",
		UnderlyingSymbol: "SUSHI",
		Decimals:         18,
		CoingeckoID:      "sushi",
	},
	{
		// LUNC - Terra Luna Classic
		//
		// Examples:
		// * https://api.wormscan.io/api/v1/vaas/4/000000000000000000000000b6f6d86a8f9879a9c87f643768d9efc38c1da6e7/243784?parsedPayload=true
		TokenChain:       sdk.ChainIDTerra,
		TokenAddress:     "010000000000000000000000000000000000000000000000000000756c756e61",
		UnderlyingSymbol: "LUNC",
		CoingeckoID:      "terra-luna",
		Decimals:         6,
	},
	// UST (Wormhole - Solana)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/2/0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585/111492?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDTerra,
		TokenAddress:     "0100000000000000000000000000000000000000000000000000000075757364",
		UnderlyingSymbol: "UST",
		Decimals:         8,
		CoingeckoID:      "terrausd-wormhole",
	},
	// USDTbs - Tether USD (Portal from BSC)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/4/000000000000000000000000b6f6d86a8f9879a9c87f643768d9efc38c1da6e7/242342?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDBSC,
		TokenAddress:     "00000000000000000000000055d398326f99059ff775485246999027b3197955",
		UnderlyingSymbol: "USDT",
		Decimals:         18,
		CoingeckoID:      "tether",
	},
	// USDCbs - USD Coin (Portal from BSC)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/4/000000000000000000000000b6f6d86a8f9879a9c87f643768d9efc38c1da6e7/243491?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDBSC,
		TokenAddress:     "0000000000000000000000008ac76a51cc950d9822d68b83fe1ad97b32cd580d",
		UnderlyingSymbol: "USDC",
		Decimals:         18,
		CoingeckoID:      "usd-coin",
	},
	// BNB - Binance Coin (Portal)
	//
	// Examples:
	// * 21/ccceeb29348f71bdd22ffef43a2a19c1f5b5e17c5cca5411529120182672ade5/910
	{
		TokenChain:       sdk.ChainIDBSC,
		TokenAddress:     "000000000000000000000000bb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c",
		UnderlyingSymbol: "BNB",
		Decimals:         18,
		CoingeckoID:      "binancecoin",
	},
	// WOM - Wombat exchange
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/4/000000000000000000000000b6f6d86a8f9879a9c87f643768d9efc38c1da6e7/243788?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDBSC,
		TokenAddress:     "000000000000000000000000ad6742a35fb341a9cc6ad674738dd8da98b94fb1",
		UnderlyingSymbol: "WOM",
		Decimals:         18,
		CoingeckoID:      "wombat-exchange",
	},
	// BUSDbs - Binance USD (Portal from BSC)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/4/000000000000000000000000b6f6d86a8f9879a9c87f643768d9efc38c1da6e7/243489?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDBSC,
		TokenAddress:     "000000000000000000000000e9e7cea3dedca5984780bafc599bd69add087d56",
		UnderlyingSymbol: "BUSD",
		Decimals:         18,
		CoingeckoID:      "binance-usd",
	},
	// MagicCraft
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/4/000000000000000000000000b6f6d86a8f9879a9c87f643768d9efc38c1da6e7/244486?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDBSC,
		TokenAddress:     "",
		UnderlyingSymbol: "MCRT",
		Decimals:         8,
		CoingeckoID:      "magiccraft",
	},
	// USDCpo -	USD Coin (PoS) (Portal from Polygon)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/1/ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5/289376?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDPolygon,
		TokenAddress:     "0000000000000000000000002791bca1f2de4661ed88a30c99a7a9449aa84174",
		UnderlyingSymbol: "USDC",
		Decimals:         6,
		CoingeckoID:      "usd-coin",
	},
	// USDTpo - Tether USD (PoS) (Portal from Polygon)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/5/0000000000000000000000005a58505a96d1dbf8df91cb21b54419fc36e93fde/100225?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDPolygon,
		TokenAddress:     "000000000000000000000000c2132d05d31c914a87c6611c10748aeb04b58e8f",
		UnderlyingSymbol: "USDT",
		Decimals:         6,
		CoingeckoID:      "tether",
	},
	// MATICpo - MATIC (Portal from Polygon)
	//
	// Examples:
	// * 21/ccceeb29348f71bdd22ffef43a2a19c1f5b5e17c5cca5411529120182672ade5/913
	{
		TokenChain:       sdk.ChainIDPolygon,
		TokenAddress:     "0000000000000000000000000d500b1d8e8ef31e21c99d1db9a6444d3adf1270",
		UnderlyingSymbol: "MATIC",
		Decimals:         18,
		CoingeckoID:      "matic-network",
	},
	// AVAX (Portal)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/6/0000000000000000000000000e082f06ff657d94310cb8ce8b0d9a04541d8052/94692?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDAvalanche,
		TokenAddress:     "000000000000000000000000b31f66aa3c1e785363f0875a1b74e27b85fd66c7",
		UnderlyingSymbol: "AVAX",
		Decimals:         18,
		CoingeckoID:      "avalanche-2",
	},
	// USDCav - USD Coin (Portal from Avalanche)
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/6/0000000000000000000000000e082f06ff657d94310cb8ce8b0d9a04541d8052/94690?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDAvalanche,
		TokenAddress:     "000000000000000000000000b97ef9ef8734c71904d8002f8b6bc66dd9c48a6e",
		UnderlyingSymbol: "USDC",
		Decimals:         6,
		CoingeckoID:      "usd-coin",
	},
	// WFTM - Wrapped Fantom
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/10/0000000000000000000000007c9fc5741288cdfdd83ceb07f3ea7e22618d79d2/25144?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDFantom,
		TokenAddress:     "00000000000000000000000021be370d5312f44cb42ce377bc9b8a0cef1a4c83",
		UnderlyingSymbol: "FTM",
		Decimals:         8,
		CoingeckoID:      "fantom",
	},
	// SUI
	//
	// Examples:
	// * 21/ccceeb29348f71bdd22ffef43a2a19c1f5b5e17c5cca5411529120182672ade5/1247
	{
		TokenChain:       sdk.ChainIDSui,
		TokenAddress:     "9258181f5ceac8dbffb7030890243caed69a9599d2886d957a9cb7656af3bdb3",
		UnderlyingSymbol: "SUI",
		Decimals:         9,
		CoingeckoID:      "sui",
	},
	{
		//TODO find the ContractAddress, decimals and an example VAA for this token.
		TokenChain:       sdk.ChainIDAcala,
		UnderlyingSymbol: "ACA",
		CoingeckoID:      "acala",
	},
	{
		//TODO find the ContractAddress, decimals and an example VAA for this token.
		TokenChain:       sdk.ChainIDAlgorand,
		UnderlyingSymbol: "ALGO",
		CoingeckoID:      "algorand",
	},
	{
		//TODO find the ContractAddress, decimals and an example VAA for this token.
		TokenChain:       sdk.ChainIDAptos,
		UnderlyingSymbol: "APT",
		CoingeckoID:      "aptos",
	},
	// USD Coin
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/5/0000000000000000000000005a58505a96d1dbf8df91cb21b54419fc36e93fde/101667?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDAptos,
		TokenAddress:     "6155e0a106aeb3b0944388613027aee11c84921969ff775727e8046b17b17154",
		UnderlyingSymbol: "USDC",
		Decimals:         6,
		CoingeckoID:      "usd-coin",
	},
	{
		// WGLMR
		//
		// Examples:
		// * https://api.wormscan.io/api/v1/vaas/16/000000000000000000000000b1731c586ca89a23809861c6103f0b96b3f57d92/5897?parsedPayload=true
		TokenChain:       sdk.ChainIDMoonbeam,
		TokenAddress:     "000000000000000000000000acc15dc74880c9944775448304b263d191c6077f",
		UnderlyingSymbol: "WGLMR",
		Decimals:         8,
		CoingeckoID:      "moonbeam",
	},
	// WETH - moonbeam
	//
	// Examples:
	// * https://api.wormscan.io/api/v1/vaas/16/000000000000000000000000b1731c586ca89a23809861c6103f0b96b3f57d92/5898?parsedPayload=true
	{
		TokenChain:       sdk.ChainIDMoonbeam,
		TokenAddress:     "00000000000000000000000030d2a9f5fdf90ace8c17952cbb4ee48a55d916a7",
		UnderlyingSymbol: "WETH",
		Decimals:         8,
		CoingeckoID:      "weth",
	},
	{
		//TODO find missing data for this token
		// aUST
		// ContractAddress: "000000000000000000000000b8ae5604d7858eaa46197b19494b595b586e466c",
	},
	{
		//TODO find missing data for this token
		// WBNB
		// ContractAddress: "000000000000000000000000bb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c",
	},
}
