package infrastructure

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/build"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	checks []health.Check
	logger *zap.Logger
}

// NewController creates a Controller instance.
func NewController(checks []health.Check, logger *zap.Logger) *Controller {
	return &Controller{checks: checks, logger: logger}
}

// HealthCheck is the HTTP route handler for the endpoint `GET /api/v1/health`.
// HealthCheck godoc
// @Description Health check
// @Tags wormholescan
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

// ReadyCheck is the HTTP handler for the endpoint `GET /api/v1/ready`.
// ReadyCheck godoc
// @Description Ready check
// @Tags wormholescan
// @ID ready-check
// @Success 200 {object} object{ready=string}
// @Failure 400
// @Failure 500
// @Router /api/v1/ready [get]
func (c *Controller) ReadyCheck(ctx *fiber.Ctx) error {
	rctx := ctx.Context()
	requestID := fmt.Sprintf("%v", rctx.Value("requestid"))
	for _, check := range c.checks {
		if err := check(rctx); err != nil {
			c.logger.Error("Ready check failed", zap.Error(err), zap.String("requestID", requestID))
			return ctx.Status(fiber.StatusInternalServerError).JSON(struct {
				Ready string `json:"ready"`
				Error string `json:"error"`
			}{Ready: "NO", Error: err.Error()})
		}
	}
	return ctx.Status(fiber.StatusOK).JSON(struct {
		Ready string `json:"ready"`
	}{Ready: "OK"})
}

// VersionResponse is the JSON model for the 200 OK response in `GET /api/v1/version`.
type VersionResponse struct {
	BuildDate string `json:"build_date"`
	Build     string `json:"build"`
	Branch    string `json:"branch"`
	Machine   string `json:"machine"`
	User      string `json:"user"`
}

// Version is the HTTP route handler for the endpoint `GET /api/v1/version`.
// Version godoc
// @Description Get version/release information.
// @Tags wormholescan
// @ID get-version
// @Success 200 {object} VersionResponse
// @Failure 400
// @Failure 500
// @Router /api/v1/version [get]
func (c *Controller) Version(ctx *fiber.Ctx) error {
	return ctx.JSON(VersionResponse{
		BuildDate: build.Time,
		Branch:    build.Branch,
		Build:     build.Build,
		Machine:   build.Machine,
		User:      build.User,
	})
}
