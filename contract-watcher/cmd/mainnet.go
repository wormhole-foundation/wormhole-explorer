package main

import "github.com/wormhole-foundation/wormhole/sdk/vaa"

var ETHEREUM_MAINNET = watcherBlockchain{
	chainID:      vaa.ChainIDEthereum,
	name:         "eth",
	address:      "0x3ee18B2214AFF97000D974cf647E7C347E8fa585",
	sizeBlocks:   100,
	waitSeconds:  10,
	initialBlock: 16820790,
}

var POLYGON_MAINNET = watcherBlockchain{
	chainID:      vaa.ChainIDPolygon,
	name:         "polygon",
	address:      "0x5a58505a96D1dbf8dF91cB21B54419FC36e93fdE",
	sizeBlocks:   100,
	waitSeconds:  10,
	initialBlock: 40307020,
}

var BSC_MAINNET = watcherBlockchain{
	chainID:      vaa.ChainIDBSC,
	name:         "bsc",
	address:      "0xB6F6D86a8f9879A9c87f643768d9efc38c1Da6E7",
	sizeBlocks:   100,
	waitSeconds:  10,
	initialBlock: 26436320,
}

var FANTOM_MAINNET = watcherBlockchain{
	chainID:      vaa.ChainIDFantom,
	name:         "fantom",
	address:      "0x7C9Fc5741288cDFdD83CeB07f3ea7e22618D79D2",
	sizeBlocks:   100,
	waitSeconds:  10,
	initialBlock: 57525624,
}

var SOLANA_MAINNET = watcherBlockchain{
	chainID:      vaa.ChainIDSolana,
	name:         "solana",
	address:      "wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb",
	sizeBlocks:   50,
	waitSeconds:  10,
	initialBlock: 183675278,
}

var TERRA_MAINNET = watcherBlockchain{
	chainID:      vaa.ChainIDTerra,
	name:         "terra",
	address:      "terra10nmmwe8r3g99a9newtqa7a75xfgs2e8z87r2sf",
	sizeBlocks:   0,
	waitSeconds:  10,
	initialBlock: 3911168,
}

var AVALANCHE_MAINNET = watcherBlockchain{
	chainID:      vaa.ChainIDAvalanche,
	name:         "avalanche",
	address:      "0x0e082F06FF657D94310cB8cE8B0D9a04541d8052",
	sizeBlocks:   100,
	waitSeconds:  10,
	initialBlock: 8237181,
}

var APTOS_MAINNET = watcherBlockchain{
	chainID:      vaa.ChainIDAptos,
	name:         "aptos",
	address:      "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f",
	sizeBlocks:   50,
	waitSeconds:  10,
	initialBlock: 1094430,
}

var OASIS_MAINNET = watcherBlockchain{
	chainID:      vaa.ChainIDOasis,
	name:         "oasis",
	address:      "0x5848C791e09901b40A9Ef749f2a6735b418d7564",
	sizeBlocks:   50,
	waitSeconds:  10,
	initialBlock: 1762,
}

var MOONBEAM_MAINNET = watcherBlockchain{
	chainID:      vaa.ChainIDMoonbeam,
	name:         "moonbeam",
	address:      "0xb1731c586ca89a23809861c6103f0b96b3f57d92",
	sizeBlocks:   50,
	waitSeconds:  10,
	initialBlock: 1853330,
}
