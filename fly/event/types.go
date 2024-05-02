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
	//Chains    []*storage.ChainGovernorStatusChain `json:"chains"`
}

type event struct {
	TrackID string `json:"trackId"`
	Type    string `json:"type"`
	Source  string `json:"source"`
	Data    any    `json:"data"`
}

type EventDispatcher interface {
	NewDuplicateVaa(ctx context.Context, e DuplicateVaa) error
	NewGovernorStatus(ctx context.Context, e GovernorStatus) error
}
