package infrastructure

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	srv    *Service
	logger *zap.Logger
}

// NewController creates a Controller instance.
func NewController(serv *Service, logger *zap.Logger) *Controller {
	return &Controller{srv: serv, logger: logger}
}

// HealthCheck handler for the endpoint /health.
func (c *Controller) HealthCheck(ctx *fiber.Ctx) error {
	return ctx.JSON(struct {
		Status string `json:"status"`
	}{Status: "OK"})
}

// ReadyCheck handler for the endpoint /ready.
func (c *Controller) ReadyCheck(ctx *fiber.Ctx) error {
	ready, err := c.srv.CheckIsReady(ctx.Context())
	if ready {
		return ctx.Status(fiber.StatusOK).JSON(struct {
			Ready string `json:"ready"`
		}{Ready: "OK"})
	}
	c.logger.Error("Ready check failed", zap.Error(err))
	return ctx.Status(fiber.StatusInternalServerError).JSON(struct {
		Ready string `json:"ready"`
		Error string `json:"error"`
	}{Ready: "NO", Error: err.Error()})
}
