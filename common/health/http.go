package health

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"go.uber.org/zap"
)

type Server struct {
	app    *fiber.App
	port   string
	logger *zap.Logger
}

func NewServer(logger *zap.Logger, port string, pprofEnabled bool, checks ...Check) *Server {

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	// config use of middlware.
	if pprofEnabled {
		app.Use(pprof.New())
	}

	ctrl := newController(checks, logger)
	api := app.Group("/api")
	api.Get("/health", ctrl.healthCheck)
	api.Get("/ready", ctrl.readinessCheck)

	return &Server{
		app:    app,
		port:   port,
		logger: logger,
	}
}

// Start initiates the serving of HTTP requests.
func (s *Server) Start() {

	addr := ":" + s.port
	s.logger.Info("Monitoring server starting", zap.String("bindAddress", addr))

	go func() {
		err := s.app.Listen(addr)
		if err != nil {
			s.logger.Error("Failed to start monitoring server", zap.Error(err), zap.String("bindAddress", addr))
		}
	}()
}

// Stop gracefully shuts down the server.
//
// Blocks until all active connections are closed.
func (s *Server) Stop() {
	_ = s.app.Shutdown()
}

type controller struct {
	checks []Check
	logger *zap.Logger
}

// newController creates a Controller instance.
func newController(checks []Check, logger *zap.Logger) *controller {
	return &controller{checks: checks, logger: logger}
}

// healthCheck is the HTTP handler for the route `GET /health`.
func (c *controller) healthCheck(ctx *fiber.Ctx) error {

	response := ctx.JSON(struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	})

	return response
}

// readinessCheck is the HTTP handler for the route `GET /ready`.
func (c *controller) readinessCheck(ctx *fiber.Ctx) error {

	requestCtx := ctx.Context()
	requestID := fmt.Sprintf("%v", requestCtx.Value("requestid"))

	// For every callback, check whether it is passing
	for _, check := range c.checks {
		if err := check(requestCtx); err != nil {

			c.logger.Error(
				"Readiness check failed",
				zap.Error(err),
				zap.String("requestID", requestID),
			)

			// Return error information to the caller
			response := ctx.
				Status(fiber.StatusInternalServerError).
				JSON(struct {
					Ready string `json:"ready"`
					Error string `json:"error"`
				}{
					Ready: "NO",
					Error: err.Error(),
				})
			return response
		}
	}

	// All checks passed
	response := ctx.Status(fiber.StatusOK).
		JSON(struct {
			Ready string `json:"ready"`
		}{
			Ready: "OK",
		})
	return response
}
