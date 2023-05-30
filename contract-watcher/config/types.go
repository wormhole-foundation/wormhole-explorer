package config

import "github.com/wormhole-foundation/wormhole/sdk/vaa"

type WatcherBlockchain struct {
	ChainID      vaa.ChainID
	Name         string
	Address      string
	SizeBlocks   uint8
	WaitSeconds  uint16
	InitialBlock int64
}
