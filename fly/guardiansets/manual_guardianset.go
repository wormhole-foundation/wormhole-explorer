package guardiansets

import (
	"context"
	"time"

	"github.com/certusone/wormhole/node/pkg/common"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"go.uber.org/zap"
)

type manualGuardianSet struct {
	gsth   *GuardianSetHistory
	logger *zap.Logger
}

// Get get guardianset config by enviroment.
func GetManualByEnv(p2pNetwork string, alertClient alert.AlertClient, logger *zap.Logger) *manualGuardianSet {
	var gsth GuardianSetHistory
	switch p2pNetwork {
	case domain.P2pTestNet:
		gsth = getTestnetGuardianSet(alertClient)
	default:
		gsth = getMainnetGuardianSet(alertClient)
	}
	return &manualGuardianSet{gsth: &gsth, logger: logger}
}

func getTestnetGuardianSet(alertClient alert.AlertClient) GuardianSetHistory {
	guardianSetsByIndex, expirationTimesByIndex := domain.GetTestnetGuardianSet()
	return GuardianSetHistory{
		guardianSetsByIndex:    guardianSetsByIndex,
		expirationTimesByIndex: expirationTimesByIndex,
		alertClient:            alertClient,
	}
}

func getMainnetGuardianSet(alertClient alert.AlertClient) GuardianSetHistory {
	guardianSetsByIndex, expirationTimesByIndex := domain.GetMainnetGuardianSet()
	return GuardianSetHistory{
		guardianSetsByIndex:    guardianSetsByIndex,
		expirationTimesByIndex: expirationTimesByIndex,
		alertClient:            alertClient,
	}
}

func (m *manualGuardianSet) GetGuardianSet(ctx context.Context, index uint32) (*common.GuardianSet, *time.Time, error) {
	m.logger.Debug("Fetching manual guardian set", zap.Uint32("index", index))
	if int(index) >= len(m.gsth.guardianSetsByIndex) {
		return nil, nil, nil
	}
	return &m.gsth.guardianSetsByIndex[index], &m.gsth.expirationTimesByIndex[index], nil
}

func (e *manualGuardianSet) GetGuardianSetHistory(ctx context.Context) (*GuardianSetHistory, error) {
	return e.gsth, nil
}
