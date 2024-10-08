// Package governor handle the request of governor data from governor endpoint defined in the api.
package governor

import (
	"context"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/api/cacheable"
	errs "github.com/wormhole-foundation/wormhole-explorer/api/internal/errors"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/cache"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type Service struct {
	mongoRepo         *MongoRepository
	postgresRepo      *PostgresRepository
	cache             cache.Cache
	metrics           metrics.Metrics
	supportedChainIDs map[vaa.ChainID]string
	logger            *zap.Logger
}

const (
	availableNotionByChain = "wormscan:available-notion-by-chain"
	tokenList              = "wormscan:token-list"
)

// NewService create a new governor.Service.
func NewService(mongoRepo *MongoRepository, postgresRepo *PostgresRepository, cache cache.Cache, metrics metrics.Metrics, logger *zap.Logger) *Service {
	supportedChainIDs := domain.GetSupportedChainIDs()
	return &Service{mongoRepo: mongoRepo, postgresRepo: postgresRepo, cache: cache, metrics: metrics, supportedChainIDs: supportedChainIDs, logger: logger.With(zap.String("module", "GovernorService"))}
}

// FindGovernorConfig get a list of governor configurations.
func (s *Service) FindGovernorConfig(
	ctx context.Context,
	p *pagination.Pagination,
	usePostgres bool,
) (*response.Response[[]*GovConfig], error) {

	// set default pagination if not provided
	if p == nil {
		p = pagination.Default()
	}

	query := NewGovernorQuery().SetPagination(p)

	var govConfigs []*GovConfig
	var err error
	if usePostgres {
		govConfigs, err = s.postgresRepo.FindGovConfigurations(ctx, query)
	} else {
		govConfigs, err = s.mongoRepo.FindGovConfigurations(ctx, query)
	}

	if err != nil {
		return nil, err
	}

	res := response.Response[[]*GovConfig]{Data: govConfigs}
	return &res, err
}

// FindGovernorConfigByGuardianAddress get a governor configuration by guardianAddress.
func (s *Service) FindGovernorConfigByGuardianAddress(
	ctx context.Context,
	guardianAddress *types.Address,
	usePostgres bool,
) ([]*GovConfig, error) {

	p := pagination.
		Default().
		SetLimit(1)

	query := NewGovernorQuery().
		SetID(guardianAddress).
		SetPagination(p)

	var govConfigs []*GovConfig
	var err error
	if usePostgres {
		govConfigs, err = s.postgresRepo.FindGovConfigurations(ctx, query)
	} else {
		govConfigs, err = s.mongoRepo.FindGovConfigurations(ctx, query)
	}

	if err != nil {
		return nil, err
	}

	return govConfigs, err
}

// FindGovernorStatus get a list of governor status.
func (s *Service) FindGovernorStatus(ctx context.Context, usePostgres bool, p *pagination.Pagination) (*response.Response[[]*GovStatus], error) {
	if p == nil {
		p = pagination.Default()
	}
	query := NewGovernorQuery().SetPagination(p)
	var govStatus []*GovStatus
	var err error
	if usePostgres {
		govStatus, err = s.postgresRepo.FindGovernorStatus(ctx, query)
	} else {
		govStatus, err = s.mongoRepo.FindGovernorStatus(ctx, query)
	}

	if err != nil {
		return nil, err
	}

	res := response.Response[[]*GovStatus]{Data: govStatus}
	return &res, err
}

