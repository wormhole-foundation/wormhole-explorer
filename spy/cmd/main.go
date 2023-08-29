package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/certusone/wormhole/node/pkg/supervisor"
	"github.com/go-redis/redis/v8"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/spy/config"
	"github.com/wormhole-foundation/wormhole-explorer/spy/grpc"
	"github.com/wormhole-foundation/wormhole-explorer/spy/http/infraestructure"
	"github.com/wormhole-foundation/wormhole-explorer/spy/source"
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

func newHealthChecks(
	ctx context.Context,
	client *redis.Client,
) ([]health.Check, error) {

	healthChecks := []health.Check{
		health.Redis(client),
	}
	return healthChecks, nil
}

func main() {

	defer handleExit()

	rootCtx, rootCtxCancel := context.WithCancel(context.Background())

	config, err := config.New(rootCtx)
	if err != nil {
		log.Fatal("Error creating config", err)
	}

	logger := logger.New("wormhole-explorer-spy", logger.WithLevel(config.LogLevel))

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

	client := redis.NewClient(&redis.Options{Addr: config.RedisURI})

	watcher, err := source.NewRedisSubscriber(rootCtx, client, config.RedisPrefix, config.RedisChannel, publisher.Publish, logger)
	if err != nil {
		logger.Fatal("failed to create redis subscriber", zap.Error(err))
	}
	err = watcher.Start(rootCtx)
	if err != nil {
		logger.Fatal("failed to watch MongoDB", zap.Error(err))
	}

	// get health check functions.
	logger.Info("creating health check functions...")
	healthChecks, err := newHealthChecks(rootCtx, client)
	if err != nil {
		logger.Fatal("failed to create health checks", zap.Error(err))
	}

	server := infraestructure.NewServer(logger, config.Port, config.PprofEnabled, healthChecks...)
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

	logger.Info("Closing Redis connection...")
	if err := watcher.Close(rootCtx); err != nil {
		logger.Error("Error closing watcher", zap.Error(err))
	}
	if err := client.Close(); err != nil {
		logger.Error("Error closing redis client", zap.Error(err))
	}

	logger.Info("Closing Http server ...")
	server.Stop()
	logger.Info("Finished wormhole-explorer-spy")
}
