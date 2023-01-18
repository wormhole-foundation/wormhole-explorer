package queue

import (
	"context"
	"fmt"
)

// VaaEvent represents a vaa data to be handle by the pipeline.
type VaaEvent struct {
	ChainID        uint16 `json:"chainId"`
	EmitterAddress string `json:"emitter"`
	Sequence       uint64 `json:"sequence"`
	Vaa            []byte `json:"vaa"`
}

// ConsumerMessage defition.
type ConsumerMessage struct {
	Data      *VaaEvent
	Ack       func()
	IsExpired func() bool
}

// ID get vaa ID (chainID/emiiterAddress/sequence)
func (v *VaaEvent) ID() string {
	return fmt.Sprintf("%d/%s/%d", v.ChainID, v.EmitterAddress, v.Sequence)
}

// VAAPushFunc is a function to push VAAEvent.
type VAAPushFunc func(context.Context, *VaaEvent) error

// VAAConsumeFunc is a function to consume VAAEvent.
type VAAConsumeFunc func(context.Context) <-chan *ConsumerMessage
