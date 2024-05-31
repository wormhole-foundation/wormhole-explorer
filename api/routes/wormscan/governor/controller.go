// Package governor handle the request of governor data from governor endpoint defined in the api.
package governor

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/governor"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	_ "github.com/wormhole-foundation/wormhole-explorer/api/response" // needed by swaggo docs
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	srv    *governor.Service
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(serv *governor.Service, logger *zap.Logger) *Controller {
	return &Controller{srv: serv, logger: logger.With(zap.String("module", "GovernorController"))}
}

// FindGovernorConfigurations godoc
// @Description Returns governor configuration for all guardians.
// @Tags wormholescan
// @ID governor-config
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Success 200 {object} response.Response[governor.GovConfig]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/config [get]
func (c *Controller) FindGovernorConfigurations(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	// Check pagination max limit
	if p.Limit > 1000 {
		return response.NewInvalidParamError(ctx, "pageSize cannot be greater than 1000", nil)
	}

	governorConfigs, err := c.srv.FindGovernorConfig(ctx.Context(), p)
	if err != nil {
		return err
	}

	return ctx.JSON(governorConfigs)
}

// FindGovernorConfigurationByGuardianAddress godoc
// @Description Returns governor configuration for a given guardian.
// @Tags wormholescan
// @ID governor-config-by-guardian-address
// @Success 200 {object} response.Response[governor.GovConfig]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/config/:guardian_address [get]
func (c *Controller) FindGovernorConfigurationByGuardianAddress(ctx *fiber.Ctx) error {

	// extract query params
	guardianAddress, err := middleware.ExtractGuardianAddress(ctx, c.logger)
	if err != nil {
		return err
	}

	// query the database
	govConfigs, err := c.srv.FindGovernorConfigByGuardianAddress(ctx.Context(), guardianAddress)
	if err != nil {
		return err
	} else if len(govConfigs) == 0 {
		return response.NewNotFoundError(ctx)
	} else if len(govConfigs) > 1 {
		err = fmt.Errorf("expected at most 1 governor config, but found %d", len(govConfigs))
		return response.NewInternalError(ctx, err)
	}

	// populate the response struct and return
	res := response.Response[*governor.GovConfig]{Data: govConfigs[0]}
	return ctx.JSON(res)
}

// FindGovernorStatus godoc
// @Description Returns the governor status for all guardians.
// @Tags wormholescan
// @ID governor-status
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Success 200 {object} response.Response[[]governor.GovStatus]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/status [get]
func (c *Controller) FindGovernorStatus(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	// Check pagination max limit
	if p.Limit > 1000 {
		return response.NewInvalidParamError(ctx, "pageSize cannot be greater than 1000", nil)
	}

	governorStatus, err := c.srv.FindGovernorStatus(ctx.Context(), p)
	if err != nil {
		return err
	}

	return ctx.JSON(governorStatus)
}

// FindGovernorStatusByGuardianAddress godoc
// @Description Returns the governor status for a given guardian.
// @Tags wormholescan
// @ID governor-status-by-guardian-address
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Success 200 {object} response.Response[governor.GovStatus]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/status/:guardian_address [get]
func (c *Controller) FindGovernorStatusByGuardianAddress(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	// Check pagination max limit
	if p.Limit > 1000 {
		return response.NewInvalidParamError(ctx, "pageSize cannot be greater than 1000", nil)
	}

	guardianAddress, err := middleware.ExtractGuardianAddress(ctx, c.logger)
	if err != nil {
		return err
	}

	govStatus, err := c.srv.FindGovernorStatusByGuardianAddress(ctx.Context(), guardianAddress, p)
	if err != nil {
		return err
	}

	return ctx.JSON(govStatus)
}

// GetGovernorLimit godoc
// @Description Returns the governor limit for all blockchains.
// @Tags wormholescan
// @ID governor-notional-limit
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Success 200 {object} response.Response[[]governor.GovernorLimit]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/limit [get]
func (c *Controller) GetGovernorLimit(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	// Check pagination max limit
	if p.Limit > 1000 {
		return response.NewInvalidParamError(ctx, "pageSize cannot be greater than 1000", nil)
	}

	governorLimit, err := c.srv.GetGovernorLimit(ctx.Context(), p)
	if err != nil {
		return err
	}

	return ctx.JSON(governorLimit)
}

// FindNotionalLimit godoc
// @Description Returns the detailed notional limit for all blockchains.
// @Tags wormholescan
// @ID governor-notional-limit-detail
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Success 200 {object} response.Response[[]governor.NotionalLimitDetail]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/notional/limit [get]
func (c *Controller) FindNotionalLimit(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	// Check pagination max limit
	if p.Limit > 1000 {
		return response.NewInvalidParamError(ctx, "pageSize cannot be greater than 1000", nil)
	}

	notionalLimit, err := c.srv.FindNotionalLimit(ctx.Context(), p)
	if err != nil {
		return err
	}

	return ctx.JSON(notionalLimit)
}

