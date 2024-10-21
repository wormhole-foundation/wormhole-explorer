package relays

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type Service struct {
	mongoRepo    *MongoRepository
	postgresRepo *PostgresRepository
	logger       *zap.Logger
}

// NewService create a new Service.
func NewService(mongoRepo *MongoRepository, postgresRepo *PostgresRepository, logger *zap.Logger) *Service {
	return &Service{mongoRepo: mongoRepo, postgresRepo: postgresRepo, logger: logger.With(zap.String("module", "RelaysService"))}
}

// Find by VAA by chainID, emitter address, sequence
func (s *Service) FindByVAA(
	ctx context.Context,
	usePostgres bool,
	chainID vaa.ChainID,
	emitterAddr *types.Address,
	seq string,
) (*RelayDoc, error) {

	query := Query().
		SetChain(chainID).
		SetEmitter(emitterAddr.Hex()).
		SetSequence(seq)

	if usePostgres {
		return s.postgresRepo.FindOne(ctx, query)
	}

	return s.mongoRepo.FindOne(ctx, query)
}
