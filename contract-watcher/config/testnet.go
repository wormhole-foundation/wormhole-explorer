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
		strings.ToLower("0xF890982f9310df57d00f659cf4fd87e65adEd8d7"): {
			{
				ID:   MethodIDCompleteTransfer,
				Name: MethodCompleteTransfer,
			},
			{
				ID:   MethodIDCompleteAndUnwrapETH,
				Name: MethodCompleteAndUnwrapETH,
			},
			{
				ID:   MethodIDCreateWrapped,
				Name: MethodCreateWrapped,
			},
			{
				ID:   MethodIDUpdateWrapped,
				Name: MethodUpdateWrapped,
			},
		},
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
		strings.ToLower("0x377D55a7928c046E18eEbb61977e714d2a76472a"): {
			{
				ID:   MethodIDCompleteTransfer,
				Name: MethodCompleteTransfer,
			},
			{
				ID:   MethodIDCompleteAndUnwrapETH,
				Name: MethodCompleteAndUnwrapETH,
			},
			{
				ID:   MethodIDCreateWrapped,
				Name: MethodCreateWrapped,
			},
			{
				ID:   MethodIDUpdateWrapped,
				Name: MethodUpdateWrapped,
			},
		},
		strings.ToLower("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
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
		strings.ToLower("0x9dcF9D205C9De35334D646BeE44b2D2859712A09"): {
			{
				ID:   MethodIDCompleteTransfer,
				Name: MethodCompleteTransfer,
			},
			{
				ID:   MethodIDCompleteAndUnwrapETH,
				Name: MethodCompleteAndUnwrapETH,
			},
			{
				ID:   MethodIDCreateWrapped,
				Name: MethodCreateWrapped,
			},
			{
				ID:   MethodIDUpdateWrapped,
				Name: MethodUpdateWrapped,
			},
		},
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
		strings.ToLower("0x599CEa2204B4FaECd584Ab1F2b6aCA137a0afbE8"): {
			{
				ID:   MethodIDCompleteTransfer,
				Name: MethodCompleteTransfer,
			},
			{
				ID:   MethodIDCompleteAndUnwrapETH,
				Name: MethodCompleteAndUnwrapETH,
			},
			{
				ID:   MethodIDCreateWrapped,
				Name: MethodCreateWrapped,
			},
			{
				ID:   MethodIDUpdateWrapped,
				Name: MethodUpdateWrapped,
			},
		},
		strings.ToLower("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var SOLANA_TESTNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDSolana,
	Name:         "solana",
	Address:      "DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe",
	SizeBlocks:   10,
	WaitSeconds:  10,
	InitialBlock: 16820790,
}

var AVALANCHE_TESTNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDAvalanche,
	Name:         "avalanche_fuji",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 11014526,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0x61E44E506Ca5659E6c0bba9b678586fA2d729756"): {
			{
				ID:   MethodIDCompleteTransfer,
				Name: MethodCompleteTransfer,
			},
			{
				ID:   MethodIDCompleteAndUnwrapETH,
				Name: MethodCompleteAndUnwrapETH,
			},
			{
				ID:   MethodIDCreateWrapped,
				Name: MethodCreateWrapped,
			},
			{
				ID:   MethodIDUpdateWrapped,
				Name: MethodUpdateWrapped,
			},
		},
		strings.ToLower("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var APTOS_TESTNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDAptos,
	Name:         "aptos",
	Address:      "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f",
	SizeBlocks:   50,
	WaitSeconds:  10,
	InitialBlock: 21522262,
}

var OASIS_TESTNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDOasis,
	Name:         "oasis",
	SizeBlocks:   50,
	WaitSeconds:  10,
	InitialBlock: 130400,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0x88d8004A9BdbfD9D28090A02010C19897a29605c"): {
			{
				ID:   MethodIDCompleteTransfer,
				Name: MethodCompleteTransfer,
			},
			{
				ID:   MethodIDCompleteAndUnwrapETH,
				Name: MethodCompleteAndUnwrapETH,
			},
			{
				ID:   MethodIDCreateWrapped,
				Name: MethodCreateWrapped,
			},
			{
				ID:   MethodIDUpdateWrapped,
				Name: MethodUpdateWrapped,
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
		strings.ToLower("0xbc976D4b9D57E57c3cA52e1Fd136C45FF7955A96"): {
			{
				ID:   MethodIDCompleteTransfer,
				Name: MethodCompleteTransfer,
			},
			{
				ID:   MethodIDCompleteAndUnwrapETH,
				Name: MethodCompleteAndUnwrapETH,
			},
			{
				ID:   MethodIDCreateWrapped,
				Name: MethodCreateWrapped,
			},
			{
				ID:   MethodIDUpdateWrapped,
				Name: MethodUpdateWrapped,
			},
		},
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
		strings.ToLower("0x05ca6037eC51F8b712eD2E6Fa72219FEaE74E153"): {
			{
				ID:   MethodIDCompleteTransfer,
				Name: MethodCompleteTransfer,
			},
			{
				ID:   MethodIDCompleteAndUnwrapETH,
				Name: MethodCompleteAndUnwrapETH,
			},
			{
				ID:   MethodIDCreateWrapped,
				Name: MethodCreateWrapped,
			},
			{
				ID:   MethodIDUpdateWrapped,
				Name: MethodUpdateWrapped,
			},
		},
		strings.ToLower("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}
