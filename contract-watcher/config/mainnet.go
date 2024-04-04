package config

import (
	"strings"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

var ETHEREUM_MAINNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDEthereum,
	Name:         "eth",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 16820790,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var POLYGON_MAINNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDPolygon,
	Name:         "polygon",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 40307020,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
		strings.ToLower("0x09959798B95d00a3183d20FaC298E4594E599eab"): {
			{
				ID:   MethodIDReceiveTbtc,
				Name: MethodReceiveTbtc,
			},
		},
	},
}

var BSC_MAINNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDBSC,
	Name:         "bsc",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 26436320,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var FANTOM_MAINNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDFantom,
	Name:         "fantom",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 57525624,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var TERRA_MAINNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDTerra,
	Name:         "terra",
	Address:      "terra10nmmwe8r3g99a9newtqa7a75xfgs2e8z87r2sf",
	SizeBlocks:   0,
	WaitSeconds:  10,
	InitialBlock: 3911168,
}

var AVALANCHE_MAINNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDAvalanche,
	Name:         "avalanche",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 8237181,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var MOONBEAM_MAINNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDMoonbeam,
	Name:         "moonbeam",
	SizeBlocks:   50,
	WaitSeconds:  10,
	InitialBlock: 1853330,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var CELO_MAINNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDCelo,
	Name:         "celo",
	SizeBlocks:   50,
	WaitSeconds:  10,
	InitialBlock: 12947239,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var ARBITRUM_MAINNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDArbitrum,
	Name:         "arbitrum",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 75_577_070,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0x1293a54e160D1cd7075487898d65266081A15458"): {
			{
				ID:   MethodIDReceiveTbtc,
				Name: MethodReceiveTbtc,
			},
		},
	},
}

var OPTIMISM_MAINNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDOptimism,
	Name:         "optimism",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 89_900_107,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0x1293a54e160D1cd7075487898d65266081A15458"): {
			{
				ID:   MethodIDReceiveTbtc,
				Name: MethodReceiveTbtc,
			},
		},
	},
}
