package tokens

import "fmt"

type TokenDictionary struct {
	tokens []TokenConfigEntry
}

func NewTokenDictionary() *TokenDictionary {
	return &TokenDictionary{
		tokens: TokenList(),
	}
}

func (td *TokenDictionary) GetTokenByChainAndAddress(chainID uint16, address string) (*TokenConfigEntry, error) {
	for _, t := range td.tokens {
		if t.Chain == chainID && t.Addr == address {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("token not found")
}
