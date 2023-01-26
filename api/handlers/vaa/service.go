package vaa

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/cache"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Service definition.
type Service struct {
	repo         *Repository
	getCacheFunc cache.CacheGetFunc
	logger       *zap.Logger
}

// NewService create a new Service.
func NewService(r *Repository, getCacheFunc cache.CacheGetFunc, logger *zap.Logger) *Service {
	return &Service{repo: r, getCacheFunc: getCacheFunc, logger: logger.With(zap.String("module", "VaaService"))}
}

// FindAll get all the the vaa.
func (s *Service) FindAll(ctx context.Context, p *pagination.Pagination, txHash *vaa.Address) (*response.Response[[]*VaaDoc], error) {
	if p == nil {
		p = pagination.FirstPage()
	}
	query := Query().SetPagination(p)
	if txHash != nil {
		query = query.SetTxHash(txHash.String())
	}
	vaas, err := s.repo.Find(ctx, query)
	res := response.Response[[]*VaaDoc]{Data: vaas}
	return &res, err
}

// FindByChain get all the vaa by chainID.
func (s *Service) FindByChain(ctx context.Context, chain vaa.ChainID, p *pagination.Pagination) (*response.Response[[]*VaaDoc], error) {
	query := Query().SetChain(chain).SetPagination(p)
	vaas, err := s.repo.Find(ctx, query)
	res := response.Response[[]*VaaDoc]{Data: vaas}
	return &res, err
}

// FindByEmitter get all the vaa by chainID and emitter address.
func (s *Service) FindByEmitter(ctx context.Context, chain vaa.ChainID, emitter vaa.Address, p *pagination.Pagination) (*response.Response[[]*VaaDoc], error) {
	query := Query().SetChain(chain).SetEmitter(emitter.String()).SetPagination(p)
	vaas, err := s.repo.Find(ctx, query)
	res := response.Response[[]*VaaDoc]{Data: vaas}
	return &res, err
}

// FindById get a vaa by chainID, emitter address and sequence number.
func (s *Service) FindById(ctx context.Context, chain vaa.ChainID, emitter vaa.Address, seq string) (*response.Response[*VaaDoc], error) {
	// check vaa sequence indexed
	isVaaNotIndexed := s.discardVaaNotIndexed(ctx, chain, emitter, seq)
	if isVaaNotIndexed {
		return nil, errs.ErrNotFound
	}

	query := Query().SetChain(chain).SetEmitter(emitter.String()).SetSequence(seq)
	vaas, err := s.repo.FindOne(ctx, query)
	res := response.Response[*VaaDoc]{Data: vaas}
	return &res, err
}

// GetVaaCount get a list a list of vaa count grouped by chainID.
func (s *Service) GetVaaCount(ctx context.Context, p *pagination.Pagination) (*response.Response[[]*VaaStats], error) {
	if p == nil {
		p = pagination.FirstPage()
	}
	query := Query().SetPagination(p)
	stats, err := s.repo.GetVaaCount(ctx, query)
	res := response.Response[[]*VaaStats]{Data: stats}
	return &res, err
}

// discardVaaNotIndexed discard a vaa request if the input sequence for a chainID, address is greatter than or equals
// the cached value of the sequence for this chainID, address.
// If the sequence does not exist we can not discard the request.
func (s *Service) discardVaaNotIndexed(ctx context.Context, chain vaa.ChainID, emitter vaa.Address, seq string) bool {
	key := fmt.Sprintf("%s:%d:%s", "wormscan:vaa-max-sequence", chain, emitter.String())
	sequence, err := s.getCacheFunc(ctx, key)
	if err != nil {
		if errors.Is(err, errs.ErrInternalError) {
			requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
			s.logger.Error("error getting value from cache",
				zap.Error(err), zap.String("requestID", requestID))
		}
		return false
	}

	inputSquence, err := strconv.ParseUint(seq, 10, 64)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		s.logger.Error("error invalid input sequence number",
			zap.Error(err), zap.String("seq", seq), zap.String("requestID", requestID))
		return false
	}
	cacheSequence, err := strconv.ParseUint(sequence, 10, 64)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		s.logger.Error("error invalid cached sequence number",
			zap.Error(err), zap.String("sequence", sequence), zap.String("requestID", requestID))
		return false
	}

	// Check that the input sequence is indexed.
	if cacheSequence >= inputSquence {
		return false
	}
	return true
}
