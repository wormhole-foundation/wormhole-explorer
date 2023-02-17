package queue

import "github.com/wormhole-foundation/wormhole/sdk/vaa"

// PythFilter filter vaa event from pyth chain.
func PythFilter(vaaEvent *VaaEvent) bool {
	return vaaEvent.ChainID == uint16(vaa.ChainIDPythNet)
}

// NonFilter non filter vaa evant.
func NonFilter(vaaEvent *VaaEvent) bool {
	return false
}
