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

var BASE_MAINNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDBase,
	Name:         "base",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 1_422_314,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0x8d2de8d2f73F1F4cAB472AC9A881C9b123C79627"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}
