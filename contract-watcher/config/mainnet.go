package config

import (
	"strings"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

var TERRA_MAINNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDTerra,
	Name:         "terra",
	Address:      "terra10nmmwe8r3g99a9newtqa7a75xfgs2e8z87r2sf",
	SizeBlocks:   0,
	WaitSeconds:  10,
	InitialBlock: 3911168,
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
