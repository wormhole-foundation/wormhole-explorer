package infraestructure

import "github.com/gofiber/fiber/v2"

// Controller definition.
type Controller struct {
	srv *Service
}

// NewController creates a Controller instance.
func NewController(serv *Service) *Controller {
	return &Controller{srv: serv}
}

// HealthCheck handler for the endpoint /health.
func (c *Controller) HealthCheck(ctx *fiber.Ctx) error {
	return ctx.JSON(struct {
		Status string `json:"status"`
	}{Status: "OK"})
}

// ReadyCheck handler for the endpoint /ready
func (c *Controller) ReadyCheck(ctx *fiber.Ctx) error {
	ready, _ := c.srv.CheckMongoServerStatus(ctx.Context())
	if ready {
		return ctx.Status(fiber.StatusOK).JSON(struct {
			Ready string `json:"ready"`
		}{Ready: "OK"})
	}
	return ctx.Status(fiber.StatusInternalServerError).JSON(struct {
		Ready string `json:"ready"`
	}{Ready: "NO"})
}
