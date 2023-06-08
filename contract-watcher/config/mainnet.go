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
		strings.ToLower("0x3ee18B2214AFF97000D974cf647E7C347E8fa585"): {
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
		strings.ToLower("0x5a58505a96D1dbf8dF91cB21B54419FC36e93fdE"): {
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
		strings.ToLower("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
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
		strings.ToLower("0xB6F6D86a8f9879A9c87f643768d9efc38c1Da6E7"): {
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
		strings.ToLower("0x7C9Fc5741288cDFdD83CeB07f3ea7e22618D79D2"): {
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
		strings.ToLower("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var SOLANA_MAINNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDSolana,
	Name:         "solana",
	Address:      "wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb",
	SizeBlocks:   50,
	WaitSeconds:  10,
	InitialBlock: 183675278,
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
		strings.ToLower("0x0e082F06FF657D94310cB8cE8B0D9a04541d8052"): {
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
		strings.ToLower("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}

var APTOS_MAINNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDAptos,
	Name:         "aptos",
	Address:      "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f",
	SizeBlocks:   50,
	WaitSeconds:  10,
	InitialBlock: 1094430,
}

var OASIS_MAINNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDOasis,
	Name:         "oasis",
	SizeBlocks:   50,
	WaitSeconds:  10,
	InitialBlock: 1762,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0x5848C791e09901b40A9Ef749f2a6735b418d7564"): {
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

var MOONBEAM_MAINNET = WatcherBlockchainAddresses{
	ChainID:      vaa.ChainIDMoonbeam,
	Name:         "moonbeam",
	SizeBlocks:   50,
	WaitSeconds:  10,
	InitialBlock: 1853330,
	MethodsByAddress: map[string][]BlockchainMethod{
		strings.ToLower("0xb1731c586ca89a23809861c6103f0b96b3f57d92"): {
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
		strings.ToLower("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca"): {
			{
				ID:   MetehodIDCompleteTransferWithRelay,
				Name: MetehodCompleteTransferWithRelay,
			},
		},
	},
}
