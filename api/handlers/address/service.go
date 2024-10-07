package address

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"go.uber.org/zap"
)

type Service struct {
	mongoRepo *MongoRepository
	logger    *zap.Logger
}

func NewService(r *MongoRepository, logger *zap.Logger) *Service {
	srv := Service{
		mongoRepo: r,
		logger:    logger.With(zap.String("module", "AddressService")),
	}
	return &srv
}

func (s *Service) GetAddressOverview(
	ctx context.Context,
	address string,
	pagination *pagination.Pagination,
	usePostgres bool,
) (*response.Response[*AddressOverview], error) {

	response := &response.Response[*AddressOverview]{}

	p := GetAddressOverviewParams{
		Address: address,
		Skip:    pagination.Skip,
		Limit:   pagination.Limit,
	}

	var overview *AddressOverview
	var err error

	if usePostgres {
		// Deprecated endpoint.
		overview = &AddressOverview{}
	} else {
		overview, err = s.mongoRepo.GetAddressOverview(ctx, &p)
	}

	if err != nil {
		return response, err
	}

	response.Data = overview
	return response, nil
}
