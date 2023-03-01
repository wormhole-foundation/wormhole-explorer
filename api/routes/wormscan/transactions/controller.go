package transactions

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/transactions"
	"go.uber.org/zap"
)

// Controller is the controller for the transactions resource.
type Controller struct {
	srv    *transactions.Service
	logger *zap.Logger
}

// NewController create a new controler.
func NewController(transactionsService *transactions.Service, logger *zap.Logger) *Controller {
	return &Controller{
		srv:    transactionsService,
		logger: logger.With(zap.String("module", "TransactionsController")),
	}
}

// GetLastTrx godoc
func (c *Controller) GetLastTrx(ctx *fiber.Ctx) error {
	return nil
}
