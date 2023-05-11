package transactions

import (
	"context"
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/api/types"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService create a new Service.
func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger.With(zap.String("module", "TransactionService"))}
}

// GetTransactionCount get the last transactions.
func (s *Service) GetTransactionCount(ctx context.Context, q *TransactionCountQuery) ([]TransactionCountResult, error) {
	return s.repo.GetTransactionCount(ctx, q)
}

func (s *Service) GetScorecards(ctx context.Context) (*Scorecards, error) {
	return s.repo.GetScorecards(ctx)
}

func (s *Service) GetTopAssetsByVolume(ctx context.Context, timeSpan *TopAssetsTimeSpan) ([]AssetDTO, error) {
	return s.repo.GetTopAssetsByVolume(ctx, timeSpan)
}

func (s *Service) GetTopChainPairsByNumTransfers(ctx context.Context, timeSpan *TopAssetsTimeSpan) ([]ChainPairDTO, error) {
	return s.repo.GetTopChainPairsByNumTransfers(ctx, timeSpan)
}

// GetChainActivity get chain activity.
func (s *Service) GetChainActivity(ctx context.Context, q *ChainActivityQuery) ([]ChainActivityResult, error) {
	return s.repo.FindChainActivity(ctx, q)
}

// FindGlobalTransactionByID find a global transaction by id.
func (s *Service) FindGlobalTransactionByID(ctx context.Context, chainID vaa.ChainID, emitter *types.Address, seq string) (*GlobalTransactionDoc, error) {

	key := fmt.Sprintf("%d/%s/%s", chainID, emitter.Hex(), seq)
	q := GlobalTransactionQuery{id: key}

	return s.repo.FindGlobalTransactionByID(ctx, &q)
}
