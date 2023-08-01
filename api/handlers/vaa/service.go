package vaa

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"github.com/wormhole-foundation/wormhole-explorer/api/types"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
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

	s := Service{
		repo:         r,
		getCacheFunc: getCacheFunc,
		logger:       logger.With(zap.String("module", "VaaService")),
	}

	return &s
}

// FindAllParams passes input data to the function `FindAll`.
type FindAllParams struct {
	Pagination           *pagination.Pagination
	TxHash               *types.TxHash
	IncludeParsedPayload bool
	AppId                string
}

// FindAll returns all VAAs.
func (s *Service) FindAll(
	ctx context.Context,
	params *FindAllParams,
) (*response.Response[[]*VaaDoc], error) {

	// Populate query parameters
	query := Query().
		IncludeParsedPayload(params.IncludeParsedPayload)
	if params.Pagination != nil {
		query.SetPagination(params.Pagination)
	}
	if params.TxHash != nil {
		query.SetTxHash(params.TxHash.String())
	}
	if params.AppId != "" {
		query.SetAppId(params.AppId)
	}

	// Execute the database query
	//
	// Unfortunately, for Aptos and Solana, the real transaction hashes are stored
	// in a different collection from other chains.
	//
	// This block of code has additional logic to handle that case.
	var err error
	var vaas []*VaaDoc
	if query.txHash != "" {
		vaas, err = s.repo.FindVaasByTxHashWorkaround(ctx, query)
	} else {
		vaas, err = s.repo.FindVaas(ctx, query)
	}
	if err != nil {
		return nil, err
	}

	// Return the matching documents
	res := response.Response[[]*VaaDoc]{Data: vaas}
	return &res, nil
}

// FindByChain get all the vaa by chainID.
func (s *Service) FindByChain(
	ctx context.Context,
	chain sdk.ChainID,
	p *pagination.Pagination,
) (*response.Response[[]*VaaDoc], error) {

	query := Query().
		SetChain(chain).
		SetPagination(p).
		IncludeParsedPayload(false)

	vaas, err := s.repo.FindVaas(ctx, query)

	res := response.Response[[]*VaaDoc]{Data: vaas}
	return &res, err
}

// FindByEmitter get all the vaa by chainID and emitter address.
func (s *Service) FindByEmitter(
	ctx context.Context,
	emitterChain sdk.ChainID,
	emitterAddress *types.Address,
	toChain *sdk.ChainID,
	includeParsedPayload bool,
	p *pagination.Pagination,
) (*response.Response[[]*VaaDoc], error) {

	query := Query().
		SetChain(emitterChain).
		SetEmitter(emitterAddress.Hex()).
		SetPagination(p).
		IncludeParsedPayload(includeParsedPayload)

	// In most cases, the data is obtained from the VAA collection.
	//
	// The special case of filtering VAAs by `toChain` requires querying
	// the data from a different collection.
	var vaas []*VaaDoc
	var err error
	if toChain != nil {
		vaas, err = s.repo.FindVaasByEmitterAndToChain(ctx, query, *toChain)
	} else {
		vaas, err = s.repo.FindVaas(ctx, query)
	}

	res := response.Response[[]*VaaDoc]{Data: vaas}
	return &res, err
}

// If the parameter [payload] is true, the parse payload is added in the response.
func (s *Service) FindById(
	ctx context.Context,
	chain sdk.ChainID,
	emitter *types.Address,
	seq string,
	includeParsedPayload bool,
) (*response.Response[*VaaDoc], error) {

	// check vaa sequence indexed
	isVaaNotIndexed := s.discardVaaNotIndexed(ctx, chain, emitter, seq)
	if isVaaNotIndexed {
		return nil, errs.ErrNotFound
	}

	// execute the database query
	vaa, err := s.findById(ctx, chain, emitter, seq, includeParsedPayload)
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
	chain sdk.ChainID,
	emitter *types.Address,
	seq string,
	includeParsedPayload bool,
) (*VaaDoc, error) {

	// query matching documents from the database
	query := Query().
		SetChain(chain).
		SetEmitter(emitter.Hex()).
		SetSequence(seq).
		IncludeParsedPayload(includeParsedPayload)
	docs, err := s.repo.FindVaas(ctx, query)
	if err != nil {
		return nil, err
	}

	// we're expecting exactly one document
	if len(docs) == 0 {
		return nil, errs.ErrNotFound
	}
	if len(docs) > 1 {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		s.logger.Error("can not get more that one vaa by chainID/address/sequence",
			zap.Any("q", query),
			zap.String("requestID", requestID))
		return nil, errs.ErrInternalError
	}

	return docs[0], nil
}

// GetVaaCount get a list a list of vaa count grouped by chainID.
func (s *Service) GetVaaCount(ctx context.Context) (*response.Response[[]*VaaStats], error) {
	q := Query()
	stats, err := s.repo.GetVaaCount(ctx, q)
	res := response.Response[[]*VaaStats]{Data: stats}
	return &res, err
}

// discardVaaNotIndexed discard a vaa request if the input sequence for a chainID, address is greatter than or equals
// the cached value of the sequence for this chainID, address.
// If the sequence does not exist we can not discard the request.
func (s *Service) discardVaaNotIndexed(ctx context.Context, chain sdk.ChainID, emitter *types.Address, seq string) bool {
	key := fmt.Sprintf("%s:%d:%s", "wormscan:vaa-max-sequence", chain, emitter.Hex())
	sequence, err := s.getCacheFunc(ctx, key)
	if err != nil {
		if errors.Is(err, cache.ErrInternal) {
			requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
			s.logger.Error("encountered an internal error while getting value from cache",
				zap.Error(err),
				zap.String("requestID", requestID),
			)
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
