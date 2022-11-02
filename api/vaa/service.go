package vaa

import (
	"context"
	"github.com/certusone/wormhole/node/pkg/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/api/pagination"
	"go.uber.org/zap"
)

type Service struct {
	repo   *Repository
	logger *zap.Logger
}

func NewService(r *Repository, logger *zap.Logger) *Service {
	return &Service{repo: r, logger: logger.With(zap.String("module", "VaaService"))}
}

func (s *Service) FindAll(ctx context.Context, p *pagination.Pagination) ([]*VaaDoc, error) {
	if p == nil {
		p = pagination.FirstPage()
	}

	query := Query().SetPagination(p)
	return s.repo.Find(ctx, query)
}

func (s *Service) FindByChain(ctx context.Context, chain vaa.ChainID, p *pagination.Pagination) ([]*VaaDoc, error) {
	query := Query().SetChain(chain).SetPagination(p)
	return s.repo.Find(ctx, query)
}

func (s *Service) FindByEmitter(ctx context.Context, chain vaa.ChainID, emitter vaa.Address, p *pagination.Pagination) ([]*VaaDoc, error) {
	query := Query().SetChain(chain).SetEmitter(emitter.String()).SetPagination(p)
	return s.repo.Find(ctx, query)
}

func (s *Service) FindById(ctx context.Context, chain vaa.ChainID, emitter vaa.Address, seq uint64) (*VaaDoc, error) {
	query := Query().SetChain(chain).SetEmitter(emitter.String()).SetSequence(seq)
	return s.repo.FindOne(ctx, query)
}

func (s *Service) GetVAAStats(ctx context.Context) ([]*VaaStats, error) {
	return s.repo.FindStats(ctx)
}
