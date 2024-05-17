package domain

import (
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/queue"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type Node struct {
	NodeName    string
	NodeAddress string
}

type NodeGovernorVaa struct {
	Node
	GovernorVaas map[string]GovernorVaa
}

type GovernorVaa struct {
	ChainID        sdk.ChainID
	EmitterAddress string
	Sequence       string
	TxHash         string
	ReleaseTime    time.Time
	Amount         uint64
}

// ConvertEventToGovernorVaa convert a event *queue.EventGovernorStatus to a *NodeGovernorVaa.
func ConvertEventToGovernorVaa(event *queue.EventGovernorStatus) *NodeGovernorVaa {

	// check if event is nil.
	if event == nil {
		return nil
	}

	// check if chains is empty.
	if len(event.Data.Chains) == 0 {
		return nil
	}

	governorVaas := make(map[string]GovernorVaa)
	for _, chain := range event.Data.Chains {
		for _, emitter := range chain.Emitters {
			for _, enqueuedVAA := range emitter.EnqueuedVaas {
				gs := GovernorVaa{
					ChainID:        sdk.ChainID(chain.ChainId),
					EmitterAddress: emitter.EmitterAddress,
					Sequence:       enqueuedVAA.Sequence,
					TxHash:         enqueuedVAA.TxHash,
					ReleaseTime:    time.Unix(int64(enqueuedVAA.ReleaseTime), 0),
					Amount:         enqueuedVAA.NotionalValue,
				}

				vaaID := fmt.Sprintf("%d/%s/%s", chain.ChainId, emitter.EmitterAddress, enqueuedVAA.Sequence)
				governorVaas[vaaID] = gs
			}
		}
	}

	return &NodeGovernorVaa{
		Node: Node{
			NodeName:    event.Data.NodeName,
			NodeAddress: event.Data.NodeAddress,
		},
		GovernorVaas: governorVaas,
	}
}
