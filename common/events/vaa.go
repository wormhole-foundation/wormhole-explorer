package events

import (
	"encoding/hex"
	"fmt"
	"strings"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

func CreateUnsignedVAA(plm *LogMessagePublished) (*sdk.VAA, error) {

	address, err := sdk.StringToAddress(plm.Attributes.Sender)
	if err != nil {
		return nil, fmt.Errorf("error converting emitter address: %w", err)
	}
	payload, err := hex.DecodeString(strings.TrimPrefix(plm.Attributes.Payload, "0x"))
	if err != nil {
		return nil, fmt.Errorf("error converting payload: %w", err)
	}

	vaa := sdk.VAA{
		Version:          sdk.SupportedVAAVersion,
		GuardianSetIndex: 1,
		EmitterChain:     sdk.ChainID(plm.ChainID),
		EmitterAddress:   address,
		Sequence:         plm.Attributes.Sequence,
		Timestamp:        plm.BlockTime,
		Payload:          payload,
		Nonce:            plm.Attributes.Nonce,
		ConsistencyLevel: plm.Attributes.ConsistencyLevel,
	}

	return &vaa, nil
}
