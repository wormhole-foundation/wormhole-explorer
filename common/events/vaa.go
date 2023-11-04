package events

import (
	"fmt"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

func CreateUnsignedVAA(plm *LogMessagePublished) (*sdk.VAA, error) {

	address, err := sdk.StringToAddress(plm.EmitterAddress)
	if err != nil {
		return nil, fmt.Errorf("error converting emitter address: %w", err)
	}

	vaa := sdk.VAA{
		Version:          sdk.SupportedVAAVersion,
		GuardianSetIndex: 1,
		EmitterChain:     sdk.ChainID(plm.ChainID),
		EmitterAddress:   address,
		Sequence:         plm.Attributes.Sequence,
		Timestamp:        plm.BlockTime,
		Payload:          plm.Attributes.Payload,
		Nonce:            plm.Attributes.Nonce,
		ConsistencyLevel: plm.Attributes.ConsistencyLevel,
	}

	return &vaa, nil
}
