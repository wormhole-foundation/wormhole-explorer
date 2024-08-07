package builder

import (
	"context"
	"fmt"

	"github.com/certusone/wormhole/node/pkg/common"
	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
	"github.com/wormhole-foundation/wormhole-explorer/fly/guardiansets"
	"go.uber.org/zap"
)

const (
	mainnetEthContract = "0x98f3c9e6e3face36baad05fe09d375ef1464288b"
	testnetEthContract = "0x4a8bc80ed5a4067f1ccf107057b8270e0cc11a78"
)

func NewGuardianSetSynchronizer(ctx context.Context, mongo *dbutil.Session, postgres *db.DB, heartbeatChannel chan *gossipv1.Heartbeat, logger *zap.Logger, cfg *config.Configuration, alertClient alert.AlertClient) (*guardiansets.GuardianSetSynchronizer, error) {
	var ethContract string
	switch cfg.P2pNetwork {
	case domain.P2pMainNet:
		ethContract = mainnetEthContract
	case domain.P2pTestNet:
		ethContract = testnetEthContract
	default:
		return nil, fmt.Errorf("unable to fetch guardian set for unknown network %s", cfg.P2pNetwork)
	}

	ethGuardianSet, err := guardiansets.NewEthGuardianSet(ctx, ethContract, cfg.EthereumUrl, alertClient, logger)
	if err != nil {
		return nil, err
	}

	manualGuardianSet := guardiansets.GetManualByEnv(cfg.P2pNetwork, alertClient, logger)

	var guardianSetRepository repository.GuardianSetStorager

	if cfg.DbLayer == config.DbLayerPostgres {
		guardianSetRepository = repository.NewPostgresGuardianSetRepository(postgres, logger)
	} else {
		guardianSetRepository = repository.NewMongoGuardianSetRepository(mongo.Database, logger)
	}

	dbGuardianSet := guardiansets.NewDbGuardianSet(ethGuardianSet, guardianSetRepository, manualGuardianSet, logger)

	err = dbGuardianSet.Sync(ctx)
	if err != nil {
		return nil, err
	}

	gst := common.NewGuardianSetState(heartbeatChannel)

	compositeGuardianSet := guardiansets.NewCompositeGuardianSet(ethGuardianSet, dbGuardianSet, manualGuardianSet, alertClient)

	guardianSetSyncronizer, err := guardiansets.NewGuardianSetSynchronizer(ctx, gst, compositeGuardianSet, logger)
	if err != nil {
		return nil, err
	}

	return guardianSetSyncronizer, nil
}
