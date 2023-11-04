package queue

import "github.com/wormhole-foundation/wormhole/sdk/vaa"

// PythFilter filter vaa event from pyth chain.
func PythFilter(vaaEvent *Event) bool {
	return vaaEvent.ChainID == uint16(vaa.ChainIDPythNet)
}

// NonFilter non filter vaa evant.
func NonFilter(vaaEvent *Event) bool {
	return false
}