// FindGovernorStatusByGuardianAddress get a governor status by guardianAddress.
func (s *Service) FindGovernorStatusByGuardianAddress(
	ctx context.Context,
	usePostgres bool,
	guardianAddress *types.Address,
	p *pagination.Pagination,
) (*response.Response[*GovStatus], error) {

	query := NewGovernorQuery().
		SetID(guardianAddress).
		SetPagination(p)

	var govStatus *GovStatus
	var err error
	if usePostgres {
		govStatus, err = s.postgresRepo.FindOneGovernorStatus(ctx, query)
	} else {
		govStatus, err = s.mongoRepo.FindOneGovernorStatus(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	res := response.Response[*GovStatus]{Data: govStatus}
	return &res, err
}

// FindNotionalLimit get a notional limit for each chainID.
func (s *Service) FindNotionalLimit(ctx context.Context, usePostgres bool, p *pagination.Pagination) (*response.Response[[]*NotionalLimit], error) {
	if p == nil {
		p = pagination.Default()
	}
	query := QueryNotionalLimit().SetPagination(p)
	var notionalLimit []*NotionalLimit
	var err error
	if usePostgres {
		notionalLimit, err = s.postgresRepo.GetGovernorNotionalLimit(ctx, query)
	} else {
		notionalLimit, err = s.mongoRepo.FindNotionalLimit(ctx, query)
	}
	res := response.Response[[]*NotionalLimit]{Data: notionalLimit}
	return &res, err
}

// GetNotionalLimitByChainID get a notional limit by chainID.
func (s *Service) GetNotionalLimitByChainID(ctx context.Context, usePostgres bool, p *pagination.Pagination, chainID vaa.ChainID) (*response.Response[[]*NotionalLimitDetail], error) {
	query := QueryNotionalLimit().SetPagination(p).SetChain(chainID)

	var notionalLimit []*NotionalLimitDetail
	var err error
	if usePostgres {
		notionalLimit, err = s.postgresRepo.GetNotionalLimitByChainID(ctx, query)
	} else {
		notionalLimit, err = s.mongoRepo.GetNotionalLimitByChainID(ctx, query)
	}

	res := response.Response[[]*NotionalLimitDetail]{Data: notionalLimit}
	return &res, err
}

// GetAvailableNotional get a available notional for each chainID.
func (s *Service) GetAvailableNotional(ctx context.Context, p *pagination.Pagination) (*response.Response[[]*NotionalAvailable], error) {
	if p == nil {
		p = pagination.Default()
	}
	query := QueryNotionalLimit().SetPagination(p)
	notionalAvailability, err := s.mongoRepo.GetAvailableNotional(ctx, query)
	res := response.Response[[]*NotionalAvailable]{Data: notionalAvailability}
	return &res, err
}

// GetAvailableNotionalByChainID get a available notional by chainID.
func (s *Service) GetAvailableNotionalByChainID(ctx context.Context, usePostgres bool, p *pagination.Pagination, chainID vaa.ChainID) (*response.Response[[]*NotionalAvailableDetail], error) {
	// check if chainID is valid
	if _, ok := s.supportedChainIDs[chainID]; !ok {
		return nil, errs.ErrNotFound
	}
	query := QueryNotionalLimit().SetPagination(p).SetChain(chainID)
	var notionalAvailability []*NotionalAvailableDetail
	var err error

	if usePostgres {
		notionalAvailability, err = s.postgresRepo.GetAvailableNotionalByChainID(ctx, query)
	} else {
		notionalAvailability, err = s.mongoRepo.GetAvailableNotionalByChainID(ctx, query)
	}
	res := response.Response[[]*NotionalAvailableDetail]{Data: notionalAvailability}
	return &res, err
}

// GetMaxNotionalAvailableByChainID get a maximun notional value by chainID.
func (s *Service) GetMaxNotionalAvailableByChainID(ctx context.Context, chainID vaa.ChainID) (*response.Response[*MaxNotionalAvailableRecord], error) {
	// check if chainID is valid
	if _, ok := s.supportedChainIDs[chainID]; !ok {
		return nil, errs.ErrNotFound
	}
	query := QueryNotionalLimit().SetChain(chainID)
	maxNotionaLAvailable, err := s.mongoRepo.GetMaxNotionalAvailableByChainID(ctx, query)
	res := response.Response[*MaxNotionalAvailableRecord]{Data: maxNotionaLAvailable}
	return &res, err
}

// GetEnqueueVaas get all the enqueued vaa.
func (s *Service) GetEnqueueVass(ctx context.Context, p *pagination.Pagination) (*response.Response[[]*EnqueuedVaas], error) {
	if p == nil {
		p = pagination.Default()
	}
	query := QueryEnqueuedVaa().SetPagination(p)
	enqueuedVaaResponse, err := s.mongoRepo.GetEnqueueVass(ctx, query)
	res := response.Response[[]*EnqueuedVaas]{Data: enqueuedVaaResponse}
	return &res, err
}

// GetEnqueueVassByChainID get enequeued vaa by chainID.
func (s *Service) GetEnqueueVassByChainID(ctx context.Context, p *pagination.Pagination, chainID vaa.ChainID) (*response.Response[[]*EnqueuedVaaDetail], error) {
	if p == nil {
		p = pagination.Default()
	}
	query := QueryEnqueuedVaa().SetPagination(p).SetChain(chainID)
	enqueuedVaaRecord, err := s.mongoRepo.GetEnqueueVassByChainID(ctx, query)
	res := response.Response[[]*EnqueuedVaaDetail]{Data: enqueuedVaaRecord}
	return &res, err
}

// GetGovernorLimit get governor limit.
func (s *Service) GetGovernorLimit(ctx context.Context, usePostgres bool, p *pagination.Pagination) (*response.Response[[]*GovernorLimit], error) {
	if p == nil {
		p = pagination.Default()
	}
	query := NewGovernorQuery().SetPagination(p)
	var governorLimit []*GovernorLimit
	var err error
	if usePostgres {
		governorLimit, err = s.postgresRepo.GetGovernorLimit(ctx, query)
	} else {
		governorLimit, err = s.mongoRepo.GetGovernorLimit(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	res := response.Response[[]*GovernorLimit]{Data: governorLimit}
	return &res, err
}

// GetAvailNotionByChain get governor limit for each chainID.
// Guardian api migration.
func (s *Service) GetAvailNotionByChain(ctx context.Context) ([]*AvailableNotionalByChain, error) {
	key := availableNotionByChain
	return cacheable.GetOrLoad(ctx, s.logger, s.cache, 1*time.Minute, key, s.metrics,
		func() ([]*AvailableNotionalByChain, error) {
			return s.mongoRepo.GetAvailNotionByChain(ctx)
		})
}

// Get governor token list.
// Guardian api migration.
func (s *Service) GetTokenList(ctx context.Context) ([]*TokenList, error) {
	key := tokenList
	return cacheable.GetOrLoad(ctx, s.logger, s.cache, 1*time.Minute, key, s.metrics,
		func() ([]*TokenList, error) {
			return s.mongoRepo.GetTokenList(ctx)
		})

}

// GetEnqueuedVaas get enqueued vaas.
// Guardian api migration.
func (s *Service) GetEnqueuedVaas(ctx context.Context) ([]*EnqueuedVaaItem, error) {
	entries, err := s.mongoRepo.GetEnqueuedVaas(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*EnqueuedVaaItem, 0)
	existingEnqueuedVaa := map[string]bool{}
	for _, e := range entries {
		// remove duplicates
		key := fmt.Sprintf("%s/%s/%s", e.EmitterChain, e.EmitterAddress, e.Sequence)
		if _, exists := existingEnqueuedVaa[key]; !exists {
			result = append(result, e)
			existingEnqueuedVaa[key] = true
		}
	}
	return result, nil
}

// IsVaaEnqueued check vaa is enqueued.
// Guardian api migration.
func (s *Service) IsVaaEnqueued(ctx context.Context, chainID vaa.ChainID, emitter *types.Address, seq string) (bool, error) {
	isEnqueued, err := s.mongoRepo.IsVaaEnqueued(ctx, chainID, emitter, seq)
	return isEnqueued, err
}

// GetGovernorVaas get enqueued vaas.
// Guardian api migration.
func (s *Service) GetGovernorVaas(ctx context.Context) ([]GovernorVaaDoc, error) {
	result, err := s.mongoRepo.GetGovernorVaas(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}