// GetNotionalLimitByChainID godoc
// @Description Returns the detailed notional limit available for a given blockchain.
// @Tags wormholescan
// @ID governor-notional-limit-detail-by-chain
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Success 200 {object} response.Response[[]governor.NotionalLimitDetail]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/notional/limit/:chain [get]
func (c *Controller) GetNotionalLimitByChainID(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	// Check pagination max limit
	if p.Limit > 1000 {
		return response.NewInvalidParamError(ctx, "pageSize cannot be greater than 1000", nil)
	}

	chainID, err := middleware.ExtractChainID(ctx, c.logger)
	if err != nil {
		return err
	}

	notionalLimit, err := c.srv.GetNotionalLimitByChainID(ctx.Context(), p, chainID)
	if err != nil {
		return err
	}

	return ctx.JSON(notionalLimit)
}

// GetAvailableNotional godoc
// @Description Returns the amount of notional value available for each blockchain.
// @Tags wormholescan
// @ID governor-notional-available
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} response.Response[[]governor.NotionalAvailable]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/notional/available [get]
func (c *Controller) GetAvailableNotional(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	// Check pagination max limit
	if p.Limit > 1000 {
		return response.NewInvalidParamError(ctx, "pageSize cannot be greater than 1000", nil)
	}

	notionalAvaialabilies, err := c.srv.GetAvailableNotional(ctx.Context(), p)
	if err != nil {
		return err
	}

	return ctx.JSON(notionalAvaialabilies)
}

// GetAvailableNotionalByChainID godoc
// @Description Returns the amount of notional value available for a given blockchain.
// @Tags wormholescan
// @ID governor-notional-available-by-chain
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Success 200 {object} response.Response[[]governor.NotionalAvailableDetail]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/notional/available/:chain [get]
func (c *Controller) GetAvailableNotionalByChainID(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	// Check pagination max limit
	if p.Limit > 1000 {
		return response.NewInvalidParamError(ctx, "pageSize cannot be greater than 1000", nil)
	}

	chainID, err := middleware.ExtractChainID(ctx, c.logger)
	if err != nil {
		return err
	}

	response, err := c.srv.GetAvailableNotionalByChainID(ctx.Context(), p, chainID)
	if err != nil {
		return err
	}

	return ctx.JSON(response)
}

// GetMaxNotionalAvailableByChainID godoc
// @Description Returns the maximum amount of notional value available for a given blockchain.
// @Tags wormholescan
// @ID governor-max-notional-available-by-chain
// @Success 200 {object} response.Response[governor.MaxNotionalAvailableRecord]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/notional/max_available/:chain [get]
func (c *Controller) GetMaxNotionalAvailableByChainID(ctx *fiber.Ctx) error {

	chainID, err := middleware.ExtractChainID(ctx, c.logger)
	if err != nil {
		return err
	}

	response, err := c.srv.GetMaxNotionalAvailableByChainID(ctx.Context(), chainID)
	if err != nil {
		return err
	}

	return ctx.JSON(response)
}

// GetEnqueuedVaas godoc
// @Description Returns enqueued VAAs for each blockchain.
// @Tags wormholescan
// @ID governor-enqueued-vaas
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} response.Response[[]governor.EnqueuedVaas]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/enqueued_vaas/ [get]
func (c *Controller) GetEnqueuedVaas(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	// Check pagination max limit
	if p.Limit > 1000 {
		return response.NewInvalidParamError(ctx, "pageSize cannot be greater than 1000", nil)
	}

	enqueuedVaas, err := c.srv.GetEnqueueVass(ctx.Context(), p)
	if err != nil {
		return err
	}

	return ctx.JSON(enqueuedVaas)
}

// GetEnqueuedVaasByChainID godoc
// @Description Returns all enqueued VAAs for a given blockchain.
// @Tags wormholescan
// @ID guardians-enqueued-vaas-by-chain
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} response.Response[[]governor.EnqueuedVaaDetail]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/enqueued_vaas/:chain [get]
func (c *Controller) GetEnqueuedVaasByChainID(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	// Check pagination max limit
	if p.Limit > 1000 {
		return response.NewInvalidParamError(ctx, "pageSize cannot be greater than 1000", nil)
	}
	chainID, err := middleware.ExtractChainID(ctx, c.logger)
	if err != nil {
		return err
	}

	enqueuedVaas, err := c.srv.GetEnqueueVassByChainID(ctx.Context(), p, chainID)
	if err != nil {
		return err
	}

	return ctx.JSON(enqueuedVaas)
}

// GetGovernorVaas godoc
// @Description Returns all vaas in Governor.
// @Tags wormholescan
// @ID governor-vaas
// @Success 200 {object} response.Response[[]governor.GovernorVaasResponse]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/vaas [get]
func (c *Controller) GetGovernorVaas(ctx *fiber.Ctx) error {
	enqueuedVaas, err := c.srv.GetGovernorVaas(ctx.Context())
	if err != nil {
		return err
	}

	result := make([]GovernorVaasResponse, 0)
	for _, v := range enqueuedVaas {
		status := "pending"
		if len(v.Vaas) > 0 {
			status = "issued"
		}
		result = append(result, GovernorVaasResponse{
			VaaID:          v.ID,
			ChainID:        v.ChainID,
			EmitterAddress: v.EmitterAddress,
			Sequence:       v.Sequence,
			TxHash:         v.TxHash,
			ReleaseTime:    v.ReleaseTime,
			Amount:         uint64(v.Amount),
			Status:         status,
		})
	}

	return ctx.JSON(result)
}
