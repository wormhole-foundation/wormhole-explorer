package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/domain"
)

const (
	DeduplicateVaaEventType = "duplicated-vaa"
	GovernorStatusEventType = "governor-status"
)

// sqsEvent represents a event data from SQS.
type sqsEvent struct {
	MessageID string `json:"MessageId"`
	Message   string `json:"Message"`
}

type Event interface {
	EventDuplicateVaa | EventGovernorStatus
}

type EventDuplicateVaa struct {
	TrackID string       `json:"trackId"`
	Type    string       `json:"type"`
	Source  string       `json:"source"`
	Data    DuplicateVaa `json:"data"`
}

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

type EventGovernorStatus struct {
	TrackID string         `json:"trackId"`
	Type    string         `json:"type"`
	Source  string         `json:"source"`
	Data    GovernorStatus `json:"data"`
}

type GovernorStatus struct {
	NodeAddress string         `json:"nodeAddress"`
	NodeName    string         `json:"nodeName"`
	Counter     int64          `json:"counter"`
	Timestamp   int64          `json:"timestamp"`
	Chains      []*ChainStatus `json:"chains"`
}

type ChainStatus struct {
	ChainId                    uint32     `json:"chainId"`
	RemainingAvailableNotional uint64     `json:"remainingAvailableNotional"` // TODO Uint64
	Emitters                   []*Emitter `json:"emitters"`
}

type Emitter struct {
	EmitterAddress    string         `bson:"emitteraddress" json:"emitterAddress"`
	TotalEnqueuedVaas uint64         `bson:"totalenqueuedvaas" json:"totalEnqueuedVaas"`
	EnqueuedVaas      []*EnqueuedVAA `bson:"enqueuedvaas" json:"enqueuedVaas"`
}

type EnqueuedVAA struct {
	Sequence      string `bson:"sequence" json:"sequence"`
	ReleaseTime   uint32 `bson:"releasetime" json:"releaseTime"`
	NotionalValue uint64 `bson:"notionalvalue" json:"notionalValue"`
	TxHash        string `bson:"txhash" json:"txHash"`
}

func (e *EventGovernorStatus) ToMapGovernorStatus() map[string]domain.GovernorStatus {

	// check if chains is empty
	if len(e.Data.Chains) == 0 {
		return nil
	}

	// create a new map map[string]domain.GovernorStatus
	governorStatus := make(map[string]domain.GovernorStatus)

	// iterate over chains
	for _, chain := range e.Data.Chains {
		// iterate over emitters
		for _, emitter := range chain.Emitters {
			// iterate over enqueued vaas
			for _, enqueuedVAA := range emitter.EnqueuedVaas {
				// create a new GovernorStatus
				gs := domain.GovernorStatus{
					ChainID:        chain.ChainId,
					EmitterAddress: emitter.EmitterAddress,
					Sequence:       enqueuedVAA.Sequence,
					GovernorTxHash: enqueuedVAA.TxHash,
					ReleaseTime:    time.Unix(int64(enqueuedVAA.ReleaseTime), 0),
					Amount:         enqueuedVAA.NotionalValue,
				}

				vaaId := fmt.Sprintf("%d/%s/%d", chain.ChainId, emitter.EmitterAddress, enqueuedVAA.Sequence)
				// add GovernorStatus to the map
				governorStatus[vaaId] = gs
			}
		}
	}
	return governorStatus
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
