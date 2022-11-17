// Package governor handle the request of governor data from governor endpoint defined in the api.
package governor

import (
	"context"

	"github.com/certusone/wormhole/node/pkg/vaa"
	"github.com/wormhole-foundation/wormhole-explorer/api/pagination"
	"github.com/wormhole-foundation/wormhole-explorer/api/services"
	"go.uber.org/zap"
)

type Service struct {
	repo   *Repository
	logger *zap.Logger
}

// NewService create a new governor.Service.
func NewService(dao *Repository, logger *zap.Logger) *Service {
	return &Service{repo: dao, logger: logger.With(zap.String("module", "GovernorService"))}
}

// FindGovernorConfig get a list of governor configurations.
func (s *Service) FindGovernorConfig(ctx context.Context, p *pagination.Pagination) (*services.Response[[]*GovConfig], error) {
	if p == nil {
		p = pagination.FirstPage()
	}
	query := QueryGovernor().SetPagination(p)
	govConfigs, err := s.repo.FindGovConfigurations(ctx, query)
	res := services.Response[[]*GovConfig]{Data: govConfigs}
	return &res, err
}

// FindGovernorConfigByGuardianAddress get a governor configuration by guardianAddress.
func (s *Service) FindGovernorConfigByGuardianAddress(ctx context.Context, guardianAddress string, p *pagination.Pagination) (*services.Response[*GovConfig], error) {
	query := QueryGovernor().SetID(guardianAddress).SetPagination(p)
	govConfig, err := s.repo.FindGovConfiguration(ctx, query)
	res := services.Response[*GovConfig]{Data: govConfig}
	return &res, err
}

// FindGovernorStatus get a list of governor status.
func (s *Service) FindGovernorStatus(ctx context.Context, p *pagination.Pagination) (*services.Response[[]*GovStatus], error) {
	if p == nil {
		p = pagination.FirstPage()
	}
	query := QueryGovernor().SetPagination(p)
	govStatus, err := s.repo.FindGovernorStatus(ctx, query)
	res := services.Response[[]*GovStatus]{Data: govStatus}
	return &res, err
}

// FindGovernorStatusByGuardianAddress get a governor status by guardianAddress.
func (s *Service) FindGovernorStatusByGuardianAddress(ctx context.Context, guardianAddress string, p *pagination.Pagination) (*services.Response[*GovStatus], error) {
	query := QueryGovernor().SetID(guardianAddress).SetPagination(p)
	govStatus, err := s.repo.FindOneGovernorStatus(ctx, query)
	res := services.Response[*GovStatus]{Data: govStatus}
	return &res, err
}

// FindNotionalLimit get a notional limit for each chainID.
func (s *Service) FindNotionalLimit(ctx context.Context, p *pagination.Pagination) (*services.Response[[]*NotionalLimit], error) {
	if p == nil {
		p = pagination.FirstPage()
	}
	query := QueryNotionalLimit().SetPagination(p)
	notionalLimit, err := s.repo.FindNotionalLimit(ctx, query)
	res := services.Response[[]*NotionalLimit]{Data: notionalLimit}
	return &res, err
}

// GetNotionalLimitByChainID get a notional limit by chainID.
func (s *Service) GetNotionalLimitByChainID(ctx context.Context, p *pagination.Pagination, chainID vaa.ChainID) (*services.Response[[]*NotionalLimitDetail], error) {
	query := QueryNotionalLimit().SetPagination(p).SetChain(chainID)
	notionalLimit, err := s.repo.GetNotionalLimitByChainID(ctx, query)
	res := services.Response[[]*NotionalLimitDetail]{Data: notionalLimit}
	return &res, err
}

// GetAvailableNotional get a available notional for each chainID.
func (s *Service) GetAvailableNotional(ctx context.Context, p *pagination.Pagination) (*services.Response[[]*NotionalAvailable], error) {
	if p == nil {
		p = pagination.FirstPage()
	}
	query := QueryNotionalLimit().SetPagination(p)
	notionalAvailability, err := s.repo.GetAvailableNotional(ctx, query)
	res := services.Response[[]*NotionalAvailable]{Data: notionalAvailability}
	return &res, err
}

// GetAvailableNotionalByChainID get a available notional by chainID.
func (s *Service) GetAvailableNotionalByChainID(ctx context.Context, p *pagination.Pagination, chainID vaa.ChainID) (*services.Response[[]*NotionalAvailableDetail], error) {
	query := QueryNotionalLimit().SetPagination(p).SetChain(chainID)
	notionaLAvailability, err := s.repo.GetAvailableNotionalByChainID(ctx, query)
	res := services.Response[[]*NotionalAvailableDetail]{Data: notionaLAvailability}
	return &res, err
}

// GetMaxNotionalAvailableByChainID get a maximun notional value by chainID.
func (s *Service) GetMaxNotionalAvailableByChainID(ctx context.Context, p *pagination.Pagination, chainID vaa.ChainID) (*services.Response[*MaxNotionalAvailableRecord], error) {
	query := QueryNotionalLimit().SetPagination(p).SetChain(chainID)
	maxNotionaLAvailable, err := s.repo.GetMaxNotionalAvailableByChainID(ctx, query)
	res := services.Response[*MaxNotionalAvailableRecord]{Data: maxNotionaLAvailable}
	return &res, err
}

// GetEnqueueVaas get all the enqueued vaa.
func (s *Service) GetEnqueueVass(ctx context.Context, p *pagination.Pagination) (*services.Response[[]*EnqueuedVaas], error) {
	if p == nil {
		p = pagination.FirstPage()
	}
	query := QueryEnqueuedVaa().SetPagination(p)
	enqueuedVaaResponse, err := s.repo.GetEnqueueVass(ctx, query)
	res := services.Response[[]*EnqueuedVaas]{Data: enqueuedVaaResponse}
	return &res, err
}

// GetEnqueueVassByChainID get enequeued vaa by chainID.
func (s *Service) GetEnqueueVassByChainID(ctx context.Context, p *pagination.Pagination, chainID vaa.ChainID) (*services.Response[[]*EnqueuedVaaDetail], error) {
	if p == nil {
		p = pagination.FirstPage()
	}
	query := QueryEnqueuedVaa().SetPagination(p).SetChain(chainID)
	enqueuedVaaRecord, err := s.repo.GetEnqueueVassByChainID(ctx, query)
	res := services.Response[[]*EnqueuedVaaDetail]{Data: enqueuedVaaRecord}
	return &res, err
}

// GetGovernorLimit get governor limit.
func (s *Service) GetGovernorLimit(ctx context.Context, p *pagination.Pagination) (*services.Response[[]*GovernorLimit], error) {
	if p == nil {
		p = pagination.FirstPage()
	}
	query := QueryGovernor().SetPagination(p)
	governorLimit, err := s.repo.GetGovernorLimit(ctx, query)
	res := services.Response[[]*GovernorLimit]{Data: governorLimit}
	return &res, err
}
