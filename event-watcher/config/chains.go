package config

type EvmParams struct {
	StartingBlock   uint64
	ContractAddress string
	Topic           string
}

var ETHEREUM_MAINNET = EvmParams{
	StartingBlock:   12_959_638,
	ContractAddress: "0x98f3c9e6e3face36baad05fe09d375ef1464288b",
	Topic:           "0x6eb224fb001ed210e379b335e35efe88672a8ce935d981a6896b27ffdf52a3b2",
}

var ETHEREUM_GOERLI = EvmParams{
	StartingBlock:   5_896_171,
	ContractAddress: "0x706abc4e45d419950511e474c7b9ed348a4a716c",
	Topic:           "0x6eb224fb001ed210e379b335e35efe88672a8ce935d981a6896b27ffdf52a3b2",
}
