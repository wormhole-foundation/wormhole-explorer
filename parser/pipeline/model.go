package pipeline

import (
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// VaaEvent represents a vaa data to be handle by the pipeline.
type VaaEvent struct {
	ChainID  vaa.ChainID `json:"chainId"`
	Emitter  string      `json:"emitter"`
	Sequence string      `json:"sequence"`
	Vaa      []byte      `json:"vaa"`
}
