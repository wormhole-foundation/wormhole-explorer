package topic

import (
	"context"
	"encoding/json"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"time"
)

// Event represents a vaa data to be handle by the pipeline.
type Event struct {
	ID               string      `json:"id"`
	ChainID          sdk.ChainID `json:"emitterChain"`
	EmitterAddress   string      `json:"emitterAddr"`
	Sequence         string      `json:"sequence"`
	GuardianSetIndex uint32      `json:"guardianSetIndex"`
	Vaa              []byte      `json:"vaas"`
	IndexedAt        time.Time   `json:"indexedAt"`
	Timestamp        *time.Time  `json:"timestamp"`
	UpdatedAt        *time.Time  `json:"updatedAt"`
	TxHash           string      `json:"txHash"`
	Version          uint16      `json:"version"`
	Revision         uint16      `json:"revision"`
	Digest           string      `json:"digest"`
	Overwrite        bool        `json:"overwrite"`
}

type SnsMessage interface {
	GetGroupID() string
	GetDeduplicationID() string
	GetChainID() sdk.ChainID
	Body() ([]byte, error)
}

func (e *Event) GetGroupID() string {
	return e.ID
}

func (e *Event) GetDeduplicationID() string {
	return e.ID
}

func (e *Event) GetChainID() sdk.ChainID {
	return e.ChainID
}

func (e *Event) Body() ([]byte, error) {
	return json.Marshal(e)
}

// PushFunc is a function to push VAAEvent.
type PushFunc func(context.Context, SnsMessage) error
