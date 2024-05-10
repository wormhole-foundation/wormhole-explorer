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
