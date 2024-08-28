package domain

import (
	"fmt"
	"strings"

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

type TokenProvider struct {
	p2pNetwork                 string
	tokenMetadata              []TokenMetadata
	tokenMetadataByContractID  map[string]*TokenMetadata
	tokenMetadataByCoingeckoID map[string]*TokenMetadata
	coingeckIdBySymbol         map[string]string
	tokenMetadataBySymbol      map[string][]*TokenMetadata
}

func (t *TokenMetadata) GetTokenID() string {
	return fmt.Sprintf("%d/%s", t.TokenChain, t.TokenAddress)
}

func makeContractID(tokenChain sdk.ChainID, tokenAddress string) string {
	return fmt.Sprintf("%d-%s", tokenChain, tokenAddress)
}

func NewTokenProvider(p2pNetwork string) *TokenProvider {
	var tokenMetadata []TokenMetadata

	switch p2pNetwork {
	case P2pMainNet:
		tokenMetadata = mainnetTokenList()
	case P2pTestNet:
		tokenMetadata = manualTestnetTokenList()
	default:
		panic(fmt.Sprintf("unknown p2p network: %s", p2pNetwork))
	}

	tokenMetadataByContractID := make(map[string]*TokenMetadata)
	tokenMetadataByCoingeckoID := make(map[string]*TokenMetadata)
	coingeckoIDBySymbol := make(map[string]string)
	tokenMetadataBySymbol := make(map[string][]*TokenMetadata)

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

		// populete the map `coingeckoIDBySymbol`.
		symbol := strings.ToUpper(tokenMetadata[i].Symbol.String())
		coingeckoIDBySymbol[symbol] = tokenMetadata[i].CoingeckoID
		tokenMetadataBySymbol[symbol] = append(tokenMetadataBySymbol[symbol], &tokenMetadata[i])
	}
	return &TokenProvider{
		p2pNetwork:                 p2pNetwork,
		tokenMetadata:              tokenMetadata,
		tokenMetadataByContractID:  tokenMetadataByContractID,
		tokenMetadataByCoingeckoID: tokenMetadataByCoingeckoID,
		tokenMetadataBySymbol:      tokenMetadataBySymbol,
		coingeckIdBySymbol:         coingeckoIDBySymbol,
	}
}

// GetAllTokens returns a list of all tokens that exist in the database.
//
// The caller must not modify the `[]TokenMetadata` returned.
func (t *TokenProvider) GetAllTokens() []TokenMetadata {
	return t.tokenMetadata
}

// GetAllCoingeckoIDs returns a list of all coingecko IDs that exist in the database.
func (t *TokenProvider) GetAllCoingeckoIDs() []string {

	// use a map to remove duplicates
	uniqueIDs := make(map[string]bool, len(t.tokenMetadata))
	for i := range t.tokenMetadata {
		uniqueIDs[t.tokenMetadata[i].CoingeckoID] = true
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
func (t *TokenProvider) GetTokenByCoingeckoID(coingeckoID string) (*TokenMetadata, bool) {

	result, ok := t.tokenMetadataByCoingeckoID[coingeckoID]
	if !ok {
		return nil, false
	}

	return result, true
}

// GetTokenByAddress returns information about a token identified by its original mint address.
//
// The caller must not modify the `*TokenMetadata` returned.
func (t *TokenProvider) GetTokenByAddress(tokenChain sdk.ChainID, tokenAddress string) (*TokenMetadata, bool) {

	key := makeContractID(tokenChain, tokenAddress)

	result, ok := t.tokenMetadataByContractID[key]
	if !ok {
		return nil, false
	}

	return result, true
}

func (t *TokenProvider) GetCoingeckoIDBySymbol(symbol string) string {
	return t.coingeckIdBySymbol[symbol]
}

func (t *TokenProvider) GetP2pNewtork() string {
	return t.p2pNetwork
}

func (t *TokenProvider) GetTokensBySymbol(symbol string) ([]*TokenMetadata, bool) {
	symbol = strings.ToUpper(symbol)
	tokens, ok := t.tokenMetadataBySymbol[symbol]
	if !ok {
		return nil, false
	}
	return tokens, true
}
