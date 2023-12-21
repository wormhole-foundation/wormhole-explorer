package redeem

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/watcher"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	watcherByBlockchain map[string]watcher.ContractWatcher
	logger              *zap.Logger
}

// NewController creates a Controller instance.
func NewController(watckers []watcher.ContractWatcher, logger *zap.Logger) *Controller {
	watcherByBlockchain := make(map[string]watcher.ContractWatcher)
	for _, w := range watckers {
		watcherByBlockchain[w.GetBlockchain()] = w
	}
	return &Controller{watcherByBlockchain: watcherByBlockchain, logger: logger}
}

func (c *Controller) Backfill(ctx *fiber.Ctx) error {
	payload := struct {
		Blockchain string `json:"blockchain"`
		FromBlock  uint64 `json:"fromBlock"`
		ToBlock    uint64 `json:"toBlock"`
	}{}

	if err := ctx.BodyParser(&payload); err != nil {
		return err
	}

	c.logger.Info("Executing contract-watcher", zap.Any("payload", payload))

	watcher, ok := c.watcherByBlockchain[payload.Blockchain]
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "Blockchain not found")
	}

	watcher.Backfill(ctx.Context(), payload.FromBlock, payload.ToBlock, 100, false)

	return nil
}
