package queue

import (
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// VaaEvent represents a vaa data to be handled by the pipeline.
type VaaEvent struct {
	ID               string      `json:"id"`
	VaaID            string      `json:"vaaId"`
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
	Overwrite        bool        `json:"overwrite"`
}

// NewVaaConverter converts a message from a VAAEvent.
func NewVaaConverter(_ *zap.Logger) ConverterFunc {

	return func(msg string) (*Event, error) {
		// unmarshal message to vaaEvent
		var vaaEvent VaaEvent
		err := json.Unmarshal([]byte(msg), &vaaEvent)
		if err != nil {
			return nil, err
		}
		return &Event{
			Source:         "fly",
			TrackID:        fmt.Sprintf("fly-%s", vaaEvent.ID),
			Type:           SourceChainEvent,
			ID:             vaaEvent.ID,
			VaaId:          vaaEvent.VaaID,
			ChainID:        vaaEvent.ChainID,
			EmitterAddress: vaaEvent.EmitterAddress,
			Sequence:       vaaEvent.Sequence,
			Timestamp:      vaaEvent.Timestamp,
			Vaa:            vaaEvent.Vaa,
			IsVaaSigned:    true,
		}, nil
	}
}
