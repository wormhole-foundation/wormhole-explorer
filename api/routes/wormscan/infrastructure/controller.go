package infrastructure

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/api/handlers/infrastructure"
	"github.com/wormhole-foundation/wormhole-explorer/api/internal/build"
)

// Controller definition.
type Controller struct {
	srv *infrastructure.Service
}

// NewController creates a Controller instance.
func NewController(serv *infrastructure.Service) *Controller {
	return &Controller{srv: serv}
}

// HealthCheck is the HTTP route handler for the endpoint `GET /api/v1/health`.
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

// ReadyCheck is the HTTP handler for the endpoint `GET /api/v1/ready`.
// ReadyCheck godoc
// @Description Ready check
// @Tags Wormscan
// @ID ready-check
// @Success 200 {object} object{ready=string}
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

// VersionResponse is the JSON model for the 200 OK response in `GET /api/v1/version`.
type VersionResponse struct {
	BuildDate string `json:"buildDate"`
	Build     string `json:"build"`
	Branch    string `json:"branch"`
	Machine   string `json:"machine"`
	User      string `json:"user"`
}

// Version is the HTTP route handler for the endpoint `GET /api/v1/version`.
// Version godoc
// @Description Get version/release information.
// @Tags Wormscan
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
