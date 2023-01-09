// Package governor handle the request of governor data from governor endpoint defined in the api.
package governor

import (
	"strconv"

	"github.com/certusone/wormhole/node/pkg/vaa"
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
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
	p := middleware.GetPaginationFromContext(ctx)
	governorConfigs, err := c.srv.FindGovernorConfig(ctx.Context(), p)
	if err != nil {
		return err
	}
	return ctx.JSON(governorConfigs)
}

// FindGovernorConfigurationByGuardianAddress handler for the endpoint /governor/config/:guardian_address.
func (c *Controller) FindGovernorConfigurationByGuardianAddress(ctx *fiber.Ctx) error {
	p := middleware.GetPaginationFromContext(ctx)
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

// FindGovernorStatus handler for the endpoint /governor/status/.
func (c *Controller) FindGovernorStatus(ctx *fiber.Ctx) error {
	p := middleware.GetPaginationFromContext(ctx)
	governorStatus, err := c.srv.FindGovernorStatus(ctx.Context(), p)
	if err != nil {
		return err
	}
	return ctx.JSON(governorStatus)
}

// FindGovernorStatusByGuardianAddress handler for the endpoint /governor/status/:guardian_address.
func (c *Controller) FindGovernorStatusByGuardianAddress(ctx *fiber.Ctx) error {
	p := middleware.GetPaginationFromContext(ctx)
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

// GetGovernorLimit handler for the endpoint /governor/limit/
func (c *Controller) GetGovernorLimit(ctx *fiber.Ctx) error {
	p := middleware.GetPaginationFromContext(ctx)
	governorLimit, err := c.srv.GetGovernorLimit(ctx.Context(), p)
	if err != nil {
		return err
	}
	return ctx.JSON(governorLimit)
}

// FindNotionalLimit handler for the endpoint governor/notional/limit/
func (c *Controller) FindNotionalLimit(ctx *fiber.Ctx) error {
	p := middleware.GetPaginationFromContext(ctx)
	notionalLimit, err := c.srv.FindNotionalLimit(ctx.Context(), p)
	if err != nil {
		return err
	}
	return ctx.JSON(notionalLimit)
}

// GetNotionalLimitByChainID handler for the endpoint governor/notional/limit/:chain.
func (c *Controller) GetNotionalLimitByChainID(ctx *fiber.Ctx) error {
	p := middleware.GetPaginationFromContext(ctx)
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

// GetAvailableNotional handler for the endpoint governor/notional/available/
func (c *Controller) GetAvailableNotional(ctx *fiber.Ctx) error {
	p := middleware.GetPaginationFromContext(ctx)
	notionalAvaialabilies, err := c.srv.GetAvailableNotional(ctx.Context(), p)
	if err != nil {
		return err
	}
	return ctx.JSON(notionalAvaialabilies)
}

// GetAvailableNotionalByChainID handler for the endpoint governor/notional/available/:chain
func (c *Controller) GetAvailableNotionalByChainID(ctx *fiber.Ctx) error {
	p := middleware.GetPaginationFromContext(ctx)
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

// GetMaxNotionalAvailableByChainID handler for the endpoint governor/max_available/:chain.
func (c *Controller) GetMaxNotionalAvailableByChainID(ctx *fiber.Ctx) error {
	p := middleware.GetPaginationFromContext(ctx)
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

// GetEnqueueVass handler for the endpoint governor/enqueued_vaas/
func (c *Controller) GetEnqueueVass(ctx *fiber.Ctx) error {
	p := middleware.GetPaginationFromContext(ctx)
	enqueuedVaas, err := c.srv.GetEnqueueVass(ctx.Context(), p)
	if err != nil {
		return err
	}
	return ctx.JSON(enqueuedVaas)
}

// GetEnqueueVassByChainID handler for the endpoint governor/enqueued_vaas/:chain.
func (c *Controller) GetEnqueueVassByChainID(ctx *fiber.Ctx) error {
	p := middleware.GetPaginationFromContext(ctx)
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

// AvailableNotionalResponse response compatible with grpc api.
type AvailableNotionalResponse struct {
	Entries []*AvailableNotionalItemResponse `json:"entries"`
}

type AvailableNotionalItemResponse struct {
	ChainID            vaa.ChainID `json:"chainId"`
	AvailableNotional  string      `json:"remainingAvailableNotional"`
	NotionalLimit      string      `json:"notionalLimit"`
	MaxTransactionSize string      `json:"bigTransactionSize"`
}

// GetAvailNotionByChain godoc
// @Description Get available notional by chainID
// @Description Since from the wormhole-explorer point of view it is not a node, but has the information of all nodes,
// @Description in order to build the endpoints it was assumed:
// @Description There are N number of remainingAvailableNotional values in the GovernorConfig collection. N = number of guardians
// @Description for a chainID. The smallest remainingAvailableNotional value for a chainID is used for the endpoint response.
// @Tags Guardian
// @ID governor-available-notional-by-chain
// @Success 200 {object} AvailableNotionalResponse
// @Failure 400
// @Failure 500
// @Router /v1/governor/available_notional_by_chain [get]
func (c *Controller) GetAvailNotionByChain(ctx *fiber.Ctx) error {
	// call service to get available notional by chainID
	availableNotional, err := c.srv.GetAvailNotionByChain(ctx.Context())
	if err != nil {
		return err
	}

	// build response compatible with node grpc api.
	entries := make([]*AvailableNotionalItemResponse, 0, len(availableNotional))
	for _, v := range availableNotional {
		r := AvailableNotionalItemResponse{
			ChainID:            v.ChainID,
			AvailableNotional:  v.AvailableNotional.String(),
			NotionalLimit:      v.NotionalLimit.String(),
			MaxTransactionSize: v.MaxTransactionSize.String(),
		}
		entries = append(entries, &r)
	}
	response := AvailableNotionalResponse{
		Entries: entries,
	}
	return ctx.JSON(response)
}

// AvailableNotionalResponse response compatible with grpc api.
type EnqueuedVaaResponse struct {
	Entries []*EnqueuedVaaItemResponse `json:"entries"`
}

type EnqueuedVaaItemResponse struct {
	EmitterChain   vaa.ChainID `json:"emitterChain"`
	EmitterAddress string      `json:"emitterAddress"`
	Sequence       uint64      `json:"sequence"`
	ReleaseTime    int64       `json:"releaseTime"`
	NotionalValue  string      `json:"notionalValue"`
	TxHash         string      `json:"txHash"`
}

// GetEnqueuedVaas godoc
// @Description Get enqueued vaa's
// @Tags Guardian
// @ID guardians-enqueued-vaas
// @Success 200 {object} EnqueuedVaaResponse
// @Failure 400
// @Failure 500
// @Router /v1/governor/enqueued_vaas [get]
func (c *Controller) GetEnqueuedVaas(ctx *fiber.Ctx) error {
	enqueuedVaa, err := c.srv.GetEnqueuedVaas(ctx.Context())
	if err != nil {
		return err
	}

	// build response compatible with node grpc api.
	entries := make([]*EnqueuedVaaItemResponse, 0, len(enqueuedVaa))
	for _, v := range enqueuedVaa {
		seqUint64, err := strconv.ParseUint(v.Sequence, 10, 64)
		if err != nil {
			return err
		}
		r := EnqueuedVaaItemResponse{
			EmitterChain:   v.EmitterChain,
			EmitterAddress: v.EmitterAddress,
			Sequence:       seqUint64,
			ReleaseTime:    v.ReleaseTime,
			NotionalValue:  v.NotionalValue.String(),
			TxHash:         v.TxHash,
		}
		entries = append(entries, &r)
	}
	response := EnqueuedVaaResponse{
		Entries: entries,
	}

	return ctx.JSON(response)
}

// IsVaaEnqueued godoc
// @Description Check if vaa is enqueued
// @Tags Guardian
// @ID guardians-is-vaa-enqueued
// @Param chain_id path integer true "id of the blockchain"
// @Param emitter path string true "address of the emitter"
// @Param seq path integer true "sequence of the vaa"
// @Success 200 {object} EnqueuedVaaResponse
// @Failure 400
// @Failure 500
// @Router /v1/governor/is_vaa_enqueued/{chain_id}/{emitter}/{seq} [get]
func (c *Controller) IsVaaEnqueued(ctx *fiber.Ctx) error {
	chainID, emitter, seq, err := middleware.ExtractVAAParams(ctx, c.logger)
	if err != nil {
		return err
	}
	isEnqueued, err := c.srv.IsVaaEnqueued(ctx.Context(), chainID, *emitter, strconv.FormatUint(seq, 10))
	if err != nil {
		return err
	}

	// build reponse compatible with node grpc api.
	response := struct {
		IsEnqueued bool `json:"isEnqueued"`
	}{
		IsEnqueued: isEnqueued,
	}
	return ctx.JSON(response)
}

// GetTokenList godoc
// @Description Get token list
// @Description Since from the wormhole-explorer point of view it is not a node, but has the information of all nodes,
// @Description in order to build the endpoints it was assumed:
// @Description For tokens with the same originChainId and originAddress and different price values for each node,
// @Description the price that has most occurrences in all the nodes for an originChainId and originAddress is returned.
// @Tags Guardian
// @ID guardians-token-list
// @Success 200 {object} []TokenList
// @Failure 400
// @Failure 500
// @Router /v1/governor/token_list [get]
func (c *Controller) GetTokenList(ctx *fiber.Ctx) error {
	tokenList, err := c.srv.GetTokenList(ctx.Context())
	if err != nil {
		return err
	}

	// build reponse compatible with node grpc api.
	response := struct {
		Entries []*TokenList `json:"entries"`
	}{
		Entries: tokenList,
	}
	return ctx.JSON(response)
}
