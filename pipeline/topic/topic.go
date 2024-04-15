package topic

import (
	"context"
	"time"
)

// Event represents a vaa data to be handle by the pipeline.
type Event struct {
	ID               string     `json:"id"`
	ChainID          uint16     `json:"emitterChain"`
	EmitterAddress   string     `json:"emitterAddr"`
	Sequence         string     `json:"sequence"`
	GuardianSetIndex uint32     `json:"guardianSetIndex"`
	Vaa              []byte     `json:"vaas"`
	IndexedAt        time.Time  `json:"indexedAt"`
	Timestamp        *time.Time `json:"timestamp"`
	UpdatedAt        *time.Time `json:"updatedAt"`
	TxHash           string     `json:"txHash"`
	Version          uint16     `json:"version"`
	Revision         uint16     `json:"revision"`
	Hash             []byte     `json:"hash"`
}

// PushFunc is a function to push VAAEvent.
type PushFunc func(context.Context, *Event) error
