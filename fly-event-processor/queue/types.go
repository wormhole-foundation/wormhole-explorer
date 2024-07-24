package queue

import (
	"context"
	"encoding/json"
	"time"
)

const (
	DeduplicateVaaEventType = "duplicated-vaa"
	GovernorStatusEventType = "governor-status"
	GovernorConfigEventType = "governor-config"
)

// sqsEvent represents a event data from SQS.
type sqsEvent struct {
	MessageID string `json:"MessageId"`
	Message   string `json:"Message"`
}

// Event represents a event data.
type Event interface {
	EventDuplicateVaa | EventGovernor
}

// EventDuplicateVaa defition.
type EventDuplicateVaa struct {
	TrackID string       `json:"trackId"`
	Type    string       `json:"type"`
	Source  string       `json:"source"`
	Data    DuplicateVaa `json:"data"`
}

// DuplicateVaa defition.
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

type EventGovernor struct {
	TrackID string          `json:"trackId"`
	Type    string          `json:"type"`
	Source  string          `json:"source"`
	Data    json.RawMessage `json:"data"`
}

// type EventGovernorConfig struct {
// 	TrackID string         `json:"trackId"`
// 	Type    string         `json:"type"`
// 	Source  string         `json:"source"`
// 	Data    GovernorConfig `json:"data"`
// }

type GovernorConfig struct {
	NodeAddress string         `json:"nodeAddress"`
	NodeName    string         `json:"nodeName"`
	Counter     int64          `json:"counter"`
	Timestamp   int64          `json:"timestamp"`
	Chains      []*ChainConfig `json:"chains"`
}

type ChainConfig struct {
	ChainId            uint16 `json:"chainId"`
	NotionalLimit      uint64 `json:"notionalLimit"`
	BigTransactionSize uint64 `json:"bigTransactionSize"`
}

// EventGovernorStatus defition.
// type EventGovernorStatus struct {
// 	TrackID string         `json:"trackId"`
// 	Type    string         `json:"type"`
// 	Source  string         `json:"source"`
// 	Data    GovernorStatus `json:"data"`
// }

// GovernorStatus defition.
type GovernorStatus struct {
	NodeAddress string         `json:"nodeAddress"`
	NodeName    string         `json:"nodeName"`
	Counter     int64          `json:"counter"`
	Timestamp   int64          `json:"timestamp"`
	Chains      []*ChainStatus `json:"chains"`
}

// ChainStatus defition.
type ChainStatus struct {
	ChainId                    uint32     `json:"chainId"`
	RemainingAvailableNotional uint64     `json:"remainingAvailableNotional"`
	Emitters                   []*Emitter `json:"emitters"`
}

// Emitter defition.
type Emitter struct {
	EmitterAddress    string         `bson:"emitteraddress" json:"emitterAddress"`
	TotalEnqueuedVaas uint64         `bson:"totalenqueuedvaas" json:"totalEnqueuedVaas"`
	EnqueuedVaas      []*EnqueuedVAA `bson:"enqueuedvaas" json:"enqueuedVaas"`
}

// EnqueuedVAA defition.
type EnqueuedVAA struct {
	Sequence      string `bson:"sequence" json:"sequence"`
	ReleaseTime   uint64 `bson:"releasetime" json:"releaseTime"`
	NotionalValue uint64 `bson:"notionalvalue" json:"notionalValue"`
	TxHash        string `bson:"txhash" json:"txHash"`
}

// ConsumerMessage defition.
type ConsumerMessage[T any] interface {
	Retry() uint8
	Data() T
	Done()
	Failed()
	IsExpired() bool
}

// ConsumeFunc is a function to consume Event.
type ConsumeFunc[T any] func(context.Context) <-chan ConsumerMessage[T]
