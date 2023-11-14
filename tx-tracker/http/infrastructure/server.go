package infrastructure

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	health "github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/http/vaa"
	"go.uber.org/zap"
)

type Server struct {
	app    *fiber.App
	port   string
	logger *zap.Logger
}

func NewServer(logger *zap.Logger, port string, pprofEnabled bool, vaaController *vaa.Controller, checks ...health.Check) *Server {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	prometheus := fiberprometheus.New("wormscan-tx-tracker")
	prometheus.RegisterAt(app, "/metrics")

	// config use of middlware.
	if pprofEnabled {
		app.Use(pprof.New())
	}
	app.Use(prometheus.Middleware)

	ctrl := NewController(checks, logger)
	api := app.Group("/api")
	api.Get("/health", ctrl.HealthCheck)
	api.Get("/ready", ctrl.ReadyCheck)

	api.Post("/vaa/process", vaaController.Process)

	return &Server{
		app:    app,
		port:   port,
		logger: logger,
	}
}

// Start listen serves HTTP requests from addr.
func (s *Server) Start() {
	addr := ":" + s.port
	s.logger.Info("Listening on " + addr)
	go func() {
		s.app.Listen(addr)
	}()
}

// Stop gracefull server.
func (s *Server) Stop() {
	_ = s.app.Shutdown()
}
