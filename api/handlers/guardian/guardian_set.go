package guardian

import (
	"time"

	"github.com/certusone/wormhole/node/pkg/common"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/config"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
)

// GuardianSet definition.
type GuardianSet struct {
	GstByIndex            []common.GuardianSet
	ExpirationTimeByIndex []time.Time
}

// Get get guardianset config by enviroment.
func getByEnv(enviroment string) GuardianSet {
	switch enviroment {
	case config.P2pTestNet:
		return getTestnetGuardianSet()
	default:
		return getMainnetGuardianSet()
	}
}

// IsValid check if a guardianSet is valid.
func (gs GuardianSet) IsValid(gsIx uint32, t time.Time) bool {
	if int(gsIx) > len(gs.GstByIndex) {
		return false
	}
	return gs.ExpirationTimeByIndex[gsIx].After(t)
}

// GetLatest get the lastest guardianset.
func (gs GuardianSet) GetLatest() common.GuardianSet {
	return gs.GstByIndex[len(gs.GstByIndex)-1]
}

func getTestnetGuardianSet() GuardianSet {
	gstByIndex, expirationTimeByIndex := domain.GetTestnetGuardianSet()
	return GuardianSet{
		GstByIndex:            gstByIndex,
		ExpirationTimeByIndex: expirationTimeByIndex,
	}
}

func getMainnetGuardianSet() GuardianSet {
	gstByIndex, expirationTimeByIndex := domain.GetMainnetGuardianSet()
	return GuardianSet{
		GstByIndex:            gstByIndex,
		ExpirationTimeByIndex: expirationTimeByIndex,
	}
}
