package transactions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/api/cacheable"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/api/types"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type Service struct {
	repo              *Repository
	cache             cache.Cache
	supportedChainIDs map[vaa.ChainID]string
	logger            *zap.Logger
}

// NewService create a new Service.
func NewService(repo *Repository, cache cache.Cache, logger *zap.Logger) *Service {
	supportedChainIDs := domain.GetSupportedChainIDs()
	return &Service{repo: repo, supportedChainIDs: supportedChainIDs, cache: cache, logger: logger.With(zap.String("module", "TransactionService"))}
}

// GetTransactionCount get the last transactions.
func (s *Service) GetTransactionCount(ctx context.Context, q *TransactionCountQuery) ([]TransactionCountResult, error) {
	return s.repo.GetTransactionCount(ctx, q)
}

func (s *Service) GetScorecards(ctx context.Context) (*Scorecards, error) {
	return s.repo.GetScorecards(ctx)
}

func (s *Service) GetTopAssets(ctx context.Context, timeSpan *TopStatisticsTimeSpan) ([]AssetDTO, error) {
	return s.repo.GetTopAssets(ctx, timeSpan)
}

func (s *Service) GetTopChainPairs(ctx context.Context, timeSpan *TopStatisticsTimeSpan) ([]ChainPairDTO, error) {
	return s.repo.GetTopChainPairs(ctx, timeSpan)
}

// GetChainActivity get chain activity.
func (s *Service) GetChainActivity(ctx context.Context, q *ChainActivityQuery) ([]ChainActivityResult, error) {
	key := fmt.Sprintf("wormscan:chain-activity:%s:%v:%s", q.TimeSpan, q.IsNotional, strings.Join(q.GetAppIDs(), ","))
	return cacheable.GetOrLoad(ctx, s.logger, s.cache, 5*time.Minute, key,
		func() ([]ChainActivityResult, error) {
			return s.repo.FindChainActivity(ctx, q)
		})
}

// FindGlobalTransactionByID find a global transaction by id.
func (s *Service) FindGlobalTransactionByID(ctx context.Context, chainID vaa.ChainID, emitter *types.Address, seq string) (*GlobalTransactionDoc, error) {

	key := fmt.Sprintf("%d/%s/%s", chainID, emitter.Hex(), seq)
	q := GlobalTransactionQuery{id: key}

	return s.repo.FindGlobalTransactionByID(ctx, &q)
}

// GetTokenByChainAndAddress get token by chain and address.
func (s *Service) GetTokenByChainAndAddress(ctx context.Context, chainID vaa.ChainID, tokenAddress *types.Address) (*Token, error) {
	// check if chainID is valid
	if _, ok := s.supportedChainIDs[chainID]; !ok {
		return nil, errs.ErrNotFound
	}

	//get token by contractID (chainID + tokenAddress)
	tokenMetadata, ok := domain.GetTokenByAddress(chainID, tokenAddress.Hex())
	if !ok {
		return nil, errs.ErrNotFound
	}

	return &Token{
		Symbol:      tokenMetadata.Symbol,
		CoingeckoID: tokenMetadata.CoingeckoID,
		Decimals:    tokenMetadata.Decimals,
	}, nil
}

func (s *Service) ListTransactions(
	ctx context.Context,
	pagination *pagination.Pagination,
) (*ListTransactonsOutput, error) {

	return s.repo.ListTransactions(ctx, pagination)
}

func (s *Service) ListTransactionsByAddress(
	ctx context.Context,
	address *types.Address,
	pagination *pagination.Pagination,
) (*ListTransactonsOutput, error) {

	return s.repo.ListTransactionsByAddress(ctx, address, pagination)
}
