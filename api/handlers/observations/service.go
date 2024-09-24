// Package observations handle the request of observations data from governor endpoint defined in the api.
package observations

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Service definition.
type Service struct {
	mongoRepo    *MongoRepository
	postgresRepo *PostgresRepository
	logger       *zap.Logger
}

// NewService create a new Service.
func NewService(mongoRepo *MongoRepository, postgresRepo *PostgresRepository, logger *zap.Logger) *Service {
	return &Service{mongoRepo: mongoRepo,
		postgresRepo: postgresRepo,
		logger:       logger.With(zap.String("module", "ObservationsService"))}
}

// FindAll get all the observations.
func (s *Service) FindAll(ctx context.Context, usePostgres bool, p *FindAllParams) ([]*ObservationDoc, error) {
	query := Query().SetPagination(p.Pagination).SetTxHash(p.TxHash)
	if usePostgres {
		return s.postgresRepo.Find(ctx, query)
	}
	return s.mongoRepo.Find(ctx, query)
}

// FindByChain get all the observations by chainID.
func (s *Service) FindByChain(ctx context.Context, usePostgres bool, chain vaa.ChainID, p *pagination.Pagination) ([]*ObservationDoc, error) {
	query := Query().SetChain(chain).SetPagination(p)
	if usePostgres {
		return s.postgresRepo.Find(ctx, query)
	}
	return s.mongoRepo.Find(ctx, query)
}

// FindByEmitter get all the observations by chainID and emitter address.
func (s *Service) FindByEmitter(
	ctx context.Context,
	usePostgres bool,
	chain vaa.ChainID,
	emitter *types.Address,
	p *pagination.Pagination,
) ([]*ObservationDoc, error) {

	query := Query().
		SetChain(chain).
		SetEmitter(emitter.Hex()).
		SetPagination(p)

	if usePostgres {
		return s.postgresRepo.Find(ctx, query)
	}
	return s.mongoRepo.Find(ctx, query)
}

// FindByVAA get all the observations for a VAA (chainID, emitter addrress and sequence number).
func (s *Service) FindByVAA(
	ctx context.Context,
	usePostgres bool,
	chain vaa.ChainID,
	emitter *types.Address,
	seq string,
	p *pagination.Pagination,
) ([]*ObservationDoc, error) {

	query := Query().
		SetChain(chain).
		SetEmitter(emitter.Hex()).
		SetSequence(seq).
		SetPagination(p)

	if usePostgres {
		return s.postgresRepo.Find(ctx, query)
	}
	return s.mongoRepo.Find(ctx, query)
}

// FindOne get a observation by chainID, emitter address, sequence, signer address and hash.
func (s *Service) FindOne(
	ctx context.Context,
	usePostgres bool,
	chainID vaa.ChainID,
	emitterAddr *types.Address,
	seq string,
	signerAddr *vaa.Address,
	hash []byte,
) (*ObservationDoc, error) {

	query := Query().
		SetChain(chainID).
		SetEmitter(emitterAddr.Hex()).
		SetSequence(seq).
		SetGuardianAddr(signerAddr.String()).
		SetHash(hash)

	if usePostgres {
		return s.postgresRepo.FindOne(ctx, query)
	}
	return s.mongoRepo.FindOne(ctx, query)
}
