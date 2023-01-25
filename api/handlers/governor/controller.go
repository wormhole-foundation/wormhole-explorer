// Package governor handle the request of governor data from governor endpoint defined in the api.
package governor

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	_ "github.com/wormhole-foundation/wormhole-explorer/api/response" // needed by swaggo docs
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
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
	p := middleware.GetPaginationFromContext(ctx)
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
	p := middleware.GetPaginationFromContext(ctx)
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
	p := middleware.GetPaginationFromContext(ctx)
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
	p := middleware.GetPaginationFromContext(ctx)
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
	p := middleware.GetPaginationFromContext(ctx)
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

// GetEnqueueVaas godoc
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
func (c *Controller) GetEnqueueVaas(ctx *fiber.Ctx) error {
	p := middleware.GetPaginationFromContext(ctx)
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
// @Description Get enqueued VAAs
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
