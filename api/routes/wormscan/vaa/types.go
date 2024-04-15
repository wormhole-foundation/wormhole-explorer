package vaa

import (
	"time"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type DuplicateVaaResponse struct {
	ID                string      `json:"id"`
	Version           uint8       `json:"version"`
	EmitterChain      vaa.ChainID `json:"emitterChain"`
	EmitterAddr       string      `json:"emitterAddr"`
	EmitterNativeAddr string      `json:"emitterNativeAddr,omitempty"`
	Sequence          string      `json:"sequence"`
	GuardianSetIndex  uint32      `json:"guardianSetIndex"`
	Vaa               []byte      `json:"vaa"`
	Timestamp         *time.Time  `json:"timestamp"`
	UpdatedAt         *time.Time  `json:"updatedAt"`
	IndexedAt         *time.Time  `json:"indexedAt"`
	Digest            string      `json:"digest"`
}
