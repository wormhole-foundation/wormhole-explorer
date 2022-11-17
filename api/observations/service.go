// Package observations handle the request of observations data from governor endpoint defined in the api.
package observations

import (
	"context"

	"github.com/certusone/wormhole/node/pkg/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/api/pagination"
	"go.uber.org/zap"
)

// Service definition.
type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService create a new Service.
func NewService(dao *Repository, logger *zap.Logger) *Service {
	return &Service{repo: dao, logger: logger.With(zap.String("module", "ObservationsService"))}
}

// FindAll get all the observations.
func (s *Service) FindAll(ctx context.Context, p *pagination.Pagination) ([]*ObservationDoc, error) {
	return s.repo.Find(ctx, Query().SetPagination(p))
}

// FindByChain get all the observations by chainID.
func (s *Service) FindByChain(ctx context.Context, chain vaa.ChainID, p *pagination.Pagination) ([]*ObservationDoc, error) {
	query := Query().SetChain(chain).SetPagination(p)
	return s.repo.Find(ctx, query)
}

// FindByEmitter get all the observations by chainID and emitter address.
func (s *Service) FindByEmitter(ctx context.Context, chain vaa.ChainID, emitter *vaa.Address, p *pagination.Pagination) ([]*ObservationDoc, error) {
	query := Query().SetChain(chain).SetEmitter(emitter.String()).SetPagination(p)
	return s.repo.Find(ctx, query)
}

// FindByVAA get all the observations for a VAA (chainID, emitter addrress and sequence number).
func (s *Service) FindByVAA(ctx context.Context, chain vaa.ChainID, emitter *vaa.Address, seq uint64, p *pagination.Pagination) ([]*ObservationDoc, error) {
	query := Query().SetChain(chain).SetEmitter(emitter.String()).SetSequence(seq).SetPagination(p)
	return s.repo.Find(ctx, query)
}

// FindOne get a observation by chainID, emitter address, sequence, signer address and hash.
func (s *Service) FindOne(ctx context.Context, chainID vaa.ChainID, emitterAddr *vaa.Address, seq uint64, signerAddr *vaa.Address, hash []byte) (*ObservationDoc, error) {
	query := Query().SetChain(chainID).SetEmitter(emitterAddr.String()).SetSequence(seq).SetGuardianAddr(signerAddr.String()).SetHash(hash)
	return s.repo.FindOne(ctx, query)
}
