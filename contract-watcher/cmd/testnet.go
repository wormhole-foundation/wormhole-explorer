package main

import "github.com/wormhole-foundation/wormhole/sdk/vaa"

var ETHEREUM_TESTNET = watcherBlockchain{
	chainID:      vaa.ChainIDEthereum,
	name:         "eth_goerli",
	address:      "0xF890982f9310df57d00f659cf4fd87e65adEd8d7",
	sizeBlocks:   100,
	waitSeconds:  10,
	initialBlock: 8660321,
}

var POLYGON_TESTNET = watcherBlockchain{
	chainID:      vaa.ChainIDPolygon,
	name:         "polygon_mumbai",
	address:      "0x377D55a7928c046E18eEbb61977e714d2a76472a",
	sizeBlocks:   100,
	waitSeconds:  10,
	initialBlock: 33151522,
}

var BSC_TESTNET = watcherBlockchain{
	chainID:      vaa.ChainIDBSC,
	name:         "bsc_testnet_chapel",
	address:      "0x9dcF9D205C9De35334D646BeE44b2D2859712A09",
	sizeBlocks:   100,
	waitSeconds:  10,
	initialBlock: 28071327,
}

var FANTOM_TESTNET = watcherBlockchain{
	chainID:      vaa.ChainIDFantom,
	name:         "fantom_testnet",
	address:      "0x599CEa2204B4FaECd584Ab1F2b6aCA137a0afbE8",
	sizeBlocks:   100,
	waitSeconds:  10,
	initialBlock: 14524466,
}

var SOLANA_TESTNET = watcherBlockchain{
	chainID:      vaa.ChainIDSolana,
	name:         "solana",
	address:      "DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe",
	sizeBlocks:   10,
	waitSeconds:  10,
	initialBlock: 16820790,
}

var AVALANCHE_TESTNET = watcherBlockchain{
	chainID:      vaa.ChainIDAvalanche,
	name:         "avalanche_fuji",
	address:      "0x61E44E506Ca5659E6c0bba9b678586fA2d729756",
	sizeBlocks:   100,
	waitSeconds:  10,
	initialBlock: 11014526,
}

var APTOS_TESTNET = watcherBlockchain{
	chainID:      vaa.ChainIDAptos,
	name:         "aptos",
	address:      "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f",
	sizeBlocks:   50,
	waitSeconds:  10,
	initialBlock: 21522262,
}

var OASIS_TESTNET = watcherBlockchain{
	chainID:      vaa.ChainIDOasis,
	name:         "oasis",
	address:      "0x88d8004A9BdbfD9D28090A02010C19897a29605c",
	sizeBlocks:   50,
	waitSeconds:  10,
	initialBlock: 130400,
}

var MOONBEAM_TESTNET = watcherBlockchain{
	chainID:      vaa.ChainIDMoonbeam,
	name:         "moonbeam",
	address:      "0xbc976D4b9D57E57c3cA52e1Fd136C45FF7955A96",
	sizeBlocks:   50,
	waitSeconds:  10,
	initialBlock: 2097310,
}
