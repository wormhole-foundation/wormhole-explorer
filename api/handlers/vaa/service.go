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

// NewService creates a new VAA Service.
func NewService(r *Repository, getCacheFunc cache.CacheGetFunc, logger *zap.Logger) *Service {
	return &Service{repo: r, getCacheFunc: getCacheFunc, logger: logger.With(zap.String("module", "VaaService"))}
}

// FindAllParams passes input data to the function `FindAll`.
type FindAllParams struct {
	Pagination           *pagination.Pagination
	TxHash               *vaa.Address
	IncludeParsedPayload bool
	AppId                string
}

// FindAll returns all VAAs.
func (s *Service) FindAll(
	ctx context.Context,
	params *FindAllParams,
) (*response.Response[[]*VaaDoc], error) {

	// set up query parameters
	query := Query()
	if params.Pagination != nil {
		query.SetPagination(params.Pagination)
	}
	if params.TxHash != nil {
		query.SetTxHash(params.TxHash.String())
	}
	if params.AppId != "" {
		query.SetAppId(params.AppId)
	}

	// execute the database query
	var err error
	var vaas []*VaaDoc
	if params.IncludeParsedPayload {
		vaas, err = s.repo.FindVaasWithPayload(ctx, query)
	} else {
		vaas, err = s.repo.Find(ctx, query)
	}
	if err != nil {
		return nil, err
	}

	// return the matching documents
	res := response.Response[[]*VaaDoc]{Data: vaas}
	return &res, nil
}

// FindByChain get all the vaa by chainID.
func (s *Service) FindByChain(
	ctx context.Context,
	chain vaa.ChainID,
	p *pagination.Pagination,
) (*response.Response[[]*VaaDoc], error) {

	query := Query().
		SetChain(chain).
		SetPagination(p)

	vaas, err := s.repo.Find(ctx, query)

	res := response.Response[[]*VaaDoc]{Data: vaas}
	return &res, err
}

// FindByEmitter get all the vaa by chainID and emitter address.
func (s *Service) FindByEmitter(
	ctx context.Context,
	chain vaa.ChainID,
	emitter vaa.Address,
	p *pagination.Pagination,
) (*response.Response[[]*VaaDoc], error) {

	query := Query().
		SetChain(chain).
		SetEmitter(emitter.String()).
		SetPagination(p)

	vaas, err := s.repo.Find(ctx, query)

	res := response.Response[[]*VaaDoc]{Data: vaas}
	return &res, err
}

// If the parameter [payload] is true, the parse payload is added in the response.
func (s *Service) FindById(
	ctx context.Context,
	chain vaa.ChainID,
	emitter vaa.Address,
	seq string,
	includeParsedPayload bool,
) (*response.Response[*VaaDoc], error) {

	// check vaa sequence indexed
	isVaaNotIndexed := s.discardVaaNotIndexed(ctx, chain, emitter, seq)
	if isVaaNotIndexed {
		return nil, errs.ErrNotFound
	}

	// execute the database query
	var err error
	var vaa *VaaDoc
	if includeParsedPayload {
		vaa, err = s.findByIdWithPayload(ctx, chain, emitter, seq)
	} else {
		vaa, err = s.findById(ctx, chain, emitter, seq)
	}
	if err != nil {
		return &response.Response[*VaaDoc]{}, err
	}

	// return matching documents
	resp := response.Response[*VaaDoc]{Data: vaa}
	return &resp, err
}

// findById get a vaa by chainID, emitter address and sequence number.
func (s *Service) findById(
	ctx context.Context,
	chain vaa.ChainID,
	emitter vaa.Address,
	seq string,
) (*VaaDoc, error) {

	query := Query().
		SetChain(chain).
		SetEmitter(emitter.String()).
		SetSequence(seq)

	return s.repo.FindOne(ctx, query)
}

// findByIdWithPayload get a vaa with payload data by chainID, emitter address and sequence number.
func (s *Service) findByIdWithPayload(ctx context.Context, chain vaa.ChainID, emitter vaa.Address, seq string) (*VaaDoc, error) {
	query := Query().SetChain(chain).SetEmitter(emitter.String()).SetSequence(seq)

	vaas, err := s.repo.FindVaasWithPayload(ctx, query)
	if err != nil {
		return nil, err
	} else if len(vaas) == 0 {
		return nil, errs.ErrNotFound
	} else if len(vaas) > 1 {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		s.logger.Error("can not get more that one vaa by chainID/address/sequence",
			zap.Any("q", query),
			zap.String("requestID", requestID),
		)
		return nil, errs.ErrInternalError
	}

	return vaas[0], nil
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
