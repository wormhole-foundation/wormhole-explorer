package queue

import "github.com/wormhole-foundation/wormhole/sdk/vaa"

// PythFilter filter vaa event from pyth chain.
func PythFilter(event *Event) bool {
	return event.ChainID == uint16(vaa.ChainIDPythNet)
}

// NonFilter non filter vaa evant.
func NonFilter(event *Event) bool {
	return false
}
