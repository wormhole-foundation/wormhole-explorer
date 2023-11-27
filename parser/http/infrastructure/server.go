package infrastructure

import (
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/parser/http/vaa"
	"go.uber.org/zap"
)

type Server struct {
	app    *fiber.App
	port   string
	logger *zap.Logger
}

func NewServer(logger *zap.Logger, port string, pprofEnabled bool, vaaController *vaa.Controller, checks ...health.Check) *Server {
	ctrl := health.NewController(checks, logger)
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	// config use of middlware.
	prometheus := fiberprometheus.New("wormscan-pipeline")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	if pprofEnabled {
		app.Use(pprof.New())
	}

	api := app.Group("/api")
	api.Get("/health", ctrl.HealthCheck)
	api.Get("/ready", ctrl.ReadyCheck)

	api.Post("/vaa/parse", vaaController.Parse)

	return &Server{
		app:    app,
		port:   port,
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
