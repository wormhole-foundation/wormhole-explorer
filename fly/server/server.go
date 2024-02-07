package server

import (
	"fmt"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	healthcheck "github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/health"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"
	"go.uber.org/zap"
)

type Server struct {
	app    *fiber.App
	port   string
	logger *zap.Logger
}

func NewServer(port uint, guardianCheck *health.GuardianCheck, logger *zap.Logger, repository *storage.Repository, pprofEnabled bool, alertClient alert.AlertClient, checks ...healthcheck.Check) *Server {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	ctrl := healthcheck.NewController(checks, logger)

	// Configure middleware
	prometheus := fiberprometheus.New("wormscan-fly")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	// config use of middlware.
	if pprofEnabled {
		app.Use(pprof.New())
	}
	api := app.Group("/api")
	api.Get("/health", ctrl.HealthCheck)
	api.Get("/ready", ctrl.ReadyCheck)
	return &Server{
		app:    app,
		port:   fmt.Sprintf("%d", port),
		logger: logger,
	}
}

// Start listen serves HTTP requests from addr.
func (s *Server) Start() {
	go func() {
		s.app.Listen(":" + s.port)
	}()
}

// Stop gracefull server.
func (s *Server) Stop() {
	_ = s.app.Shutdown()
}
