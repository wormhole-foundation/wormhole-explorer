package domain

import (
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	"github.com/wormhole-foundation/wormhole-explorer/fly-event-processor/queue"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type Node struct {
	Name    string
	Address string
}

type NodeGovernorVaa struct {
	Node
	GovernorVaas map[string]GovernorVaa
}

type GovernorVaa struct {
	ID             string
	ChainID        sdk.ChainID
	EmitterAddress string
	Sequence       string
	TxHash         string
	ReleaseTime    time.Time
	Amount         uint64
}

// ConvertEventToGovernorVaa convert a event *queue.EventGovernorStatus to a *NodeGovernorVaa.
func ConvertEventToGovernorVaa(event *queue.GovernorStatus) *NodeGovernorVaa {

	// check if event is nil.
	if event == nil {
		return nil
	}

	// check if chains is empty.
	if len(event.Chains) == 0 {
		return nil
	}

	governorVaas := make(map[string]GovernorVaa)
	for _, chain := range event.Chains {
		for _, emitter := range chain.Emitters {
			for _, enqueuedVAA := range emitter.EnqueuedVaas {

				normalizeEmitter := utils.NormalizeHex(emitter.EmitterAddress)
				normalizeTxHash := utils.NormalizeHex(enqueuedVAA.TxHash)
				vaaID := fmt.Sprintf("%d/%s/%s",
					chain.ChainId,
					normalizeEmitter,
					enqueuedVAA.Sequence)

				gs := GovernorVaa{
					ID:             vaaID,
					ChainID:        sdk.ChainID(chain.ChainId),
					EmitterAddress: normalizeEmitter,
					Sequence:       enqueuedVAA.Sequence,
					TxHash:         normalizeTxHash,
					ReleaseTime:    time.Unix(int64(enqueuedVAA.ReleaseTime), 0),
					Amount:         enqueuedVAA.NotionalValue,
				}

				governorVaas[vaaID] = gs
			}
		}
	}

	return &NodeGovernorVaa{
		Node: Node{
			Name:    event.NodeName,
			Address: event.NodeAddress,
		},
		GovernorVaas: governorVaas,
	}
}
