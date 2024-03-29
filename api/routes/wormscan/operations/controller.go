package operations

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/operations"
	"github.com/wormhole-foundation/wormhole-explorer/api/middleware"
	"github.com/wormhole-foundation/wormhole-explorer/api/response"
	"go.uber.org/zap"
)

// Controller is the controller for the operation resource.
type Controller struct {
	srv    *operations.Service
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(operationService *operations.Service, logger *zap.Logger) *Controller {
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

	// Check if address and txHash query param are used together
	if address != "" && txHash != nil {
		if txHash.String() != "" {
			return response.NewInvalidParamError(ctx, "address and txHash query param cannot be used together", nil)
		}
	}

	filter := operations.OperationFilter{
		TxHash:     txHash,
		Address:    address,
		Pagination: *pagination,
	}

	// Find operations by q search param.
	operations, err := c.srv.FindAll(ctx.Context(), filter)
	if err != nil {
		return err
	}

	// build response
	response := toListOperationResponse(operations, c.logger)
	return ctx.JSON(response)
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
