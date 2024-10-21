package domain

import sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"

// manualMainnetTokenList returns a list of tokens that are not generated automatically.
func manualMainnetTokenList() []TokenMetadata {
	return []TokenMetadata{
		{TokenChain: sdk.ChainIDSolana, TokenAddress: "6927fdc01ea906f96d7137874cdd7adad00ca35764619310e54196c781d84d5b", Symbol: "W", CoingeckoID: "wormhole", Decimals: 6},    // Addr: 85VBFQZC9TZkfaptBWjvUw7YbZjy52A6mjtPGjstQAmQ
		{TokenChain: sdk.ChainIDEthereum, TokenAddress: "000000000000000000000000b0ffa8000886e57f86dd5264b9582b2ad87b2b91", Symbol: "W", CoingeckoID: "wormhole", Decimals: 18}, // Addr: 0xB0fFa8000886e57F86dd5264b9582b2Ad87b2b91
		{TokenChain: sdk.ChainIDArbitrum, TokenAddress: "000000000000000000000000b0ffa8000886e57f86dd5264b9582b2ad87b2b91", Symbol: "W", CoingeckoID: "wormhole", Decimals: 18}, // Addr: 0xB0fFa8000886e57F86dd5264b9582b2Ad87b2b91
		{TokenChain: sdk.ChainIDOptimism, TokenAddress: "000000000000000000000000b0ffa8000886e57f86dd5264b9582b2ad87b2b91", Symbol: "W", CoingeckoID: "wormhole", Decimals: 18}, // Addr: 0xB0fFa8000886e57F86dd5264b9582b2Ad87b2b91
		{TokenChain: sdk.ChainIDBase, TokenAddress: "000000000000000000000000b0ffa8000886e57f86dd5264b9582b2ad87b2b91", Symbol: "W", CoingeckoID: "wormhole", Decimals: 18},     // Addr: 0xB0fFa8000886e57F86dd5264b9582b2Ad87b2b91

		{TokenChain: sdk.ChainIDSolana, TokenAddress: "270ad0028e970df757d5f14f8cbb6a6810e48139125608ea958b718eb2944920", Symbol: "BORG", CoingeckoID: "swissborg", Decimals: 9},    // Addr: 3dQTr7ror2QPKQ3GbBCokJUmjErGg8kTJzdnYjNfvi3Z
		{TokenChain: sdk.ChainIDEthereum, TokenAddress: "00000000000000000000000064d0f55cd8c7133a9d7102b13987235f486f2224", Symbol: "BORG", CoingeckoID: "swissborg", Decimals: 18}, // Addr: 0x64d0f55cd8c7133a9d7102b13987235f486f2224

		{TokenChain: sdk.ChainIDFantom, TokenAddress: "000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48", Symbol: "USDC.e", CoingeckoID: "usdc", Decimals: 6}, // Addr: 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48

		{TokenChain: sdk.ChainIDArbitrum, TokenAddress: "00000000000000000000000083e1d2310ade410676b1733d16e89f91822fd5c3", Symbol: "JitoSOL", CoingeckoID: "jito-staked-sol", Decimals: 9}, // Addr: 0x83e1d2310Ade410676B1733d16e89f91822FD5c3

		{TokenChain: sdk.ChainIDSolana, TokenAddress: "8ea6bae83ada8cc0d7be5c2816a74e95d409603129bb2ee4fa22cc6f964a4d81", Symbol: "CHEESE", CoingeckoID: "cheese-2", Decimals: 6},    // Addr: AbrMJWfDVRZ2EWCQ1xSCpoVeVgZNpq1U2AoYG98oRXfn
		{TokenChain: sdk.ChainIDArbitrum, TokenAddress: "00000000000000000000000005aea20947a9a376ef50218633bb0a5a05d40a0c", Symbol: "CHEESE", CoingeckoID: "cheese-2", Decimals: 18}, // Addr: 0x05AEa20947A9A376eF50218633BB0a5A05d40A0C

		{TokenChain: sdk.ChainIDSolana, TokenAddress: "899c27db19ddb4f36442d15010e318819deca66868415ac29b9f50f18eef2e31", Symbol: "AGA", CoingeckoID: "agorahub", Decimals: 9},    // Addr: AGAxefyrPTi63FGL2ukJUTBtLJStDpiXMdtLRWvzambv
		{TokenChain: sdk.ChainIDEthereum, TokenAddress: "00000000000000000000000087b46212e805a3998b7e8077e9019c90759ea88c", Symbol: "AGA", CoingeckoID: "agorahub", Decimals: 18}, // Addr: 0x87B46212e805A3998B7e8077E9019c90759Ea88C

		// TODO: uncomment once coingecko id is available
		// {TokenChain: 1, TokenAddress: "07bb093e9f7decab41a717b15946f6db587868a6721c0a6e3c4281ed3fef0e09", Symbol: "XBG", CoingeckoID: "", Decimals: 9},   // Addr: XBGdqJ9P175hCC1LangCEyXWNeCPHaKWA17tymz2PrY
		// {TokenChain: sdk.ChainIDEthereum, TokenAddress: "000000000000000000000000eae00d6f9b16deb1bd584c7965e4c7d762f178a1", Symbol: "XBG", CoingeckoID: "", Decimals: 18},  // Addr: 0xEaE00D6F9B16Deb1BD584c7965e4c7d762f178a1
		// {TokenChain: sdk.ChainIDArbitrum, TokenAddress: "00000000000000000000000093fa0b88c0c78e45980fa74cdd87469311b7b3e4", Symbol: "XBG", CoingeckoID: "", Decimals: 18}, // Addr: 0x93FA0B88C0C78e45980Fa74cdd87469311b7B3E4

		{TokenChain: sdk.ChainIDSolana, TokenAddress: "067fc27abcad2df07cc40437330da4fe8851680ae2b242c2ea1d86e2cfa10064", Symbol: "SNS", CoingeckoID: "synesis-one", Decimals: 9}, // Addr: SNSNkV9zfG5ZKWQs6x4hxvBRV6s8SqMfSGCtECDvdMd

		{TokenChain: sdk.ChainIDEthereum, TokenAddress: "000000000000000000000000455e53cbb86018ac2b8092fdcd39d8444affc3f6", Symbol: "POL", CoingeckoID: "polygon-ecosystem-token", Decimals: 18}, // Addr: 0x455e53CBB86018Ac2B8092FdCd39d8444aFFC3F6

		{TokenChain: sdk.ChainIDSolana, TokenAddress: "e23fb17c7d654ebaf63b3b8caeca39a82bc78159bb9cebba8dbeeff90b4190c3", Symbol: "LITT", CoingeckoID: "litlab-games", Decimals: 9}, // Addr: GEBUHM7o5T1Ws2rAWjRijtYeh9XFxKrD3B4b9HV7dxLz
		{TokenChain: sdk.ChainIDBSC, TokenAddress: "000000000000000000000000cebef3df1f3c5bfd90fde603e71f31a53b11944d", Symbol: "LITT", CoingeckoID: "litlab-games", Decimals: 18},   // Addr: 0xCEbEf3DF1F3C5Bfd90FDE603E71F31a53B11944D

		{TokenChain: sdk.ChainIDSolana, TokenAddress: "20dedacad378f74d6cb4bdb1caf262228d4083d411f1dc92473de4d00ea9d0b8", Symbol: "REZ", CoingeckoID: "renzo", Decimals: 9},    // Addr: 3DK98MXPz8TRuim7rfQnebSLpA7VSoc79Bgiee1m4Zw5
		{TokenChain: sdk.ChainIDEthereum, TokenAddress: "0000000000000000000000003b50805453023a91a8bf641e279401a0b23fa6f9", Symbol: "REZ", CoingeckoID: "renzo", Decimals: 18}, // Addr: 0x3B50805453023a91a8bf641e279401a0b23FA6F9

		{TokenChain: sdk.ChainIDEthereum, TokenAddress: "000000000000000000000000dc035d45d973e3ec169d2276ddab16f1e407384f", Symbol: "USDS", CoingeckoID: "usds", Decimals: 18}, // Addr: 0xdC035D45d973E3EC169d2276DDab16f1e407384F
		{TokenChain: sdk.ChainIDSolana, TokenAddress: "0707312938a310834eb96d9a917a2d235fb55c4bd63c62d3d2063fad48eaac97", Symbol: "USDS", CoingeckoID: "usds", Decimals: 6},    // Addr: USDSmbcVPUStXmCNeH6i13LmRPMZic7Afz7uu7nVgrJ

		{TokenChain: sdk.ChainIDEthereum, TokenAddress: "00000000000000000000000056072c95faa701256059aa122697b133aded9279", Symbol: "SKY", CoingeckoID: "sky", Decimals: 18}, // Addr: 0x56072C95FAA701256059aa122697B133aDEd9279
		{TokenChain: sdk.ChainIDSolana, TokenAddress: "067c7a69702d4523c88fb07eb91510c1615dd8886ab41767c2461170efc7f703", Symbol: "SKY", CoingeckoID: "sky", Decimals: 6},    // Addr: SKY3ns1PY4rCyyu1n5WCNGnh7MPSjJai3fRcba12NZ8

		{TokenChain: sdk.ChainIDBSC, TokenAddress: "00000000000000000000000026c5e01524d2E6280A48F2c50fF6De7e52E9611C", Symbol: "wstETH", CoingeckoID: "wrapped-steth", Decimals: 18}, // Addr: 0x26c5e01524d2E6280A48F2c50fF6De7e52E9611C
	}
}

