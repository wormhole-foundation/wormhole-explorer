package domain

func manualTestnetTokenList() []TokenMetadata {
	return []TokenMetadata{
		{TokenChain: 1, TokenAddress: "069b8857feab8184fb687f634618c035dac439dc1aeb3b5598a0f00000000001", Symbol: "SOL", CoingeckoID: "wrapped-solana", Decimals: 9},
		{TokenChain: 2, TokenAddress: "000000000000000000000000b4fbf271143f4fbf7b91a5ded31805e42b2208d6", Symbol: "WETH", CoingeckoID: "weth", Decimals: 18},
		{TokenChain: 2, TokenAddress: "00000000000000000000000011fe4b6ae13d2a6055c8d9cf65c55bac32b5d844", Symbol: "DAI", CoingeckoID: "dai", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000ae13d989dac2f0debff460ac112a837c89baa7cd", Symbol: "WBNB", CoingeckoID: "wbnb", Decimals: 18},
		{TokenChain: 5, TokenAddress: "0000000000000000000000009c3c9283d3e44854697cd22d3faa240cfb032889", Symbol: "WMATIC", CoingeckoID: "wmatic", Decimals: 18},
		{TokenChain: 6, TokenAddress: "0000000000000000000000005425890298aed601595a70ab815c96711a31bc65", Symbol: "USDC", CoingeckoID: "usd-coin", Decimals: 6},
		{TokenChain: 6, TokenAddress: "000000000000000000000000d00ae08403b9bbb9124bb305c09058e32c39a48c", Symbol: "WAVAX", CoingeckoID: "wrapped-avax", Decimals: 18},
		{TokenChain: 10, TokenAddress: "000000000000000000000000f1277d1ed8ad466beddf92ef448a132661956621", Symbol: "WFTM", CoingeckoID: "wrapped-fantom", Decimals: 18},
		{TokenChain: 14, TokenAddress: "000000000000000000000000f194afdf50b03e69bd7d057c1aa9e10c9954e4c9", Symbol: "CELO", CoingeckoID: "celo", Decimals: 18},
		{TokenChain: 16, TokenAddress: "000000000000000000000000d909178cc99d318e4d46e7e66a972955859670e1", Symbol: "GLMR", CoingeckoID: "wrapped-moonbeam", Decimals: 18},
		{TokenChain: 21, TokenAddress: "587c29de216efd4219573e08a1f6964d4fa7cb714518c2c8a0f29abfa264327d", Symbol: "SUI", CoingeckoID: "sui", Decimals: 9}}
}
