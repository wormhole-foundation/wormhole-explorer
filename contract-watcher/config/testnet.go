package config

import (
	"strings"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

var ETHEREUM_TESTNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDEthereum,
	Name:         "eth_goerli",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 8660321,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var POLYGON_TESTNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDPolygon,
	Name:         "polygon_mumbai",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 33151522,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
		strings.ToLower("0xc3D46e0266d95215589DE639cC4E93b79f88fc6C"): {
			{
				ID:   MethodIDReceiveTbtc,
				Name: MethodReceiveTbtc,
			},
		},
	},
}

var BSC_TESTNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDBSC,
	Name:         "bsc_testnet_chapel",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 28071327,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var FANTOM_TESTNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDFantom,
	Name:         "fantom_testnet",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 14524466,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var AVALANCHE_TESTNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDAvalanche,
	Name:         "avalanche_fuji",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 11014526,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var MOONBEAM_TESTNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDMoonbeam,
	Name:         "moonbeam",
	SizeBlocks:   50,
	WaitSeconds:  10,
	InitialBlock: 2097310,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var CELO_TESTNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDCelo,
	Name:         "celo",
	SizeBlocks:   50,
	WaitSeconds:  10,
	InitialBlock: 10625129,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var ARBITRUM_TESTNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDArbitrum,
	Name:         "arbitrum_goerli",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 15_470_418,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0xe3e0511EEbD87F08FbaE4486419cb5dFB06e1343"): {
			{
				ID:   MethodIDReceiveTbtc,
				Name: MethodReceiveTbtc,
			},
		},
	},
}

var OPTIMISM_TESTNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDOptimism,
	Name:         "optimism_goerli",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 7_973_025,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0xc3D46e0266d95215589DE639cC4E93b79f88fc6C"): {
			{
				ID:   MethodIDReceiveTbtc,
				Name: MethodReceiveTbtc,
			},
		},
	},
}

var BASE_TESTNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDBase,
	Name:         "base_goerli",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 902_385,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0xA31aa3FDb7aF7Db93d18DDA4e19F811342EDF780"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var BASE_SEPOLIA_TESTNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDBaseSepolia,
	Name:         "base_sepolia",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 3_415_420,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0x86F55A04690fd7815A3D802bD587e83eA888B239"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}
