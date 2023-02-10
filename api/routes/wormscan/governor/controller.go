// Package governor handle the request of governor data from governor endpoint defined in the api.
package governor

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/governor"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
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
// @Tags Wormscan
// @ID governor-config
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} response.Response[GovConfig]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/config [get]
func (c *Controller) FindGovernorConfigurations(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	governorConfigs, err := c.srv.FindGovernorConfig(ctx.Context(), p)
	if err != nil {
		return err
	}

	return ctx.JSON(governorConfigs)
}

// FindGovernorConfigurationByGuardianAddress godoc
// @Description Returns governor configuration for a given guardian.
// @Tags Wormscan
// @ID governor-config-by-guardian-address
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} response.Response[[]GovConfig]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/config/:guardian_address [get]
func (c *Controller) FindGovernorConfigurationByGuardianAddress(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	guardianAddress, err := middleware.ExtractGuardianAddress(ctx, c.logger)
	if err != nil {
		return err
	}

	govConfig, err := c.srv.FindGovernorConfigByGuardianAddress(ctx.Context(), guardianAddress, p)
	if err != nil {
		return err
	}

	return ctx.JSON(govConfig)
}

// FindGovernorStatus godoc
// @Description Returns the governor status for all guardians.
// @Tags Wormscan
// @ID governor-status
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} response.Response[[]GovStatus]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/status [get]
func (c *Controller) FindGovernorStatus(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	governorStatus, err := c.srv.FindGovernorStatus(ctx.Context(), p)
	if err != nil {
		return err
	}

	return ctx.JSON(governorStatus)
}

// FindGovernorStatusByGuardianAddress godoc
// @Description Returns the governor status for a given guardian.
// @Tags Wormscan
// @ID governor-status-by-guardian-address
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} response.Response[GovStatus]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/status/:guardian_address [get]
func (c *Controller) FindGovernorStatusByGuardianAddress(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
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
// @Tags Wormscan
// @ID governor-notional-limit
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} response.Response[[]GovernorLimit]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/limit [get]
func (c *Controller) GetGovernorLimit(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	governorLimit, err := c.srv.GetGovernorLimit(ctx.Context(), p)
	if err != nil {
		return err
	}

	return ctx.JSON(governorLimit)
}

// FindNotionalLimit godoc
// @Description Returns the detailed notional limit for all blockchains.
// @Tags Wormscan
// @ID governor-notional-limit-detail
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} response.Response[[]NotionalLimitDetail]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/notional/limit [get]
func (c *Controller) FindNotionalLimit(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	notionalLimit, err := c.srv.FindNotionalLimit(ctx.Context(), p)
	if err != nil {
		return err
	}

	return ctx.JSON(notionalLimit)
}

// GetNotionalLimitByChainID godoc
// @Description Returns the detailed notional limit available for a given blockchain.
// @Tags Wormscan
// @ID governor-notional-limit-detail-by-chain
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} response.Response[[]NotionalLimitDetail]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/notional/limit/:chain [get]
func (c *Controller) GetNotionalLimitByChainID(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
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
// @Tags Wormscan
// @ID governor-notional-available
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} response.Response[[]NotionalAvailable]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/notional/available [get]
func (c *Controller) GetAvailableNotional(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	notionalAvaialabilies, err := c.srv.GetAvailableNotional(ctx.Context(), p)
	if err != nil {
		return err
	}

	return ctx.JSON(notionalAvaialabilies)
}

// GetAvailableNotionalByChainID godoc
// @Description Returns the amount of notional value available for a given blockchain.
// @Tags Wormscan
// @ID governor-notional-available-by-chain
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} response.Response[[]NotionalAvailableDetail]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/notional/available/:chain [get]
func (c *Controller) GetAvailableNotionalByChainID(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
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
// @Tags Wormscan
// @ID governor-max-notional-available-by-chain
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} response.Response[MaxNotionalAvailableRecord]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/max_available/:chain [get]
func (c *Controller) GetMaxNotionalAvailableByChainID(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	chainID, err := middleware.ExtractChainID(ctx, c.logger)
	if err != nil {
		return err
	}

	response, err := c.srv.GetMaxNotionalAvailableByChainID(ctx.Context(), p, chainID)
	if err != nil {
		return err
	}

	return ctx.JSON(response)
}

// GetEnqueuedVaas godoc
// @Description Returns enqueued VAAs for each blockchain.
// @Tags Wormscan
// @ID governor-enqueued-vaas
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} response.Response[[]EnqueuedVaas]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/enqueued_vaas/ [get]
func (c *Controller) GetEnqueuedVaas(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	enqueuedVaas, err := c.srv.GetEnqueueVass(ctx.Context(), p)
	if err != nil {
		return err
	}

	return ctx.JSON(enqueuedVaas)
}

// GetEnqueuedVaasByChainID godoc
// @Description Returns all enqueued VAAs for a given blockchain.
// @Tags Wormscan
// @ID guardians-enqueued-vaas-by-chain
// @Param page query integer false "Page number."
// @Param pageSize query integer false "Number of elements per page."
// @Param sortOrder query string false "Sort results in ascending or descending order." Enums(ASC, DESC)
// @Success 200 {object} response.Response[[]EnqueuedVaaDetail]
// @Failure 400
// @Failure 500
// @Router /api/v1/governor/enqueued_vaas/:chain [get]
func (c *Controller) GetEnqueuedVaasByChainID(ctx *fiber.Ctx) error {

	p, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
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
