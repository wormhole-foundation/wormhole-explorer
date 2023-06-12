package main

import (
	"context"
	"log"

	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/common/settings"
	"github.com/wormhole-foundation/wormhole-explorer/core-contract-watcher/config"
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
	_, rootCtxCancel := context.WithCancel(context.Background())

	rootLogger.Info("starting service...")

	// Graceful shutdown
	rootLogger.Info("cancelling root context...")
	rootCtxCancel()
	rootLogger.Info("terminated")
}
