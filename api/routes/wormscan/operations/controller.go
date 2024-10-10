package operations

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/operations"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"github.com/wormhole-foundation/wormhole-explorer/common/types"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// Controller is the controller for the operation resource.
type Controller struct {
	srv    operationService
	logger *zap.Logger
}

// decouple operations.Service from the controller in order to make it testable
type operationService interface {
	FindById(ctx context.Context, chainID vaa.ChainID, emitter *types.Address, seq string) (*operations.OperationDto, error)
	FindAll(ctx context.Context, filter operations.OperationFilter) ([]*operations.OperationDto, error)
}

// NewController create a new controler.
func NewController(operationService operationService, logger *zap.Logger) *Controller {
	return &Controller{
		srv:    operationService,
		logger: logger.With(zap.String("module", "OperationsController")),
	}
}

// FindAll godoc
// @Description Find all operations.
// @Tags wormholescan
// @ID get-operations
// @Param address query string false "address of the emitter"
// @Param txHash query string false "hash of the transaction"
// @Param page query integer false "page number"
// @Param pageSize query integer false "pageSize". Maximum value is 100.
// @Param sourceChain query string false "source chains of the operation, separated by comma".
// @Param targetChain query string false "target chains of the operation, separated by comma".
// @Param appId query string false "appID of the operation".
// @Param exclusiveAppId query boolean false "single appId of the operation".
// @Param from query string false "beginning of period"
// @Param to query string false "end of period"
// @Success 200 {object} []OperationResponse
// @Failure 400
// @Failure 500
// @Router /api/v1/operations [get]
func (c *Controller) FindAll(ctx *fiber.Ctx) error {
	// Extract query parameters
	pagination, err := middleware.ExtractPagination(ctx)
	if err != nil {
		return err
	}

	// Check pagination max limit
	if pagination.Limit > 100 {
		return response.NewInvalidParamError(ctx, "pageSize cannot be greater than 100", nil)
	}

	address := middleware.ExtractAddressFromQueryParams(ctx, c.logger)
	txHash, err := middleware.GetTxHash(ctx, c.logger)
	if err != nil {
		return err
	}

	searchByAddress := address != ""
	searchByTxHash := txHash != nil && txHash.String() != ""

	if searchByAddress && searchByTxHash {
		return response.NewInvalidParamError(ctx, "address and txHash cannot be used at the same time", nil)
	}

	sourceChain, err := middleware.ExtractSourceChain(ctx, c.logger)
	if err != nil {
		return err
	}

	targetChain, err := middleware.ExtractTargetChain(ctx, c.logger)
	if err != nil {
		return err
	}

	var appIDs []string
	appIDQueryParam := ctx.Query("appId")
	if appIDQueryParam != "" {
		appIDs = strings.Split(appIDQueryParam, ",")
	}

	exclusiveAppId, err := middleware.ExtractExclusiveAppId(ctx)
	if err != nil {
		return err
	}

	searchBySourceTargetChain := len(sourceChain) > 0 || len(targetChain) > 0
	searchByAppId := len(appIDs) != 0

	if (searchByAddress || searchByTxHash) && (searchBySourceTargetChain || searchByAppId) {
		return response.NewInvalidParamError(ctx, "address/txHash cannot be combined with sourceChain/targetChain/appId query filter", nil)
	}

	payloadTypeParam := ctx.Query("payloadType")
	var payloadType []int
	if payloadTypeParam != "" {
		payloadTypes := strings.Split(payloadTypeParam, ",")
		for _, pt := range payloadTypes {
			ptype, errPtype := strconv.Atoi(pt)
			if errPtype != nil {
				return response.NewInvalidParamError(ctx, "invalid payloadType", errPtype)
			}
			payloadType = append(payloadType, ptype)
		}
	}

	from, err := middleware.ExtractTime(ctx, time.RFC3339, "from")
	if err != nil {
		return err
	}

	to, err := middleware.ExtractTime(ctx, time.RFC3339, "to")
	if err != nil {
		return err
	}

	filter := operations.OperationFilter{
		TxHash:         txHash,
		Address:        address,
		SourceChainIDs: sourceChain,
		TargetChainIDs: targetChain,
		AppIDs:         appIDs,
		ExclusiveAppId: exclusiveAppId,
		PayloadType:    payloadType,
		Pagination:     *pagination,
		From:           from,
		To:             to,
	}

	// Find operations by q search param.
	ops, err := c.srv.FindAll(ctx.Context(), filter)
	if err != nil {
		return err
	}

	// build response
	resp := toListOperationResponse(ops, c.logger)
	return ctx.JSON(resp)
}

// FindById godoc
// @Description Find operations by ID (chainID/emitter/sequence).
// @Tags wormholescan
// @ID get-operation-by-id
// @Param chain_id path integer true "id of the blockchain"
// @Param emitter path string true "address of the emitter"
// @Param seq path integer true "sequence of the VAA"
// @Success 200 {object} OperationResponse
// @Failure 400
// @Failure 500
// @Router /api/v1/operations/{chain_id}/{emitter}/{seq} [get]
func (c *Controller) FindById(ctx *fiber.Ctx) error {
	// Extract query params
	chainID, emitter, seq, err := middleware.ExtractVAAParams(ctx, c.logger)
	if err != nil {
		return err
	}

	// Find operations by chainID, emitter and sequence.
	operation, err := c.srv.FindById(ctx.Context(), chainID, emitter, strconv.FormatUint(seq, 10))
	if err != nil {
		return err
	}

	// build response
	response, err := toOperationResponse(operation, c.logger)
	if err != nil {
		return err
	}
	return ctx.JSON(response)
}
