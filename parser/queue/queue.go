package queue

import (
	"context"
	"fmt"

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

type ConsumerMessage struct {
	Data      *VaaEvent
	Ack       func()
	IsExpired func() bool
}

// ID get vaa ID.
func (v *VaaEvent) ID() string {
	return fmt.Sprintf("%d/%s/%d", v.ChainID, v.EmitterAddress, v.Sequence)
}

// VAAPushFunc is a function to push VAAEvent.
type VAAPushFunc func(context.Context, *VaaEvent) error

type VAAConsumeFunc func(context.Context) <-chan *ConsumerMessage
