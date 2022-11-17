package vaa

import (
	"context"

	"github.com/certusone/wormhole/node/pkg/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/api/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/api/services"
	"go.uber.org/zap"
)

// Service definition.
type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService create a new Service.
func NewService(r *Repository, logger *zap.Logger) *Service {
	return &Service{repo: r, logger: logger.With(zap.String("module", "VaaService"))}
}

// FindAll get all the the vaa.
func (s *Service) FindAll(ctx context.Context, p *pagination.Pagination) (*services.Response[[]*VaaDoc], error) {
	if p == nil {
		p = pagination.FirstPage()
	}

	query := Query().SetPagination(p)
	vaas, err := s.repo.Find(ctx, query)
	res := services.Response[[]*VaaDoc]{Data: vaas}
	return &res, err
}

// FindByChain get all the vaa by chainID.
func (s *Service) FindByChain(ctx context.Context, chain vaa.ChainID, p *pagination.Pagination) (*services.Response[[]*VaaDoc], error) {
	query := Query().SetChain(chain).SetPagination(p)
	vaas, err := s.repo.Find(ctx, query)
	res := services.Response[[]*VaaDoc]{Data: vaas}
	return &res, err
}

// FindByEmitter get all the vaa by chainID and emitter address.
func (s *Service) FindByEmitter(ctx context.Context, chain vaa.ChainID, emitter vaa.Address, p *pagination.Pagination) (*services.Response[[]*VaaDoc], error) {
	query := Query().SetChain(chain).SetEmitter(emitter.String()).SetPagination(p)
	vaas, err := s.repo.Find(ctx, query)
	res := services.Response[[]*VaaDoc]{Data: vaas}
	return &res, err
}

// FindById get a vaa by chainID, emitter address and sequence number.
func (s *Service) FindById(ctx context.Context, chain vaa.ChainID, emitter vaa.Address, seq uint64) (*services.Response[*VaaDoc], error) {
	query := Query().SetChain(chain).SetEmitter(emitter.String()).SetSequence(seq)
	vaas, err := s.repo.FindOne(ctx, query)
	res := services.Response[*VaaDoc]{Data: vaas}
	return &res, err
}

func (s *Service) GetVAAStats(ctx context.Context) (*services.Response[[]*VaaStats], error) {
	stats, err := s.repo.FindStats(ctx)
	res := services.Response[[]*VaaStats]{Data: stats}
	return &res, err
}
