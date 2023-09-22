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
	TokenChain   sdk.ChainID
	TokenAddress string
	// Symbol is the name that crypto exchanges use to list the underlying asset represented by this token.
	// For example, the underlying symbol of the token "USDCso (USDC minted on Solana)" is "USDC".
	Symbol      Symbol
	CoingeckoID string
	Decimals    int64
}

var (
	tokenMetadata              = generatedMainnetTokenList()
	tokenMetadataByContractID  = make(map[string]*TokenMetadata)
	tokenMetadataByCoingeckoID = make(map[string]*TokenMetadata)
)

func (t *TokenMetadata) GetTokenID() string {
	return fmt.Sprintf("%d/%s", t.TokenChain, t.TokenAddress)
}

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

// GetAllTokens returns a list of all tokens that exist in the database.
//
// The caller must not modify the `[]TokenMetadata` returned.
func GetAllTokens() []TokenMetadata {
	return tokenMetadata
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
