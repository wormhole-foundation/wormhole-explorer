package domain

// manualMainnetTokenList returns a list of tokens that are not generated automatically.
func manualMainnetTokenList() []TokenMetadata {
	return []TokenMetadata{
		{TokenChain: 2, TokenAddress: "000000000000000000000000b0ffa8000886e57f86dd5264b9582b2ad87b2b91", Symbol: "W", CoingeckoID: "wormhole", Decimals: 18},  // Addr: 0xB0fFa8000886e57F86dd5264b9582b2Ad87b2b91
		{TokenChain: 23, TokenAddress: "000000000000000000000000b0ffa8000886e57f86dd5264b9582b2ad87b2b91", Symbol: "W", CoingeckoID: "wormhole", Decimals: 18}, // Addr: 0xB0fFa8000886e57F86dd5264b9582b2Ad87b2b91
		{TokenChain: 24, TokenAddress: "000000000000000000000000b0ffa8000886e57f86dd5264b9582b2ad87b2b91", Symbol: "W", CoingeckoID: "wormhole", Decimals: 18}, // Addr: 0xB0fFa8000886e57F86dd5264b9582b2Ad87b2b91
		{TokenChain: 30, TokenAddress: "000000000000000000000000b0ffa8000886e57f86dd5264b9582b2ad87b2b91", Symbol: "W", CoingeckoID: "wormhole", Decimals: 18}, // Addr: 0xB0fFa8000886e57F86dd5264b9582b2Ad87b2b91

		{TokenChain: 1, TokenAddress: "270ad0028e970df757d5f14f8cbb6a6810e48139125608ea958b718eb2944920", Symbol: "BORG", CoingeckoID: "swissborg", Decimals: 9},  // Addr: 3dQTr7ror2QPKQ3GbBCokJUmjErGg8kTJzdnYjNfvi3Z
		{TokenChain: 2, TokenAddress: "00000000000000000000000064d0f55cd8c7133a9d7102b13987235f486f2224", Symbol: "BORG", CoingeckoID: "swissborg", Decimals: 18}, // Addr: 0x64d0f55cd8c7133a9d7102b13987235f486f2224

		{TokenChain: 10, TokenAddress: "000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", Symbol: "USDC.e", CoingeckoID: "usdc", Decimals: 6}, // Addr: 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48
	}
}

// mainnetTokenList returns a list of all tokens on the mainnet.
func mainnetTokenList() []TokenMetadata {
	return append(generatedMainnetTokenList(), manualMainnetTokenList()...)
}
