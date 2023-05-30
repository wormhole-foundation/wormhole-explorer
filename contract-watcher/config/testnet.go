package config

import "github.com/wormhole-foundation/wormhole/sdk/vaa"

var ETHEREUM_TESTNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDEthereum,
	Name:         "eth_goerli",
	Address:      "0xF890982f9310df57d00f659cf4fd87e65adEd8d7",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 8660321,
}

var POLYGON_TESTNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDPolygon,
	Name:         "polygon_mumbai",
	Address:      "0x377D55a7928c046E18eEbb61977e714d2a76472a",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 33151522,
}

var BSC_TESTNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDBSC,
	Name:         "bsc_testnet_chapel",
	Address:      "0x9dcF9D205C9De35334D646BeE44b2D2859712A09",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 28071327,
}

var FANTOM_TESTNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDFantom,
	Name:         "fantom_testnet",
	Address:      "0x599CEa2204B4FaECd584Ab1F2b6aCA137a0afbE8",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 14524466,
}

var SOLANA_TESTNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDSolana,
	Name:         "solana",
	Address:      "DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe",
	SizeBlocks:   10,
	WaitSeconds:  10,
	InitialBlock: 16820790,
}

var AVALANCHE_TESTNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDAvalanche,
	Name:         "avalanche_fuji",
	Address:      "0x61E44E506Ca5659E6c0bba9b678586fA2d729756",
	SizeBlocks:   100,
	WaitSeconds:  10,
	InitialBlock: 11014526,
}

var APTOS_TESTNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDAptos,
	Name:         "aptos",
	Address:      "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f",
	SizeBlocks:   50,
	WaitSeconds:  10,
	InitialBlock: 21522262,
}

var OASIS_TESTNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDOasis,
	Name:         "oasis",
	Address:      "0x88d8004A9BdbfD9D28090A02010C19897a29605c",
	SizeBlocks:   50,
	WaitSeconds:  10,
	InitialBlock: 130400,
}

var MOONBEAM_TESTNET = WatcherBlockchain{
	ChainID:      vaa.ChainIDMoonbeam,
	Name:         "moonbeam",
	Address:      "0xbc976D4b9D57E57c3cA52e1Fd136C45FF7955A96",
	SizeBlocks:   50,
	WaitSeconds:  10,
	InitialBlock: 2097310,
}
