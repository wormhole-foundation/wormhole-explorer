package server

import (
	"os"

	"github.com/gofiber/fiber/v2"
	fiberLog "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/sqs"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"
	"go.uber.org/zap"
)

type Server struct {
	app    *fiber.App
	port   string
	logger *zap.Logger
}

func NewServer(logger *zap.Logger, repository *storage.Repository, consumer *sqs.Consumer, isLocal bool) *Server {
	port := os.Getenv("API_PORT")
	if port == "" {
		logger.Fatal("You must set your 'API_PORT' environmental variable")
	}
	ctrl := NewController(repository, consumer, isLocal)
	app := fiber.New()
	app.Use(fiberLog.New(fiberLog.Config{
		Format: "level=info timestamp=${time} method=${method} path=${path} status${status} request_id=${locals:requestid}\n",
	}))
	api := app.Group("/api")
	api.Get("/health", ctrl.HealthCheck)
	api.Get("/ready", ctrl.ReadyCheck)
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
