package event

import (
	"context"
	"time"
)

type DuplicateVaa struct {
	VaaID            string     `json:"vaaId"`
	ChainID          uint16     `json:"chainId"`
	Version          uint8      `json:"version"`
	GuardianSetIndex uint32     `json:"guardianSetIndex"`
	Vaa              []byte     `json:"vaas"`
	Digest           string     `json:"digest"`
	ConsistencyLevel uint8      `json:"consistencyLevel"`
	Timestamp        *time.Time `json:"timestamp"`
}

type GovernorStatus struct {
	NodeAddress string `json:"nodeAddress"`
	NodeName    string `json:"nodeName"`
	Counter     int64  `json:"counter"`
	Timestamp   int64  `json:"timestamp"`
	Chains      any    `json:"chains"`
}

type GovernorConfig struct {
	NodeAddress string `json:"nodeAddress"`
	NodeName    string `json:"nodeName"`
	Counter     int64  `json:"counter"`
	Timestamp   int64  `json:"timestamp"`
	Chains      any    `json:"chains"`
}

type Vaa struct {
	ID               string    `json:"id"`
	VaaID            string    `json:"vaaId"`
	EmitterChainID   uint16    `json:"emitterChainId"`
	EmitterAddress   string    `json:"emitterAddress"`
	Sequence         uint64    `json:"sequence"`
	Version          uint8     `json:"version"`
	GuardianSetIndex uint32    `json:"guardianSetIndex"`
	Raw              []byte    `json:"raw"`
	Timestamp        time.Time `json:"timestamp"`
}

type event struct {
	TrackID string `json:"trackId"`
	Type    string `json:"type"`
	Source  string `json:"source"`
	Data    any    `json:"data"`
}

type EventDispatcher interface {
	NewVaa(ctx context.Context, vaa Vaa) error
	NewDuplicateVaa(ctx context.Context, e DuplicateVaa) error
	NewGovernorStatus(ctx context.Context, e GovernorStatus) error
	NewGovernorConfig(ctx context.Context, e GovernorConfig) error
}
