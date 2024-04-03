package domain

// manualMainnetTokenList returns a list of tokens that are not generated automatically.
func manualMainnetTokenList() []TokenMetadata {
	return []TokenMetadata{
		{TokenChain: 1, TokenAddress: "6927fdc01ea906f96d7137874cdd7adad00ca35764619310e54196c781d84d5b", Symbol: "W", CoingeckoID: "wormhole", Decimals: 6}, // Addr: 85VBFQZC9TZkfaptBWjvUw7YbZjy52A6mjtPGjstQAmQ
	}
}

// mainnetTokenList returns a list of all tokens on the mainnet.
func mainnetTokenList() []TokenMetadata {
	return append(generatedMainnetTokenList(), manualMainnetTokenList()...)
}
