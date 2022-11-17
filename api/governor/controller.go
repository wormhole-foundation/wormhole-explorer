// Package governor handle the request of governor data from governor endpoint defined in the api.
package governor

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/api/pagination"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	srv    *Service
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(serv *Service, logger *zap.Logger) *Controller {
	return &Controller{srv: serv, logger: logger.With(zap.String("module", "GovernorController"))}
}

// FindGovernorConfigurations handler for the endpoint /governor/config/
func (c *Controller) FindGovernorConfigurations(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	governorConfigs, err := c.srv.FindGovernorConfig(ctx.Context(), p)
	if err != nil {
		return err
	}
	return ctx.JSON(governorConfigs)
}

// FindGovernorConfigurationByGuardianAddress handler for the endpoint /governor/config/:guardian_address.
func (c *Controller) FindGovernorConfigurationByGuardianAddress(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	guardianAddress, err := middleware.ExtractGuardianAddress(ctx)
	if err != nil {
		return err
	}
	govConfig, err := c.srv.FindGovernorConfigByGuardianAddress(ctx.Context(), guardianAddress, p)
	if err != nil {
		return err
	}
	return ctx.JSON(govConfig)
}

// FindGovernorStatus handler for the endpoint /governor/status/.
func (c *Controller) FindGovernorStatus(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	governorStatus, err := c.srv.FindGovernorStatus(ctx.Context(), p)
	if err != nil {
		return err
	}
	return ctx.JSON(governorStatus)
}

// FindGovernorStatusByGuardianAddress handler for the endpoint /governor/status/:guardian_address.
func (c *Controller) FindGovernorStatusByGuardianAddress(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	guardianAddress, err := middleware.ExtractGuardianAddress(ctx)
	if err != nil {
		return err
	}
	govStatus, err := c.srv.FindGovernorStatusByGuardianAddress(ctx.Context(), guardianAddress, p)
	if err != nil {
		return err
	}
	return ctx.JSON(govStatus)
}

// GetGovernorLimit handler for the endpoint /governor/limit/
func (c *Controller) GetGovernorLimit(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	governorLimit, err := c.srv.GetGovernorLimit(ctx.Context(), p)
	if err != nil {
		return err
	}
	return ctx.JSON(governorLimit)
}

// FindNotionalLimit handler for the endpoint governor/notional/limit/
func (c *Controller) FindNotionalLimit(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	notionalLimit, err := c.srv.FindNotionalLimit(ctx.Context(), p)
	if err != nil {
		return err
	}
	return ctx.JSON(notionalLimit)
}

// GetNotionalLimitByChainID handler for the endpoint governor/notional/limit/:chain.
func (c *Controller) GetNotionalLimitByChainID(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	chainID, err := middleware.ExtractChainID(ctx)
	if err != nil {
		return err
	}
	notionalLimit, err := c.srv.GetNotionalLimitByChainID(ctx.Context(), p, chainID)
	if err != nil {
		return err
	}
	return ctx.JSON(notionalLimit)
}

// GetAvailableNotional handler for the endpoint governor/notional/available/
func (c *Controller) GetAvailableNotional(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	notionalAvaialabilies, err := c.srv.GetAvailableNotional(ctx.Context(), p)
	if err != nil {
		return err
	}
	return ctx.JSON(notionalAvaialabilies)
}

// GetAvailableNotionalByChainID handler for the endpoint governor/notional/available/:chain
func (c *Controller) GetAvailableNotionalByChainID(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	chainID, err := middleware.ExtractChainID(ctx)
	if err != nil {
		return err
	}
	response, err := c.srv.GetAvailableNotionalByChainID(ctx.Context(), p, chainID)
	if err != nil {
		return err
	}
	return ctx.JSON(response)
}

// GetMaxNotionalAvailableByChainID handler for the endpoint governor/max_available/:chain.
func (c *Controller) GetMaxNotionalAvailableByChainID(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	chainID, err := middleware.ExtractChainID(ctx)
	if err != nil {
		return err
	}
	response, err := c.srv.GetMaxNotionalAvailableByChainID(ctx.Context(), p, chainID)
	if err != nil {
		return err
	}
	return ctx.JSON(response)
}

// GetEnqueueVass handler for the endpoint governor/enqueued_vaas/
func (c *Controller) GetEnqueueVass(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	enqueuedVaas, err := c.srv.GetEnqueueVass(ctx.Context(), p)
	if err != nil {
		return err
	}
	return ctx.JSON(enqueuedVaas)
}

// GetEnqueueVassByChainID handler for the endpoint governor/enqueued_vaas/:chain.
func (c *Controller) GetEnqueueVassByChainID(ctx *fiber.Ctx) error {
	p := pagination.GetFromContext(ctx)
	chainID, err := middleware.ExtractChainID(ctx)
	if err != nil {
		return err
	}
	enqueuedVaas, err := c.srv.GetEnqueueVassByChainID(ctx.Context(), p, chainID)
	if err != nil {
		return err
	}
	return ctx.JSON(enqueuedVaas)
}
