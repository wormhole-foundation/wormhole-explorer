package queue

import (
	"context"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// VaaEvent represents a vaa data to be handle by the pipeline.
type VaaEvent struct {
	ChainID          vaa.ChainID `json:"chainId"`
	EmitterAddress   vaa.Address `json:"emitter"`
	Sequence         uint64      `json:"sequence"`
	Vaa              []byte      `json:"vaa"`
	ParserFunctionID string      `json:"parserFunctionID"`
}

// VAAPushFunc is a function to push VAAEvent.
type VAAPushFunc func(context.Context, *VaaEvent) error
