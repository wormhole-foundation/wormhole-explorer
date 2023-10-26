package address

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"go.uber.org/zap"
)

type Service struct {
	repo   *Repository
	logger *zap.Logger
}

func NewService(r *Repository, logger *zap.Logger) *Service {

	srv := Service{
		repo:   r,
		logger: logger.With(zap.String("module", "AddressService")),
	}

	return &srv
}

func (s *Service) GetAddressOverview(
	ctx context.Context,
	address string,
	pagination *pagination.Pagination,
) (*response.Response[*AddressOverview], error) {

	response := &response.Response[*AddressOverview]{}

	p := GetAddressOverviewParams{
		Address: address,
		Skip:    pagination.Skip,
		Limit:   pagination.Limit,
	}
	overview, err := s.repo.GetAddressOverview(ctx, &p)
	if err != nil {
		return response, err
	}

	response.Data = overview
	return response, nil
}
