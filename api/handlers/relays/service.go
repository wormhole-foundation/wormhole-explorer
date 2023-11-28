package relays

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/api/types"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService create a new Service.
func NewService(dao *Repository, logger *zap.Logger) *Service {
	return &Service{repo: dao, logger: logger.With(zap.String("module", "RelaysService"))}
}

// Find by VAA by chainID, emitter address, sequence
func (s *Service) FindByVAA(
	ctx context.Context,
	chainID vaa.ChainID,
	emitterAddr *types.Address,
	seq string,
) (*RelayDoc, error) {

	query := Query().
		SetChain(chainID).
		SetEmitter(emitterAddr.Hex()).
		SetSequence(seq)

	return s.repo.FindOne(ctx, query)
}
