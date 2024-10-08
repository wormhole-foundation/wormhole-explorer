package operations

import (
	"context"
	"fmt"
	"time"

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

type OperationFilter struct {
	TxHash         *types.TxHash
	Address        string
	SourceChainIDs []vaa.ChainID
	TargetChainIDs []vaa.ChainID
	AppIDs         []string
	ExclusiveAppId bool
	Pagination     pagination.Pagination
	PayloadType    []int
	From           *time.Time
	To             *time.Time
}

// FindAll returns all operations filtered by q.
func (s *Service) FindAll(ctx context.Context, filter OperationFilter) ([]*OperationDto, error) {
	var txHash string
	if filter.TxHash != nil {
		txHash = filter.TxHash.String()
	}

	operationQuery := OperationQuery{
		TxHash:         txHash,
		Address:        filter.Address,
		Pagination:     filter.Pagination,
		SourceChainIDs: filter.SourceChainIDs,
		TargetChainIDs: filter.TargetChainIDs,
		AppIDs:         filter.AppIDs,
		ExclusiveAppId: filter.ExclusiveAppId,
		PayloadType:    filter.PayloadType,
		From:           filter.From,
		To:             filter.To,
	}

	if len(operationQuery.AppIDs) != 0 || len(operationQuery.SourceChainIDs) > 0 || len(operationQuery.TargetChainIDs) > 0 || len(operationQuery.PayloadType) > 0 {
		return s.repo.FindFromParsedVaa(ctx, operationQuery)
	}

	operations, err := s.repo.FindAll(ctx, operationQuery)
	if err != nil {
		return nil, err
	}
	return operations, nil
}
