package domain

// manualMainnetTokenList returns a list of tokens that are not generated automatically.
func manualMainnetTokenList() []TokenMetadata {
	return []TokenMetadata{
		{TokenChain: 1, TokenAddress: "6927fdc01ea906f96d7137874cdd7adad00ca35764619310e54196c781d84d5b", Symbol: "W", CoingeckoID: "wormhole", Decimals: 6},   // Addr: 85VBFQZC9TZkfaptBWjvUw7YbZjy52A6mjtPGjstQAmQ
		{TokenChain: 2, TokenAddress: "000000000000000000000000b0ffa8000886e57f86dd5264b9582b2ad87b2b91", Symbol: "W", CoingeckoID: "wormhole", Decimals: 18},  // Addr: 0xB0fFa8000886e57F86dd5264b9582b2Ad87b2b91
		{TokenChain: 23, TokenAddress: "000000000000000000000000b0ffa8000886e57f86dd5264b9582b2ad87b2b91", Symbol: "W", CoingeckoID: "wormhole", Decimals: 18}, // Addr: 0xB0fFa8000886e57F86dd5264b9582b2Ad87b2b91
		{TokenChain: 24, TokenAddress: "000000000000000000000000b0ffa8000886e57f86dd5264b9582b2ad87b2b91", Symbol: "W", CoingeckoID: "wormhole", Decimals: 18}, // Addr: 0xB0fFa8000886e57F86dd5264b9582b2Ad87b2b91
		{TokenChain: 30, TokenAddress: "000000000000000000000000b0ffa8000886e57f86dd5264b9582b2ad87b2b91", Symbol: "W", CoingeckoID: "wormhole", Decimals: 18}, // Addr: 0xB0fFa8000886e57F86dd5264b9582b2Ad87b2b91

	}
}

// mainnetTokenList returns a list of all tokens on the mainnet.
func mainnetTokenList() []TokenMetadata {
	return append(generatedMainnetTokenList(), manualMainnetTokenList()...)
}
