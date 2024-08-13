package queue

import (
	"encoding/json"
	"time"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// VaaEvent represents a vaa data to be handled by the pipeline.
type VaaEvent struct {
	TrackID   string    `json:"trackId"`
	Source    string    `json:"source"`
	Event     string    `json:"event"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	SignedVaa SignedVaa `json:"data"`
}

type SignedVaa struct {
	ID               string    `json:"id"`
	VaaID            string    `json:"vaaId"`
	EmitterChain     uint16    `json:"emitterChain"`
	EmitterAddress   string    `json:"emitterAddress"`
	Sequence         uint64    `json:"sequence"`
	GuardianSetIndex uint32    `json:"guardianSetIndex"`
	Timestamp        time.Time `json:"timestamp"`
	Vaa              []byte    `json:"vaa"`
	TxHash           string    `json:"txHash"`
	Version          int       `json:"version"`
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
			TrackID:          vaaEvent.TrackID,
			Source:           vaaEvent.Source,
			ID:               vaaEvent.SignedVaa.ID,
			VaaID:            vaaEvent.SignedVaa.VaaID,
			EmitterChainID:   sdk.ChainID(vaaEvent.SignedVaa.EmitterChain),
			EmitterAddress:   vaaEvent.SignedVaa.EmitterAddress,
			Sequence:         vaaEvent.SignedVaa.Sequence,
			GuardianSetIndex: vaaEvent.SignedVaa.GuardianSetIndex,
			Timestamp:        vaaEvent.SignedVaa.Timestamp,
			Vaa:              vaaEvent.SignedVaa.Vaa,
			Version:          vaaEvent.SignedVaa.Version,
		}, nil
	}
}
