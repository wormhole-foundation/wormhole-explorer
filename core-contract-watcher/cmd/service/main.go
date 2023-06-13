package main

import (
	"context"
	"log"

	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/common/mongohelpers"
	"github.com/wormhole-foundation/wormhole-explorer/common/settings"
	"github.com/wormhole-foundation/wormhole-explorer/core-contract-watcher/config"
	"go.uber.org/zap"
)

func main() {

	// Load config
	cfg, err := settings.LoadFromEnv[config.ServiceSettings]()
	if err != nil {
		log.Fatal("Error loading config: ", err)
	}

	// Build rootLogger
	rootLogger := logger.New("wormhole-explorer-core-contract-watcher", logger.WithLevel(cfg.LogLevel))

	// Create top-level context
	rootCtx, rootCtxCancel := context.WithCancel(context.Background())

	// Connect to MongoDB
	rootLogger.Info("connecting to MongoDB...")
	db, err := mongohelpers.Connect(rootCtx, cfg.MongodbURI, cfg.MongodbDatabase)
	if err != nil {
		rootLogger.Fatal("Error connecting to MongoDB", zap.Error(err))
	}

	// Shut down gracefully
	rootLogger.Info("disconnecting from MongoDB...")
	db.Disconnect(rootCtx)
	rootLogger.Info("cancelling root context...")
	rootCtxCancel()
	rootLogger.Info("terminated")
}
