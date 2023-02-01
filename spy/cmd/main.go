package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/certusone/wormhole/node/pkg/supervisor"
	ipfslog "github.com/ipfs/go-log/v2"
	"github.com/wormhole-foundation/wormhole-explorer/spy/config"
	"github.com/wormhole-foundation/wormhole-explorer/spy/grpc"
	"github.com/wormhole-foundation/wormhole-explorer/spy/http/infraestructure"
	"github.com/wormhole-foundation/wormhole-explorer/spy/storage"
	"go.uber.org/zap"
)

type exitCode int

func handleExit() {
	if r := recover(); r != nil {
		if e, ok := r.(exitCode); ok {
			os.Exit(int(e))
		}
		panic(r) // not an Exit, bubble up
	}
}

func main() {

	defer handleExit()

	rootCtx, rootCtxCancel := context.WithCancel(context.Background())

	config, err := config.New(rootCtx)
	if err != nil {
		log.Fatal("Error creating config", err)
	}

	level, err := ipfslog.LevelFromString(config.LogLevel)
	if err != nil {
		log.Fatal("Invalid log level", err)
	}

	logger := ipfslog.Logger("wormhole-explorer-spy").Desugar()
	ipfslog.SetAllLoggers(level)

	logger.Info("Starting wormhole-explorer-spy ...")

	svs := grpc.NewSignedVaaSubscribers(logger)
	avs := grpc.NewAllVaaSubscribers(logger)
	go svs.Start(rootCtx)
	go avs.Start(rootCtx)

	handler := grpc.NewHandler(svs, avs, logger)

	grpcServer, err := grpc.NewServer(handler, logger, config.GrpcAddress)
	if err != nil {
		logger.Fatal("failed to start RPC server", zap.Error(err))
	}

	supervisor.New(rootCtx, logger, func(ctx context.Context) error {
		if err := supervisor.Run(ctx, "spyrpc", grpcServer.Runnable); err != nil {
			return err
		}
		<-ctx.Done()
		return nil
	},
		// It's safer to crash and restart the process in case we encounter a panic,
		// rather than attempting to reschedule the runnable.
		supervisor.WithPropagatePanic)

	publisher := grpc.NewPublisher(svs, avs, logger)

	db, err := storage.New(rootCtx, logger, config.MongoURI, config.MongoDatabase)
	if err != nil {
		logger.Fatal("failed to connect MongoDB", zap.Error(err))
	}

	watcher := storage.NewWatcher(db.Database, config.MongoDatabase, publisher.Publish, logger)
	err = watcher.Start(rootCtx)
	if err != nil {
		logger.Fatal("failed to watch MongoDB", zap.Error(err))
	}

	server := infraestructure.NewServer(logger, config.Port, db.Database)
	server.Start()

	logger.Info("Started wormhole-explorer-spy")

	// Waiting for signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-rootCtx.Done():
		logger.Warn("Terminating with root context cancelled.")
	case signal := <-sigterm:
		logger.Info("Terminating with signal.", zap.String("signal", signal.String()))
	}

	logger.Info("root context cancelled, exiting...")
	rootCtxCancel()

	logger.Info("Closing GRPC server ...")
	grpcServer.Stop()
	logger.Info("Closing database connections ...")
	db.Close()
	logger.Info("Closing Http server ...")
	server.Stop()
	logger.Info("Finished wormhole-explorer-spy")
}