// mainnetTokenList returns a list of all tokens on the mainnet.
func mainnetTokenList() []TokenMetadata {
	res := append(generatedMainnetTokenList(), manualMainnetTokenList()...)
	res = append(res, GasTokenList()...)
	return append(res, unknownTokenList()...)
}

// GasTokenList : gas tokens are the ones used to pay gas fees on the respective chains, they don't belong to a contract address.
func GasTokenList() []TokenMetadata {
	const nativeTokenAddress = "0000000000000000000000000000000000000000000000000000000000000000"
	return []TokenMetadata{
		{TokenChain: sdk.ChainIDSolana, TokenAddress: nativeTokenAddress, Symbol: "SOL", CoingeckoID: "solana", Decimals: 9},
		{TokenChain: sdk.ChainIDEthereum, TokenAddress: nativeTokenAddress, Symbol: "ETH", CoingeckoID: "ethereum", Decimals: 18},
		{TokenChain: sdk.ChainIDTerra, TokenAddress: nativeTokenAddress, Symbol: "LUNA", CoingeckoID: "terra-luna", Decimals: 6},
		{TokenChain: sdk.ChainIDBSC, TokenAddress: nativeTokenAddress, Symbol: "BNB", CoingeckoID: "binancecoin", Decimals: 18},
		{TokenChain: sdk.ChainIDPolygon, TokenAddress: nativeTokenAddress, Symbol: "MATIC", CoingeckoID: "matic-network", Decimals: 18},
		{TokenChain: sdk.ChainIDAvalanche, TokenAddress: nativeTokenAddress, Symbol: "AVAX", CoingeckoID: "avalanche-2", Decimals: 18},
		{TokenChain: sdk.ChainIDOasis, TokenAddress: nativeTokenAddress, Symbol: "ROSE", CoingeckoID: "oasis-network", Decimals: 18},
		{TokenChain: sdk.ChainIDAlgorand, TokenAddress: nativeTokenAddress, Symbol: "ALGO", CoingeckoID: "algorand", Decimals: 6},
		{TokenChain: sdk.ChainIDAurora, TokenAddress: nativeTokenAddress, Symbol: "AOA", CoingeckoID: "aurora", Decimals: 18},
		{TokenChain: sdk.ChainIDFantom, TokenAddress: nativeTokenAddress, Symbol: "FTM", CoingeckoID: "fantom", Decimals: 18},
		{TokenChain: sdk.ChainIDKarura, TokenAddress: nativeTokenAddress, Symbol: "KAR", CoingeckoID: "karura", Decimals: 12},
		{TokenChain: sdk.ChainIDAcala, TokenAddress: nativeTokenAddress, Symbol: "ACA", CoingeckoID: "acala", Decimals: 18},
		{TokenChain: sdk.ChainIDKlaytn, TokenAddress: nativeTokenAddress, Symbol: "KLAY", CoingeckoID: "klay-token", Decimals: 18},
		{TokenChain: sdk.ChainIDCelo, TokenAddress: nativeTokenAddress, Symbol: "CELO", CoingeckoID: "celo", Decimals: 18},
		{TokenChain: sdk.ChainIDNear, TokenAddress: nativeTokenAddress, Symbol: "NEAR", CoingeckoID: "near", Decimals: 24},
		{TokenChain: sdk.ChainIDMoonbeam, TokenAddress: nativeTokenAddress, Symbol: "GLMR", CoingeckoID: "moonbeam", Decimals: 18},
		{TokenChain: sdk.ChainIDTerra2, TokenAddress: nativeTokenAddress, Symbol: "LUNA", CoingeckoID: "terra-luna-2", Decimals: 6},
		{TokenChain: sdk.ChainIDInjective, TokenAddress: nativeTokenAddress, Symbol: "INJ", CoingeckoID: "injective-protocol", Decimals: 18},
		{TokenChain: sdk.ChainIDOsmosis, TokenAddress: nativeTokenAddress, Symbol: "OSMO", CoingeckoID: "osmosis", Decimals: 6},
		{TokenChain: sdk.ChainIDSui, TokenAddress: nativeTokenAddress, Symbol: "SUI", CoingeckoID: "sui", Decimals: 9},
		{TokenChain: sdk.ChainIDAptos, TokenAddress: nativeTokenAddress, Symbol: "APT", CoingeckoID: "aptos", Decimals: 8},
		{TokenChain: sdk.ChainIDArbitrum, TokenAddress: nativeTokenAddress, Symbol: "ARB", CoingeckoID: "arbitrum", Decimals: 18},
		{TokenChain: sdk.ChainIDOptimism, TokenAddress: nativeTokenAddress, Symbol: "OP", CoingeckoID: "optimism", Decimals: 18},
		{TokenChain: sdk.ChainIDGnosis, TokenAddress: nativeTokenAddress, Symbol: "GNO", CoingeckoID: "gnosis", Decimals: 18},
		{TokenChain: sdk.ChainIDXpla, TokenAddress: nativeTokenAddress, Symbol: "XPLA", CoingeckoID: "xpla", Decimals: 18},
		{TokenChain: sdk.ChainIDBtc, TokenAddress: nativeTokenAddress, Symbol: "BTC", CoingeckoID: "bitcoin", Decimals: 8},
		{TokenChain: sdk.ChainIDBase, TokenAddress: nativeTokenAddress, Symbol: "ETH", CoingeckoID: "ethereum", Decimals: 18},
		{TokenChain: sdk.ChainIDSei, TokenAddress: nativeTokenAddress, Symbol: "SEI", CoingeckoID: "sei-network", Decimals: 6},
		{TokenChain: sdk.ChainIDRootstock, TokenAddress: nativeTokenAddress, Symbol: "RSK", CoingeckoID: "rootstock", Decimals: 18},
		{TokenChain: sdk.ChainIDScroll, TokenAddress: nativeTokenAddress, Symbol: "ETH", CoingeckoID: "ethereum", Decimals: 18},
		{TokenChain: sdk.ChainIDMantle, TokenAddress: nativeTokenAddress, Symbol: "MNT", CoingeckoID: "mantle", Decimals: 18},
		{TokenChain: sdk.ChainIDBlast, TokenAddress: nativeTokenAddress, Symbol: "ETH", CoingeckoID: "ethereum", Decimals: 18},
		{TokenChain: sdk.ChainIDXLayer, TokenAddress: nativeTokenAddress, Symbol: "XLYR", CoingeckoID: "xlayer", Decimals: 18},
		{TokenChain: sdk.ChainIDLinea, TokenAddress: nativeTokenAddress, Symbol: "ETH", CoingeckoID: "ethereum", Decimals: 18},
		{TokenChain: sdk.ChainIDBerachain, TokenAddress: nativeTokenAddress, Symbol: "BERA", CoingeckoID: "berachain-bera", Decimals: 18},
		//{TokenChain: sdk.ChainIDWormchain, TokenAddress: nativeTokenAddress, Symbol: "WORM", CoingeckoID: "wormchain", Decimals: 18}, // Currently Wormchain doesn't charge gas fees to wormhole messages. This may change in the future: https://docs.wormhole.com/wormhole/explore-wormhole/gateway
		{TokenChain: sdk.ChainIDCosmoshub, TokenAddress: nativeTokenAddress, Symbol: "ATOM", CoingeckoID: "cosmos", Decimals: 6},
		{TokenChain: sdk.ChainIDEvmos, TokenAddress: nativeTokenAddress, Symbol: "EVMOS", CoingeckoID: "evmos", Decimals: 18},
		{TokenChain: sdk.ChainIDKujira, TokenAddress: nativeTokenAddress, Symbol: "KUJI", CoingeckoID: "kujira", Decimals: 6},
		{TokenChain: sdk.ChainIDNeutron, TokenAddress: nativeTokenAddress, Symbol: "NEUT", CoingeckoID: "neutron-3", Decimals: 6},
		{TokenChain: sdk.ChainIDCelestia, TokenAddress: nativeTokenAddress, Symbol: "TIA", CoingeckoID: "celestia", Decimals: 6},
		{TokenChain: sdk.ChainIDStargaze, TokenAddress: nativeTokenAddress, Symbol: "STARS", CoingeckoID: "stargaze", Decimals: 6},
		{TokenChain: sdk.ChainIDSeda, TokenAddress: nativeTokenAddress, Symbol: "SEDA", CoingeckoID: "seda-2", Decimals: 18},
		{TokenChain: sdk.ChainIDDymension, TokenAddress: nativeTokenAddress, Symbol: "DYM", CoingeckoID: "dymension", Decimals: 18},
		{TokenChain: sdk.ChainIDProvenance, TokenAddress: nativeTokenAddress, Symbol: "HASH", CoingeckoID: "provenance-blockchain", Decimals: 9},

		// TODO: update go sdk dependency to have snaxchain support
		{TokenChain: 43, TokenAddress: nativeTokenAddress, Symbol: "ETH", CoingeckoID: "ethereum", Decimals: 18},
	}
}

