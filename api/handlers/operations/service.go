package operations

import (
	"context"
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService create a new Service.
func NewService(repo *Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger.With(zap.String("module", "OperationService"))}
}

// FindById returns the operations for the given chainID/emitter/seq.
func (s *Service) FindById(ctx context.Context, chainID vaa.ChainID,
	emitter *types.Address, seq string) (*OperationDto, error) {
	id := fmt.Sprintf("%d/%s/%s", chainID, emitter.Hex(), seq)
	operation, err := s.repo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}
	return operation, nil
}

// FindAll returns all operations filtered by q.
func (s *Service) FindAll(ctx context.Context, q string, pagination *pagination.Pagination) ([]*OperationDto, error) {
	operations, err := s.repo.FindAll(ctx, q, pagination)
	if err != nil {
		return nil, err
	}
	return operations, nil
}
