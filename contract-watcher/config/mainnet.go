package config

import "github.com/wormhole-foundation/wormhole/sdk/vaa"

var ETHEREUM_MAINNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDEthereum,
	Name:         "eth",
	Address:      "0x3ee18B2214AFF97000D974cf647E7C347E8fa585",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 16820790,
}

var POLYGON_MAINNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDPolygon,
	Name:         "polygon",
	Address:      "0x5a58505a96D1dbf8dF91cB21B54419FC36e93fdE",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 40307020,
}

var BSC_MAINNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDBSC,
	Name:         "bsc",
	Address:      "0xB6F6D86a8f9879A9c87f643768d9efc38c1Da6E7",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 26436320,
}

var FANTOM_MAINNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDFantom,
	Name:         "fantom",
	Address:      "0x7C9Fc5741288cDFdD83CeB07f3ea7e22618D79D2",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 57525624,
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

var AVALANCHE_MAINNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDAvalanche,
	Name:         "avalanche",
	Address:      "0x0e082F06FF657D94310cB8cE8B0D9a04541d8052",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 8237181,
}

var APTOS_MAINNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDAptos,
	Name:         "aptos",
	Address:      "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f",
	SizeBlocks:   50,
	WaitSeconds:  10,
	InitialBlock: 1094430,
}

var OASIS_MAINNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDOasis,
	Name:         "oasis",
	Address:      "0x5848C791e09901b40A9Ef749f2a6735b418d7564",
	SizeBlocks:   50,
	WaitSeconds:  10,
	InitialBlock: 1762,
}

var MOONBEAM_MAINNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDMoonbeam,
	Name:         "moonbeam",
	Address:      "0xb1731c586ca89a23809861c6103f0b96b3f57d92",
	SizeBlocks:   50,
	WaitSeconds:  10,
	InitialBlock: 1853330,
}
