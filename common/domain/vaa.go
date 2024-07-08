package domain

import sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"

// ConsistencyLevelIsImmediately returns true if the VAA is to be published immediately
func ConsistencyLevelIsImmediately(v *sdk.VAA) bool {

	//https://docs.wormhole.com/wormhole/reference/constants#consistency-levels
	if v.EmitterChain == sdk.ChainIDSolana {
		return v.ConsistencyLevel == 0
	}

	if v.ConsistencyLevel == sdk.ConsistencyLevelPublishImmediately {
		return true
	}

	return false
}
