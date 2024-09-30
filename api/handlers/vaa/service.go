package vaa

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	vaaPayloadParser "github.com/wormhole-foundation/wormhole-explorer/common/client/parser"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Service definition.
type Service struct {
	mongoRepo    *MongoRepository
	postgresRepo *PostgresRepository
	getCacheFunc cache.CacheGetFunc
	parseVaaFunc vaaPayloadParser.ParseVaaFunc
	logger       *zap.Logger
}

// NewService creates a new VAA Service.
func NewService(
	r *MongoRepository,
	p *PostgresRepository,
	getCacheFunc cache.CacheGetFunc,
	parseVaaFunc vaaPayloadParser.ParseVaaFunc,
	logger *zap.Logger) *Service {

	s := Service{
		mongoRepo:    r,
		postgresRepo: p,
		getCacheFunc: getCacheFunc,
		parseVaaFunc: parseVaaFunc,
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
	usePostgres bool,
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

	// Execute the database query
	//
	// Unfortunately, for Aptos and Solana, the real transaction hashes are stored
	// in a different collection from other chains.
	//
	// This block of code has additional logic to handle that case.
	var err error
	var vaas []*VaaDoc
	if usePostgres {
		vaas, err = s.postgresRepo.Find(ctx, query)
		if err != nil {
			return nil, err
		}
	} else {
		if query.txHash != "" {
			vaas, err = s.mongoRepo.FindVaasByTxHashWorkaround(ctx, query)
		} else {
			vaas, err = s.mongoRepo.FindVaas(ctx, query)
		}
		if err != nil {
			return nil, err
		}
	}

	// Return the matching documents
	res := response.Response[[]*VaaDoc]{Data: vaas}
	return &res, nil
}

// FindByChain get all the vaa by chainID.
func (s *Service) FindByChain(
	ctx context.Context,
	usePostgres bool,
	chain sdk.ChainID,
	p *pagination.Pagination,
) (*response.Response[[]*VaaDoc], error) {

	query := Query().
		SetChain(chain).
		SetPagination(p).
		IncludeParsedPayload(false)

	var vaas []*VaaDoc
	var err error
	if usePostgres {
		vaas, err = s.postgresRepo.Find(ctx, query)
	} else {
		vaas, err = s.mongoRepo.FindVaas(ctx, query)
	}
	if err != nil {
		return nil, err
	}

	res := response.Response[[]*VaaDoc]{Data: vaas}
	return &res, err
}

// FindByEmitterParams contains the input parameters for the function `FindByEmitter`.
type FindByEmitterParams struct {
	EmitterChain         sdk.ChainID
	EmitterAddress       *types.Address
	ToChain              *sdk.ChainID
	IncludeParsedPayload bool
	Pagination           *pagination.Pagination
}

// FindByEmitter get all the vaa by chainID and emitter address.
func (s *Service) FindByEmitter(
	ctx context.Context,
	usePostgres bool,
	params *FindByEmitterParams,
) (*response.Response[[]*VaaDoc], error) {

	query := Query().
		SetChain(params.EmitterChain).
		SetEmitter(params.EmitterAddress.Hex()).
		SetPagination(params.Pagination).
		IncludeParsedPayload(params.IncludeParsedPayload)

	// In most cases, the data is obtained from the VAA collection.
	//
	// The special case of filtering VAAs by `toChain` requires querying
	// the data from a different collection.
	var vaas []*VaaDoc
	var err error

	if usePostgres {
		if params.ToChain != nil {
			query.SetToChain(*params.ToChain)
		}
		vaas, err = s.postgresRepo.Find(ctx, query)
	} else {
		if params.ToChain != nil {
			vaas, err = s.mongoRepo.FindVaasByEmitterAndToChain(ctx, query, *params.ToChain)
		} else {
			vaas, err = s.mongoRepo.FindVaas(ctx, query)
		}
	}
	if err != nil {
		return nil, err
	}

	res := response.Response[[]*VaaDoc]{Data: vaas}
	return &res, err
}

// If the parameter [payload] is true, the parse payload is added in the response.
func (s *Service) FindById(
	ctx context.Context,
	usePostgres bool,
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
	vaa, err := s.findById(ctx, usePostgres, chain, emitter, seq, includeParsedPayload)
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
	usePostgres bool,
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

	var vaas []*VaaDoc
	var err error
	if usePostgres {
		vaas, err = s.postgresRepo.Find(ctx, query)
	} else {
		vaas, err = s.mongoRepo.FindVaas(ctx, query)
	}

	if err != nil {
		return nil, err
	}

	// we're expecting exactly one document
	if len(vaas) == 0 {
		return nil, errs.ErrNotFound
	}
	if len(vaas) > 1 {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		s.logger.Error("can not get more that one vaa by chainID/address/sequence",
			zap.Any("q", query),
			zap.String("requestID", requestID))
		return nil, errs.ErrInternalError
	}

	return vaas[0], nil
}

// GetVaaCount get a list a list of vaa count grouped by chainID.
// TODO: handle this endpoint with postgres or influx?
func (s *Service) GetVaaCount(ctx context.Context) (*response.Response[[]*VaaStats], error) {
	q := Query()
	stats, err := s.mongoRepo.GetVaaCount(ctx, q)
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

// ParseVaa parse a vaa payload.
func (s *Service) ParseVaa(ctx context.Context, vaaByte []byte) (any, error) {
	// unmarshal vaa
	vaa, err := sdk.Unmarshal(vaaByte)
	if err != nil {
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		s.logger.Error("error unmarshal vaa to parse", zap.Error(err), zap.String("requestID", requestID))
		return vaaPayloadParser.ParseVaaWithStandarizedPropertiesdResponse{}, errs.ErrInternalError
	}

	// call vaa payload parser api
	parsedVaa, err := s.parseVaaFunc(vaa)
	if err != nil {
		if errors.Is(err, vaaPayloadParser.ErrNotFound) {
			return nil, errs.ErrNotFound
		}
		requestID := fmt.Sprintf("%v", ctx.Value("requestid"))
		s.logger.Error("error parse vaa", zap.Error(err), zap.String("requestID", requestID))
		return nil, errs.ErrInternalError
	}
	return parsedVaa, nil
}

// If the parameter [payload] is true, the parse payload is added in the response.
func (s *Service) FindDuplicatedById(
	ctx context.Context,
	usePostgres bool,
	chain sdk.ChainID,
	emitter *types.Address,
	seq string,
) (*response.Response[[]*VaaDoc], error) {

	// check vaa sequence indexed
	isVaaNotIndexed := s.discardVaaNotIndexed(ctx, chain, emitter, seq)
	if isVaaNotIndexed {
		return nil, errs.ErrNotFound
	}

	var vaas []*VaaDoc
	var err error

	if usePostgres {
		query := Query().
			SetChain(chain).
			SetEmitter(emitter.Hex()).
			SetSequence(seq)
		vaas, err = s.postgresRepo.Find(ctx, query)
	} else {
		vaas, err = s.mongoRepo.FindDuplicatedByID(ctx, chain, emitter, seq)
	}

	if err != nil {
		return nil, err
	}

	if len(vaas) == 0 {
		return nil, errs.ErrNotFound
	}

	// return matching documents
	resp := response.Response[[]*VaaDoc]{Data: vaas}
	return &resp, err
}
