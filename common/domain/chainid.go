package domain

import "github.com/wormhole-foundation/wormhole/sdk/vaa"

// GetSupportedChainIDs returns a map of all supported chain IDs to their respective names.
func GetSupportedChainIDs() map[vaa.ChainID]string {
	chainIDs := vaa.GetAllNetworkIDs()
	supportedChaindIDs := make(map[vaa.ChainID]string, len(chainIDs))
	for _, chainID := range chainIDs {
		supportedChaindIDs[chainID] = chainID.String()
	}
	return supportedChaindIDs
}