func unknownTokenList() []TokenMetadata {
	return []TokenMetadata{
		{TokenChain: 23, TokenAddress: "00000000000000000000000007dd5beaffb65b8ff2e575d500bdf324a05295dc", Symbol: "arbi", CoingeckoID: "arbipad", Decimals: 18},
		{TokenChain: 23, TokenAddress: "0000000000000000000000003d9907f9a368ad0a51be60f7da3b97cf940982d8", Symbol: "grail", CoingeckoID: "camelot-token", Decimals: 18},
		{TokenChain: 23, TokenAddress: "0000000000000000000000004186bfc76e2e237523cbc30fd220fe055156b41f", Symbol: "rseth", CoingeckoID: "layerzero-bridged-rseth-linea", Decimals: 18},
		{TokenChain: 23, TokenAddress: "000000000000000000000000788d86e00ab31db859c3d6b80d5a9375801d7f2a", Symbol: "tenet", CoingeckoID: "tenet-1b000f7b-59cb-4e06-89ce-d62b32d362b9", Decimals: 18},
		{TokenChain: 23, TokenAddress: "000000000000000000000000e66998533a1992ece9ea99cdf47686f4fc8458e0", Symbol: "jusdc", CoingeckoID: "jones-usdc", Decimals: 18},
		{TokenChain: 9, TokenAddress: "000000000000000000000000c42c30ac6cc15fac9bd938618bcaa1a1fae8501d", Symbol: "wnear", CoingeckoID: "wrapped-near", Decimals: 24},
		{TokenChain: 9, TokenAddress: "000000000000000000000000f4eb217ba2454613b15dbdea6e5f22276410e89e", Symbol: "wbtc", CoingeckoID: "wrapped-bitcoin", Decimals: 8},
		{TokenChain: 6, TokenAddress: "000000000000000000000000027dbca046ca156de9622cd1e2d907d375e53aa7", Symbol: "ampl", CoingeckoID: "ampleforth", Decimals: 9},
		{TokenChain: 6, TokenAddress: "000000000000000000000000184ff13b3ebcb25be44e860163a5d8391dd568c1", Symbol: "kimbo", CoingeckoID: "kimbo", Decimals: 18},
		{TokenChain: 6, TokenAddress: "0000000000000000000000001c20e891bab6b1727d14da358fae2984ed9b59eb", Symbol: "tusd", CoingeckoID: "true-usd", Decimals: 18},
		{TokenChain: 6, TokenAddress: "0000000000000000000000001db749847c4abb991d8b6032102383e6bfd9b1c7", Symbol: "don", CoingeckoID: "dogeon", Decimals: 18},
		{TokenChain: 6, TokenAddress: "00000000000000000000000037b608519f91f70f2eeb0e5ed9af4061722e4f76", Symbol: "sushi", CoingeckoID: "sushi", Decimals: 18},
		{TokenChain: 6, TokenAddress: "00000000000000000000000039fc9e94caeacb435842fadedecb783589f50f5f", Symbol: "knc", CoingeckoID: "kyber-network-crystal", Decimals: 18},
		{TokenChain: 6, TokenAddress: "000000000000000000000000420fca0121dc28039145009570975747295f2329", Symbol: "coq", CoingeckoID: "coq-inu", Decimals: 18},
		{TokenChain: 6, TokenAddress: "0000000000000000000000005c49b268c9841aff1cc3b0a418ff5c3442ee3f3b", Symbol: "mimatic", CoingeckoID: "mai-avalanche", Decimals: 18},
		{TokenChain: 6, TokenAddress: "00000000000000000000000065378b697853568da9ff8eab60c13e1ee9f4a654", Symbol: "husky", CoingeckoID: "husky-avax", Decimals: 18},
		{TokenChain: 6, TokenAddress: "0000000000000000000000008729438eb15e2c8b576fcc6aecda6a148776c0f5", Symbol: "qi", CoingeckoID: "benqi", Decimals: 18},
		{TokenChain: 6, TokenAddress: "0000000000000000000000009c9e5fd8bbc25984b178fdce6117defa39d2db39", Symbol: "busd", CoingeckoID: "binance-peg-busd", Decimals: 18},
		{TokenChain: 6, TokenAddress: "000000000000000000000000abc9547b534519ff73921b1fba6e672b5f58d083", Symbol: "woo", CoingeckoID: "woo-network", Decimals: 18},
		{TokenChain: 6, TokenAddress: "000000000000000000000000acfb898cff266e53278cc0124fc2c7c94c8cb9a5", Symbol: "nochill", CoingeckoID: "avax-has-no-chill", Decimals: 18},
		{TokenChain: 6, TokenAddress: "000000000000000000000000afe3d2a31231230875dee1fa1eef14a412443d22", Symbol: "dfiat", CoingeckoID: "defiato", Decimals: 18},
		{TokenChain: 6, TokenAddress: "000000000000000000000000b09fe1613fe03e7361319d2a43edc17422f36b09", Symbol: "bog", CoingeckoID: "bogged-finance", Decimals: 18},
		{TokenChain: 6, TokenAddress: "000000000000000000000000bd83010eb60f12112908774998f65761cf9f6f9a", Symbol: "boo", CoingeckoID: "spookyswap", Decimals: 18},
		{TokenChain: 6, TokenAddress: "000000000000000000000000fc6da929c031162841370af240dec19099861d3b", Symbol: "domi", CoingeckoID: "domi", Decimals: 18},
		{TokenChain: 30, TokenAddress: "0000000000000000000000000a1d576f3efef75b330424287a95a366e8281d54", Symbol: "ausdbc", CoingeckoID: "aave-v3-usdbc", Decimals: 6},
		{TokenChain: 30, TokenAddress: "00000000000000000000000017931cfc3217261ce0fa21bb066633c463ed8634", Symbol: "based", CoingeckoID: "basedchad", Decimals: 18},
		{TokenChain: 30, TokenAddress: "00000000000000000000000019b50c63d3d7f7a22308cb0fc8d41b66ff9c318a", Symbol: "gpx", CoingeckoID: "grabpenny", Decimals: 18},
		{TokenChain: 30, TokenAddress: "0000000000000000000000001c7a460413dd4e964f96d8dfc56e7223ce88cd85", Symbol: "seam", CoingeckoID: "seamless-protocol", Decimals: 18},
		{TokenChain: 30, TokenAddress: "0000000000000000000000002598c30330d5771ae9f983979209486ae26de875", Symbol: "ai", CoingeckoID: "any-inu", Decimals: 18},
		{TokenChain: 30, TokenAddress: "00000000000000000000000074ccbe53f77b08632ce0cb91d3a545bf6b8e0979", Symbol: "bomb", CoingeckoID: "fbomb", Decimals: 18},
		{TokenChain: 30, TokenAddress: "00000000000000000000000076734b57dfe834f102fb61e1ebf844adf8dd931e", Symbol: "weirdo", CoingeckoID: "weirdo-2", Decimals: 8},
		{TokenChain: 30, TokenAddress: "00000000000000000000000096419929d7949d6a801a6909c145c8eef6a40431", Symbol: "spec", CoingeckoID: "spectral", Decimals: 18},
		{TokenChain: 30, TokenAddress: "0000000000000000000000009a3b7959e998bf2b50ef1969067d623877050d92", Symbol: "pbb", CoingeckoID: "pepe-but-blue", Decimals: 18},
		{TokenChain: 30, TokenAddress: "000000000000000000000000a88594d404727625a9437c3f886c7643872296ae", Symbol: "well", CoingeckoID: "moonwell-artemis", Decimals: 18},
		{TokenChain: 30, TokenAddress: "000000000000000000000000ba5e6fa2f33f3955f0cef50c63dcc84861eab663", Symbol: "based", CoingeckoID: "based-markets", Decimals: 18},
		{TokenChain: 30, TokenAddress: "000000000000000000000000c1cba3fcea344f92d9239c08c0568f6f2f0ee452", Symbol: "wsteth", CoingeckoID: "wrapped-steth", Decimals: 18},
		{TokenChain: 30, TokenAddress: "000000000000000000000000f7c1cefcf7e1dd8161e00099facd3e1db9e528ee", Symbol: "tower", CoingeckoID: "tower", Decimals: 18},
		{TokenChain: 4, TokenAddress: "00000000000000000000000014016e85a25aeb13065688cafb43044c2ef86784", Symbol: "tusd", CoingeckoID: "bridged-trueusd", Decimals: 18},
		{TokenChain: 4, TokenAddress: "00000000000000000000000016faf9daa401aa42506af503aa3d80b871c467a3", Symbol: "dck", CoingeckoID: "dexcheck", Decimals: 18},
		{TokenChain: 4, TokenAddress: "0000000000000000000000001d8e589379cf74a276952b56856033ad0d489fb3", Symbol: "milkai", CoingeckoID: "milkai", Decimals: 8},
		{TokenChain: 4, TokenAddress: "0000000000000000000000001dacbcd9b3fc2d6a341dca3634439d12bc71ca4d", Symbol: "bvt", CoingeckoID: "bovineverse-bvt", Decimals: 18},
		{TokenChain: 4, TokenAddress: "0000000000000000000000002a45a892877ef383c5fc93a5206546c97496da9e", Symbol: "x", CoingeckoID: "x-ai", Decimals: 9},
		{TokenChain: 4, TokenAddress: "0000000000000000000000002b72867c32cf673f7b02d208b26889fed353b1f8", Symbol: "sqr", CoingeckoID: "magic-square", Decimals: 8},
		{TokenChain: 4, TokenAddress: "0000000000000000000000003f56e0c36d275367b8c502090edf38289b3dea0d", Symbol: "mimatic", CoingeckoID: "mai-arbitrum", Decimals: 18},
		{TokenChain: 4, TokenAddress: "00000000000000000000000042c95788f791a2be3584446854c8d9bb01be88a9", Symbol: "hbr", CoingeckoID: "harbor-3", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000465707181acba42ed01268a33f0507e320a154bd", Symbol: "step", CoingeckoID: "step", Decimals: 18},
		{TokenChain: 4, TokenAddress: "0000000000000000000000005e7f472b9481c80101b22d0ba4ef4253aa61dabc", Symbol: "mudol2", CoingeckoID: "hero-blaze-three-kingdoms", Decimals: 18},
		{TokenChain: 4, TokenAddress: "00000000000000000000000069b14e8d3cebfdd8196bfe530954a0c226e5008e", Symbol: "spacepi", CoingeckoID: "spacepi-token", Decimals: 9},
		{TokenChain: 4, TokenAddress: "000000000000000000000000734548a9e43d2d564600b1b2ed5be9c2b911c6ab", Symbol: "peel", CoingeckoID: "meta-apes-peel", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000740df024ce73f589acd5e8756b377ef8c6558bab", Symbol: "hlg", CoingeckoID: "holograph", Decimals: 18},
		{TokenChain: 4, TokenAddress: "0000000000000000000000007deb9906bd1d77b410a56e5c23c36340bd60c983", Symbol: "static", CoingeckoID: "chargedefi-static", Decimals: 18},
		{TokenChain: 4, TokenAddress: "0000000000000000000000007f280dac515121dcda3eac69eb4c13a52392cace", Symbol: "fnc", CoingeckoID: "fancy-games", Decimals: 18},
		{TokenChain: 4, TokenAddress: "0000000000000000000000007f792db54b0e580cdc755178443f0430cf799aca", Symbol: "volt", CoingeckoID: "volt-inu-2", Decimals: 9},
		{TokenChain: 4, TokenAddress: "0000000000000000000000008729438eb15e2c8b576fcc6aecda6a148776c0f5", Symbol: "qi", CoingeckoID: "benqi", Decimals: 18},
		{TokenChain: 4, TokenAddress: "0000000000000000000000008bfca09e5877ea59f85883d13a6873334b937d41", Symbol: "madpepe", CoingeckoID: "mad-pepe", Decimals: 18},
		{TokenChain: 4, TokenAddress: "0000000000000000000000008db1d28ee0d822367af8d220c0dc7cb6fe9dc442", Symbol: "ethpad", CoingeckoID: "ethpad", Decimals: 18},
		{TokenChain: 4, TokenAddress: "0000000000000000000000008f0528ce5ef7b51152a59745befdd91d97091d2f", Symbol: "alpaca", CoingeckoID: "alpaca-finance", Decimals: 18},
		{TokenChain: 4, TokenAddress: "0000000000000000000000009b83f827928abdf18cf1f7e67053572b9bceff3a", Symbol: "artem", CoingeckoID: "artem", Decimals: 18},
		{TokenChain: 4, TokenAddress: "0000000000000000000000009bf543d8460583ff8a669aae01d9cdbee4defe3c", Symbol: "sko", CoingeckoID: "sugar-kingdom-odyssey", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000a25199a79a34cc04b15e5c0bba4e3a557364e532", Symbol: "rim", CoingeckoID: "metarim", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000b0e1fc65c1a741b4662b813eb787d369b8614af1", Symbol: "if", CoingeckoID: "impossible-finance", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000b64e280e9d1b5dbec4accedb2257a87b400db149", Symbol: "lvl", CoingeckoID: "level", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000bbca42c60b5290f2c48871a596492f93ff0ddc82", Symbol: "domi", CoingeckoID: "domi", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000bceee918077f63fb1b9e10403f59ca40c7061341", Symbol: "papadoge", CoingeckoID: "papa-doge", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000bededdf2ef49e87037c4fb2ca34d1ff3d3992a11", Symbol: "feg", CoingeckoID: "feg-bsc", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000c17c30e98541188614df99239cabd40280810ca3", Symbol: "rise", CoingeckoID: "everrise", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000c28e27870558cf22add83540d2126da2e4b464c2", Symbol: "sashimi", CoingeckoID: "sashimi", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000c6759a4fc56b3ce9734035a56b36e8637c45b77e", Symbol: "grimace", CoingeckoID: "grimace-coin", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000c703da39ae3b9db67c207c7bad8100e1afdc0f9c", Symbol: "frgx", CoingeckoID: "frgx-finance", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000c7981767f644c7f8e483dabdc413e8a371b83079", Symbol: "liq", CoingeckoID: "liquidus", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000cc7a91413769891de2e9ebbfc96d2eb1874b5760", Symbol: "gov", CoingeckoID: "govworld", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000ce7de646e7208a4ef112cb6ed5038fa6cc6b12e3", Symbol: "trx", CoingeckoID: "tron-bsc", Decimals: 6},
		{TokenChain: 4, TokenAddress: "000000000000000000000000d0aa796e2160ed260c668e90ac5f237b4ebd4b0d", Symbol: "waifu", CoingeckoID: "waifu", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000d8047afecb86e44eff3add991b9f063ed4ca716b", Symbol: "ggg", CoingeckoID: "good-games-guild", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000d9c2d319cd7e6177336b0a9c93c21cb48d84fb54", Symbol: "hapi", CoingeckoID: "hapi", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000e7f72bc0252ca7b16dbb72eeee1afcdb2429f2dd", Symbol: "nftl", CoingeckoID: "nftlaunch", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000e8377a076adabb3f9838afb77bee96eac101ffb1", Symbol: "msu", CoingeckoID: "metasoccer", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000f00600ebc7633462bc4f9c61ea2ce99f5aaebd4a", Symbol: "rose", CoingeckoID: "oasis-network", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000f93f6b686f4a6557151455189a9173735d668154", Symbol: "lfg", CoingeckoID: "gamerse", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000f95a5532d67c944dfa7eddd2f8c358fe0dc7fac2", Symbol: "mbx", CoingeckoID: "marblex", Decimals: 18},
		{TokenChain: 4, TokenAddress: "000000000000000000000000fa4ba88cf97e282c505bea095297786c16070129", Symbol: "cusd", CoingeckoID: "coin98-dollar", Decimals: 6},
		{TokenChain: 14, TokenAddress: "00000000000000000000000048065fbbe25f71c9282ddf5e1cd6d6a887483d5e", Symbol: "usdt", CoingeckoID: "tether", Decimals: 6},
		{TokenChain: 14, TokenAddress: "0000000000000000000000006e512bfc33be36f2666754e996ff103ad1680cc9", Symbol: "abr", CoingeckoID: "allbridge", Decimals: 18},
		{TokenChain: 14, TokenAddress: "000000000000000000000000d15ec721c2a896512ad29c671997dd68f9593226", Symbol: "sushi", CoingeckoID: "sushi", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000005d1123878fc55fbd56b54c73963b234a64af3c", Symbol: "kiba", CoingeckoID: "kiba-inu", Decimals: 18},
		{TokenChain: 2, TokenAddress: "00000000000000000000000009395a2a58db45db0da254c7eaa5ac469d8bdc85", Symbol: "sqt", CoingeckoID: "subquery-network", Decimals: 18},
		{TokenChain: 2, TokenAddress: "00000000000000000000000009f098b155d561fc9f7bccc97038b7e3d20baf74", Symbol: "zoo", CoingeckoID: "zoodao", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000000f2d719407fdbeff09d87557abb7232601fd9f29", Symbol: "syn", CoingeckoID: "synapse-2", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000001155db64b59265f57533bc0f9ae012fffd34eb7f", Symbol: "yaku", CoingeckoID: "yaku", Decimals: 9},
		{TokenChain: 2, TokenAddress: "00000000000000000000000012970e6868f88f6557b76120662c1b3e50a646bf", Symbol: "ladys", CoingeckoID: "milady-meme-coin", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000191557728e4d8caa4ac94f86af842148c0fa8f7e", Symbol: "eco", CoingeckoID: "ormeus-ecosystem", Decimals: 8},
		{TokenChain: 2, TokenAddress: "0000000000000000000000001a4b46696b2bb4794eb3d4c26f1c55f9170fa4c5", Symbol: "bit", CoingeckoID: "bitdao", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000001a57367c6194199e5d9aea1ce027431682dfb411", Symbol: "mdf", CoingeckoID: "matrixetf", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000001da87b114f35e1dc91f72bf57fc07a768ad40bb0", Symbol: "eqz", CoingeckoID: "equalizer", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000001e4e46b7bf03ece908c88ff7cc4975560010893a", Symbol: "ioen", CoingeckoID: "internet-of-energy-network", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000001fc5ef0337aea85c5f9198853a6e3a579a7a6987", Symbol: "reap", CoingeckoID: "reapchain", Decimals: 18},
		{TokenChain: 2, TokenAddress: "00000000000000000000000020a62aca58526836165ca53fe67dd884288c8abf", Symbol: "rnb", CoingeckoID: "rentible", Decimals: 18},
		{TokenChain: 2, TokenAddress: "00000000000000000000000024fcfc492c1393274b6bcd568ac9e225bec93584", Symbol: "mavia", CoingeckoID: "heroes-of-mavia", Decimals: 18},
		{TokenChain: 2, TokenAddress: "00000000000000000000000026c8afbbfe1ebaca03c2bb082e69d0476bffe099", Symbol: "cell", CoingeckoID: "cellframe", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000298d492e8c1d909d3f63bc4a36c66c64acb3d695", Symbol: "pbr", CoingeckoID: "polkabridge", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000299a1503e88433c0fd1bd68625c25c5a703eb64f", Symbol: "tear", CoingeckoID: "tear", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000002a8e1e676ec238d8a992307b495b45b3feaa5e86", Symbol: "ousd", CoingeckoID: "origin-dollar", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000002b867efd2de4ad2b583ca0cb3df9c4040ef4d329", Symbol: "lblock", CoingeckoID: "lucky-block", Decimals: 9},
		{TokenChain: 2, TokenAddress: "0000000000000000000000002b89bf8ba858cd2fcee1fada378d5cd6936968be", Symbol: "wscrt", CoingeckoID: "secret-erc20", Decimals: 6},
		{TokenChain: 2, TokenAddress: "0000000000000000000000003b79a28264fc52c7b4cea90558aa0b162f7faf57", Symbol: "wmemo", CoingeckoID: "wrapped-memory", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000003e9bc21c9b189c09df3ef1b824798658d5011937", Symbol: "lina", CoingeckoID: "linear", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000003ea8ea4237344c9931214796d9417af1a1180770", Symbol: "flx", CoingeckoID: "flux-token", Decimals: 18},
		{TokenChain: 2, TokenAddress: "00000000000000000000000042bbfa2e77757c645eeaad1655e0911a7553efbc", Symbol: "boba", CoingeckoID: "boba-network", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000430ef9263e76dae63c84292c3409d61c598e9682", Symbol: "pyr", CoingeckoID: "vulcan-forged", Decimals: 18},
		{TokenChain: 2, TokenAddress: "00000000000000000000000043dfc4159d86f3a37a5a4b3d4580b888ad7d4ddd", Symbol: "dodo", CoingeckoID: "dodo", Decimals: 18},
		{TokenChain: 2, TokenAddress: "00000000000000000000000044709a920fccf795fbc57baa433cc3dd53c44dbe", Symbol: "radar", CoingeckoID: "dappradar", Decimals: 18},
		{TokenChain: 2, TokenAddress: "00000000000000000000000044f5909e97e1cbf5fbbdf0fc92fd83cde5d5c58a", Symbol: "acria", CoingeckoID: "acria", Decimals: 18},
		{TokenChain: 2, TokenAddress: "00000000000000000000000046d0dac0926fa16707042cadc23f1eb4141fe86b", Symbol: "snm", CoingeckoID: "sonm", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000004c9edd5852cd905f086c759e8383e09bff1e68b3", Symbol: "usde", CoingeckoID: "ethena-usde", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000004cf89ca06ad997bc732dc876ed2a7f26a9e7f361", Symbol: "myst", CoingeckoID: "mysterium", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000004dd942baa75810a3c1e876e79d5cd35e09c97a76", Symbol: "d2t", CoingeckoID: "dash-2-trade", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000004ddc2d193948926d02f9b1fe9e1daa0718270ed5", Symbol: "ceth", CoingeckoID: "compound-ether", Decimals: 8},
		{TokenChain: 2, TokenAddress: "0000000000000000000000004e3fbd56cd56c3e72c1403e103b45db9da5b9d2b", Symbol: "cvx", CoingeckoID: "convex-finance", Decimals: 18},
		{TokenChain: 2, TokenAddress: "00000000000000000000000054012cdf4119de84218f7eb90eeb87e25ae6ebd7", Symbol: "luffy", CoingeckoID: "luffy-inu", Decimals: 9},
		{TokenChain: 2, TokenAddress: "0000000000000000000000005faa989af96af85384b8a938c2ede4a7378d9875", Symbol: "gal", CoingeckoID: "project-galaxy", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000614da3b37b6f66f7ce69b4bbbcf9a55ce6168707", Symbol: "mmx", CoingeckoID: "m2-global-wealth-limited-mmx", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000006226e00bcac68b0fe55583b90a1d727c14fab77f", Symbol: "mtv", CoingeckoID: "multivac", Decimals: 18},
		{TokenChain: 2, TokenAddress: "00000000000000000000000062d0a8458ed7719fdaf978fe5929c6d342b0bfce", Symbol: "beam", CoingeckoID: "beam-2", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000630d98424efe0ea27fb1b3ab7741907dffeaad78", Symbol: "peak", CoingeckoID: "marketpeak", Decimals: 8},
		{TokenChain: 2, TokenAddress: "000000000000000000000000667102bd3413bfeaa3dffb48fa8288819e480a88", Symbol: "tkx", CoingeckoID: "tokenize-xchange", Decimals: 8},
		{TokenChain: 2, TokenAddress: "000000000000000000000000668dbf100635f593a3847c0bdaf21f0a09380188", Symbol: "bnsd", CoingeckoID: "bnsd-finance", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000006781a0f84c7e9e846dcb84a9a5bd49333067b104", Symbol: "zap", CoingeckoID: "zap", Decimals: 18},
		{TokenChain: 2, TokenAddress: "00000000000000000000000068bbed6a47194eff1cf514b50ea91895597fc91e", Symbol: "andy", CoingeckoID: "andy-the-wisguy", Decimals: 18},
		{TokenChain: 2, TokenAddress: "00000000000000000000000069fa0fee221ad11012bab0fdb45d444d3d2ce71c", Symbol: "xrune", CoingeckoID: "thorstarter", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000006a6c2ada3ce053561c2fbc3ee211f23d9b8c520a", Symbol: "ton", CoingeckoID: "tontoken", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000006c28aef8977c9b773996d0e8376d2ee379446f2f", Symbol: "quick", CoingeckoID: "quick", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000722a89f1b925fe41883978219c2176aecc7d6699", Symbol: "xnk", CoingeckoID: "kinka", Decimals: 18},
		{TokenChain: 2, TokenAddress: "00000000000000000000000075d86078625d1e2f612de2627d34c7bc411c18b8", Symbol: "agii", CoingeckoID: "agii", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000799ebfabe77a6e34311eeee9825190b9ece32824", Symbol: "btrst", CoingeckoID: "braintrust", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000007fd4d7737597e7b4ee22acbf8d94362343ae0a79", Symbol: "wmc", CoingeckoID: "wrapped-mistcoin", Decimals: 2},
		{TokenChain: 2, TokenAddress: "0000000000000000000000008185bc4757572da2a610f887561c32298f1a5748", Symbol: "aln", CoingeckoID: "aluna", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000008503a7b00b4b52692cc6c14e5b96f142e30547b7", Symbol: "meed", CoingeckoID: "meeds-dao", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000008a2279d4a90b6fe1c4b30fa660cc9f926797baa2", Symbol: "chr", CoingeckoID: "chromaway", Decimals: 6},
		{TokenChain: 2, TokenAddress: "000000000000000000000000909e34d3f6124c324ac83dcca84b74398a6fa173", Symbol: "zkp", CoingeckoID: "panther", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000009506d37f70eb4c3d79c398d326c871abbf10521d", Symbol: "mlt", CoingeckoID: "media-licensing-token", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000968cbe62c830a0ccf4381614662398505657a2a9", Symbol: "tpy", CoingeckoID: "thrupenny", Decimals: 8},
		{TokenChain: 2, TokenAddress: "000000000000000000000000970b9bb2c0444f5e81e9d0efb84c8ccdcdcaf84d", Symbol: "fuse", CoingeckoID: "fuse-network-token", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000993864e43caa7f7f12953ad6feb1d1ca635b875f", Symbol: "sdao", CoingeckoID: "singularitydao", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000009ba021b0a9b958b5e75ce9f6dff97c7ee52cb3e6", Symbol: "apxeth", CoingeckoID: "dinero-apxeth", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000009d409a0a012cfba9b15f6d4b36ac57a46966ab9a", Symbol: "yvboost", CoingeckoID: "yvboost", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000009ee91f9f426fa633d227f7a9b000e28b9dfd8599", Symbol: "stmatic", CoingeckoID: "lido-staked-matic", Decimals: 18},
		{TokenChain: 2, TokenAddress: "0000000000000000000000009fb83c0635de2e815fd1c21b3a292277540c2e8d", Symbol: "fevr", CoingeckoID: "realfevr", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000a0b73e1ff0b80914ab6fe0444e65848c4c34450b", Symbol: "cro", CoingeckoID: "crypto-com-chain", Decimals: 8},
		{TokenChain: 2, TokenAddress: "000000000000000000000000a1faa113cbe53436df28ff0aee54275c13b40975", Symbol: "alpha", CoingeckoID: "alpha-finance", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000a41f142b6eb2b164f8164cae0716892ce02f311f", Symbol: "avg", CoingeckoID: "avaocado-dao", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000a58a4f5c4bb043d2cc1e170613b74e767c94189b", Symbol: "utu", CoingeckoID: "utu-coin", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000aa6e8127831c9de45ae56bb1b0d4d4da6e5665bd", Symbol: "eth2x-fli", CoingeckoID: "eth-2x-flexible-leverage-index", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000aa8330fb2b4d5d07abfe7a72262752a8505c6b37", Symbol: "polc", CoingeckoID: "polka-city", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000aaef88cea01475125522e117bfe45cf32044e238", Symbol: "gf", CoingeckoID: "guildfi", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000ac6db8954b73ebf10e84278ac8b9b22a781615d9", Symbol: "bwb", CoingeckoID: "bitget-wallet-token", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000aea46a60368a7bd060eec7df8cba43b7ef41ad85", Symbol: "fet", CoingeckoID: "fetch-ai", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000af5191b0de278c7286d6c7cc6ab6bb8a73ba2cd6", Symbol: "stg", CoingeckoID: "stargate-finance", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000b369daca21ee035312176eb8cf9d88ce97e0aa95", Symbol: "$skol", CoingeckoID: "skol", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000b3999f658c0391d94a37f7ff328f3fec942bcadc", Symbol: "hft", CoingeckoID: "hashflow", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000b62132e35a6c13ee1ee0f84dc5d40bad8d815206", Symbol: "nexo", CoingeckoID: "nexo", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000b9f599ce614feb2e1bbe58f180f370d05b39344e", Symbol: "pork", CoingeckoID: "pepefork", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000bbc2ae13b23d715c30720f079fcd9b4a74093505", Symbol: "ern", CoingeckoID: "ethernity-chain", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000bd100d061e120b2c67a24453cf6368e63f1be056", Symbol: "idyp", CoingeckoID: "idefiyieldprotocol", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000be9895146f7af43049ca1c1ae358b0541ea49704", Symbol: "cbeth", CoingeckoID: "coinbase-wrapped-staked-eth", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000beef01060047522408756e0000a90ce195a70000", Symbol: "aptr", CoingeckoID: "aperture-finance", Decimals: 6},
		{TokenChain: 2, TokenAddress: "000000000000000000000000c55126051b22ebb829d00368f4b12bde432de5da", Symbol: "btrfly", CoingeckoID: "redacted", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000ce246eea10988c495b4a90a905ee9237a0f91543", Symbol: "vcx", CoingeckoID: "vaultcraft", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000cf0c122c6b73ff809c693db761e7baebe62b6a2e", Symbol: "floki", CoingeckoID: "floki", Decimals: 9},
		{TokenChain: 2, TokenAddress: "000000000000000000000000cfeb09c3c5f0f78ad72166d55f9e6e9a60e96eec", Symbol: "vemp", CoingeckoID: "vempire-ddao", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000d1420af453fd7bf940573431d416cace7ff8280c", Symbol: "agov", CoingeckoID: "answer-governance", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000d1d2eb1b1e90b638588728b4130137d262c87cae", Symbol: "gala", CoingeckoID: "gala", Decimals: 8},
		{TokenChain: 2, TokenAddress: "000000000000000000000000d31695a1d35e489252ce57b129fd4b1b05e6acac", Symbol: "tkp", CoingeckoID: "tokpie", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000d567b5f02b9073ad3a982a099a23bf019ff11d1c", Symbol: "game", CoingeckoID: "gamestarter", Decimals: 5},
		{TokenChain: 2, TokenAddress: "000000000000000000000000da816459f1ab5631232fe5e97a05bbbb94970c95", Symbol: "yvdai", CoingeckoID: "yvdai", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000dc5e9445169c73cf21e1da0b270e8433cac69959", Symbol: "ethereum", CoingeckoID: "ketaicoin", Decimals: 9},
		{TokenChain: 2, TokenAddress: "000000000000000000000000ddf7fd345d54ff4b40079579d4c4670415dbfd0a", Symbol: "sg", CoingeckoID: "social-good-project", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000e86df1970055e9caee93dae9b7d5fd71595d0e18", Symbol: "btc20", CoingeckoID: "bitcoin20", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000eeaa40b28a2d1b0b08f6f97bb1dd4b75316c6107", Symbol: "govi", CoingeckoID: "govi", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000ef19f4e48830093ce5bc8b3ff7f903a0ae3e9fa1", Symbol: "botx", CoingeckoID: "botxcoin", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000f1c9acdc66974dfb6decb12aa385b9cd01190e38", Symbol: "oseth", CoingeckoID: "stakewise-v3-oseth", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000f29ae508698bdef169b89834f76704c3b205aedf", Symbol: "yvsnx", CoingeckoID: "snx-yvault", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000f519381791c03dd7666c142d4e49fd94d3536011", Symbol: "asia", CoingeckoID: "asia-coin", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000f94b5c5651c888d928439ab6514b93944eee6f48", Symbol: "yld", CoingeckoID: "yield-app", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000faba6f8e4a5e8ab82f62fe7c39859fa577269be3", Symbol: "ondo", CoingeckoID: "ondo-finance", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000fae4ee59cdd86e3be9e8b90b53aa866327d7c090", Symbol: "cpc", CoingeckoID: "cpchain", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000fbeea1c75e4c4465cb2fccc9c6d6afe984558e20", Symbol: "ddim", CoingeckoID: "duckdaodime", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000fc82bb4ba86045af6f327323a46e80412b91b27d", Symbol: "prom", CoingeckoID: "prometeus", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000fceb206e1a80527908521121358b5e26caabaa75", Symbol: "main", CoingeckoID: "main", Decimals: 18},
		{TokenChain: 2, TokenAddress: "000000000000000000000000ff20817765cb7f73d4bde2e66e067e58d11095c2", Symbol: "amp", CoingeckoID: "amp-token", Decimals: 18},
		{TokenChain: 10, TokenAddress: "00000000000000000000000002838746d9e1413e07ee064fcbada57055417f21", Symbol: "grain", CoingeckoID: "granary", Decimals: 18},
		{TokenChain: 10, TokenAddress: "00000000000000000000000027749e79ad796c4251e0a0564aef45235493a0b6", Symbol: "onx", CoingeckoID: "onx-finance", Decimals: 18},
		{TokenChain: 10, TokenAddress: "00000000000000000000000029b0da86e484e1c0029b56e817912d778ac0ec69", Symbol: "yfi", CoingeckoID: "yearn-finance", Decimals: 18},
		{TokenChain: 10, TokenAddress: "0000000000000000000000005f7f94a1dd7b15594d17543beb8b30b111dd464c", Symbol: "space", CoingeckoID: "space-token-bsc", Decimals: 18},
		{TokenChain: 10, TokenAddress: "0000000000000000000000007d016eec9c25232b01f23ef992d98ca97fc2af5a", Symbol: "fxs", CoingeckoID: "frax-share", Decimals: 18},
		{TokenChain: 10, TokenAddress: "00000000000000000000000091fa20244fb509e8289ca630e5db3e9166233fdc", Symbol: "gohm", CoingeckoID: "governance-ohm", Decimals: 18},
		{TokenChain: 10, TokenAddress: "0000000000000000000000009bd0611610a0f5133e4dd1bfdd71c5479ee77f37", Symbol: "ftmo", CoingeckoID: "fantom-oasis", Decimals: 18},
		{TokenChain: 10, TokenAddress: "000000000000000000000000c758295cd1a564cdb020a78a681a838cf8e0627d", Symbol: "fs", CoingeckoID: "fantomstarter", Decimals: 18},
		{TokenChain: 10, TokenAddress: "000000000000000000000000ddc0385169797937066bbd8ef409b5b3c0dfeb52", Symbol: "wmemo", CoingeckoID: "wrapped-memory", Decimals: 18},
		{TokenChain: 10, TokenAddress: "000000000000000000000000ddcb3ffd12750b45d32e084887fdf1aabab34239", Symbol: "any", CoingeckoID: "anyswap", Decimals: 18},
		{TokenChain: 10, TokenAddress: "000000000000000000000000e1e6b01ae86ad82b1f1b4eb413b219ac32e17bf6", Symbol: "xrune", CoingeckoID: "thorstarter", Decimals: 18},
		{TokenChain: 13, TokenAddress: "00000000000000000000000017d2628d30f8e9e966c9ba831c9b9b01ea8ea75c", Symbol: "isk", CoingeckoID: "iskra-token", Decimals: 18},
		{TokenChain: 13, TokenAddress: "000000000000000000000000574e9c26bda8b95d7329505b4657103710eb32ea", Symbol: "obnb", CoingeckoID: "orbit-bridge-klaytn-binance-coin", Decimals: 18},
		{TokenChain: 16, TokenAddress: "000000000000000000000000524d524b4c9366be706d3a90dcf70076ca037ae3", Symbol: "rmrk", CoingeckoID: "rmrk", Decimals: 18},
		{TokenChain: 16, TokenAddress: "0000000000000000000000006a2d262d56735dba19dd70682b39f6be9a931d98", Symbol: "ceusdc", CoingeckoID: "usd-coin-celer", Decimals: 6},
		{TokenChain: 16, TokenAddress: "0000000000000000000000007cd3e6e1a69409def0d78d17a492e8e143f40ec5", Symbol: "zoo", CoingeckoID: "zoodao", Decimals: 18},
		{TokenChain: 16, TokenAddress: "000000000000000000000000922d641a426dcffaef11680e5358f34d97d112e1", Symbol: "wbtc", CoingeckoID: "wrapped-bitcoin", Decimals: 8},
		{TokenChain: 16, TokenAddress: "000000000000000000000000dfa46478f9e5ea86d57387849598dbfb2e964b02", Symbol: "mimatic", CoingeckoID: "mimatic", Decimals: 18},
		{TokenChain: 24, TokenAddress: "00000000000000000000000048a9f8b4b65a55cc46ea557a610acf227454ab09", Symbol: "opc", CoingeckoID: "op-chads", Decimals: 18},
		{TokenChain: 24, TokenAddress: "000000000000000000000000b0b195aefa3650a6908f15cdac7d92f8a5791b0b", Symbol: "bob", CoingeckoID: "bob", Decimals: 18},
		{TokenChain: 24, TokenAddress: "000000000000000000000000dd69db25f6d620a7bad3023c5d32761d353d3de9", Symbol: "geth", CoingeckoID: "goerli-eth", Decimals: 18},
		{TokenChain: 24, TokenAddress: "000000000000000000000000dfa46478f9e5ea86d57387849598dbfb2e964b02", Symbol: "mimatic", CoingeckoID: "mai-optimism", Decimals: 18},
		{TokenChain: 5, TokenAddress: "000000000000000000000000111111111117dc0aa78b770fa6a738034120c302", Symbol: "1inch", CoingeckoID: "1inch", Decimals: 18},
		{TokenChain: 5, TokenAddress: "0000000000000000000000001796ae0b0fa4862485106a0de9b654efe301d0b2", Symbol: "pmon", CoingeckoID: "polychain-monsters", Decimals: 18},
		{TokenChain: 5, TokenAddress: "00000000000000000000000028424507fefb6f7f8e9d3860f56504e4e5f5f390", Symbol: "amweth", CoingeckoID: "aave-polygon-weth", Decimals: 18},
		{TokenChain: 5, TokenAddress: "0000000000000000000000002da719db753dfa10a62e140f436e1d67f2ddb0d6", Symbol: "cere", CoingeckoID: "cere-network", Decimals: 10},
		{TokenChain: 5, TokenAddress: "0000000000000000000000004f604735c1cf31399c6e711d5962b2b3e0225ad3", Symbol: "usdglo", CoingeckoID: "glo-dollar", Decimals: 18},
		{TokenChain: 5, TokenAddress: "000000000000000000000000580e933d90091b9ce380740e3a4a39c67eb85b4c", Symbol: "gswift", CoingeckoID: "gameswift", Decimals: 18},
		{TokenChain: 5, TokenAddress: "0000000000000000000000006d80113e533a2c0fe82eabd35f1875dcea89ea97", Symbol: "aeurs", CoingeckoID: "aave-v3-eurs", Decimals: 2},
		{TokenChain: 5, TokenAddress: "00000000000000000000000092868a5255c628da08f550a858a802f5351c5223", Symbol: "bridge", CoingeckoID: "cross-chain-bridge", Decimals: 18},
		{TokenChain: 5, TokenAddress: "000000000000000000000000aa9654becca45b5bdfa5ac646c939c62b527d394", Symbol: "dino", CoingeckoID: "dinoswap", Decimals: 18},
		{TokenChain: 5, TokenAddress: "000000000000000000000000bc5b59ea1b6f8da8258615ee38d40e999ec5d74f", Symbol: "paw", CoingeckoID: "paw-v2", Decimals: 18},
		{TokenChain: 5, TokenAddress: "000000000000000000000000cd7361ac3307d1c5a46b63086a90742ff44c63b3", Symbol: "raider", CoingeckoID: "crypto-raiders", Decimals: 18},
		{TokenChain: 5, TokenAddress: "000000000000000000000000d13cfd3133239a3c73a9e535a5c4dadee36b395c", Symbol: "vai", CoingeckoID: "vaiot", Decimals: 18},
		{TokenChain: 5, TokenAddress: "000000000000000000000000d5d86fc8d5c0ea1ac1ac5dfab6e529c9967a45e9", Symbol: "wrld", CoingeckoID: "nft-worlds", Decimals: 18},
		{TokenChain: 5, TokenAddress: "000000000000000000000000e580074a10360404af3abfe2d524d5806d993ea3", Symbol: "pay", CoingeckoID: "paybolt", Decimals: 18},
		{TokenChain: 5, TokenAddress: "000000000000000000000000e7a24ef0c5e95ffb0f6684b813a78f2a3ad7d171", Symbol: "am3crv", CoingeckoID: "curve-fi-amdai-amusdc-amusdt", Decimals: 18},
	}
}
