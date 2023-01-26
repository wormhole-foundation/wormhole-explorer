package infrastructure

import "github.com/gofiber/fiber/v2"

// Controller definition.
type Controller struct {
	srv *Service
}

// NewController creates a Controller instance.
func NewController(serv *Service) *Controller {
	return &Controller{srv: serv}
}

// HealthCheck godoc
// @Description Health check
// @Tags Wormscan
// @ID health-check
// @Success 200 {object} object{status=string}
// @Failure 400
// @Failure 500
// @Router /api/v1/health [get]
func (c *Controller) HealthCheck(ctx *fiber.Ctx) error {
	return ctx.JSON(struct {
		Status string `json:"status"`
	}{Status: "OK"})
}

// ReadyCheck handler for the endpoint /ready
// ReadyCheck godoc
// @Description Ready check
// @Tags Wormscan
// @ID ready-check
// @Success 200 {object} object{status=string}
// @Failure 400
// @Failure 500
// @Router /api/v1/ready [get]
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
