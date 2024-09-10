package main

import (
	"context"
	"flag"
	"log"
	"time"

	"fmt"
	"os"

	healthcheck "github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/fly/builder"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
	"github.com/wormhole-foundation/wormhole-explorer/fly/gossip"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/health"
	"github.com/wormhole-foundation/wormhole-explorer/fly/migration"
	"github.com/wormhole-foundation/wormhole-explorer/fly/processor"
	"github.com/wormhole-foundation/wormhole-explorer/fly/producer"
	"github.com/wormhole-foundation/wormhole-explorer/fly/server"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"

	"github.com/certusone/wormhole/node/pkg/common"
	"github.com/certusone/wormhole/node/pkg/p2p"
	"github.com/certusone/wormhole/node/pkg/supervisor"
	"github.com/libp2p/go-libp2p/core/crypto"
	"go.uber.org/zap"
)

func main() {

	// Node's main lifecycle context.
	rootCtx, rootCtxCancel := context.WithCancel(context.Background())
	defer rootCtxCancel()

	isLocal := flag.Bool("local", false, "a bool")
	flag.Parse()

	// Load configuration
	cfg, err := config.New(rootCtx, isLocal)
	if err != nil {
		log.Fatal("Error creating config", err)
	}

	// Get p2p values to connect p2p network
	p2pNetworkConfig, err := cfg.GetP2pNetwork()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	nodeKeyPath := "/tmp/node.key"
	common.SetRestrictiveUmask()

	logger := logger.New("wormhole-fly", logger.WithLevel(cfg.LogLevel))

	// Verify flags
	if nodeKeyPath == "" {
		logger.Fatal("Please specify --nodeKey")
	}
	if p2pNetworkConfig.P2pBootstrap == "" {
		logger.Fatal("Please specify --bootstrap")
	}

	// New alert client
	alertClient, err := builder.NewAlertClient(cfg)
	if err != nil {
		logger.Fatal("could not create alert client", zap.Error(err))
	}

	// New metrics client
	metrics := builder.NewMetrics(cfg)

	// New database session
	db, err := builder.NewDatabase(rootCtx, cfg, logger)
	if err != nil {
		logger.Fatal("could not connect to DB", zap.Error(err))
	}

	// Run the database migration.
	err = migration.Run(db.Database)
	if err != nil {
		logger.Fatal("error running migration", zap.Error(err))
	}

	// Creates a callback to publish VAA messages to a redis pubsub
	vaaRedisProducerFunc, err := builder.NewVAARedisProducerFunc(cfg, logger)
	if err != nil {
		logger.Fatal("could not create vaa redis producer", zap.Error(err))
	}

	// Creates a composite callback to publish VAA messages to a redis pubsub
	producerFunc := producer.NewComposite(vaaRedisProducerFunc)

	txHashStore, err := builder.NewTxHashStore(rootCtx, cfg, metrics, db.Database, logger)
	if err != nil {
		logger.Fatal("could not create tx hash store", zap.Error(err))
	}
	eventDispatcher, healthEvents := builder.NewEventDispatcher(rootCtx, cfg, logger)

	repository := storage.NewRepository(alertClient, metrics, db.Database, producerFunc, txHashStore, eventDispatcher, logger)

	vaaNonPythDedup, err := builder.NewDeduplicator("vaas-dedup", cfg.VaasDedup, logger)
	if err != nil {
		logger.Fatal("could not create vaa deduplicator", zap.Error(err))
	}

	vaaPythDedup, err := builder.NewDeduplicator("vaas-pyth-dedup", cfg.VaasPythDedup, logger)
	if err != nil {
		logger.Fatal("could not create vaa deduplicator", zap.Error(err))
	}

	channels := builder.NewGossipChannels(cfg)

	guardianSetSyncronizer, err := builder.NewGuardianSetSynchronizer(rootCtx, db, channels.HeartbeatChannel,
		logger, cfg, alertClient)
	if err != nil {
		logger.Fatal("could not create guardian set synchronizer", zap.Error(err))
	}
	guardianSetSyncronizer.Sync(rootCtx)

	guardianSetHistory := guardianSetSyncronizer.GetGuardianSetHistory()
	gst := guardianSetSyncronizer.GetLatestGuardianSet()

	// Ignore observation requests
	// Note: without this, the whole program hangs on observation requests
	discardMessages(rootCtx, channels.ObsvReqChannel)
	guardianCheck := health.NewGuardianCheck(cfg.MaxHealthTimeSeconds)

	healthObservations, observationQueueConsume, observationPublish := builder.NewObservationConsumePublish(rootCtx, cfg, logger)
	observationGossipConsumer := processor.NewObservationGossipConsumer(observationPublish, gst, p2pNetworkConfig.Enviroment,
		cfg.ObservationsChannelSize, cfg.ObservationsWorkersSize, metrics, txHashStore, repository, logger)
	observationQueueConsumer := processor.NewObservationQueueConsumer(observationQueueConsume, repository, metrics, logger)
	observationGossipConsumer.Start(rootCtx)
	observationQueueConsumer.Start(rootCtx)

	// Log observations
	observationHandler := gossip.NewObservationHandler(channels.ObsvChannel, channels.BatchObsvC, observationGossipConsumer.Push, observationGossipConsumer.PushBatch, guardianCheck, metrics)
	observationHandler.Start(rootCtx)

	// Log signed VAAs
	// Creates two callbacks
	healthVaas, vaaQueueConsume, nonPythVaaPublish := builder.NewVAAConsumePublish(rootCtx, cfg, logger)
	// Create a vaa notifier
	notifierFunc := builder.NewVAANotifierFunc(cfg, logger)
	// Creates a instance to consume VAA messages from Gossip network and handle the messages
	// When recive a message, the message filter by deduplicator
	// if VAA is from pyhnet should be saved directly to repository
	// if VAA is from non pyhnet should be publish with nonPythVaaPublish
	vaaGossipConsumer := processor.NewVAAGossipConsumer(guardianSetHistory, vaaNonPythDedup, vaaPythDedup, nonPythVaaPublish, repository.UpsertVaa, metrics, repository, logger)
	// Creates a instance to consume VAA messages (non pyth) from a queue and store in a storage
	vaaQueueConsumer := processor.NewVAAQueueConsumer(vaaQueueConsume, repository, notifierFunc, metrics, logger)
	// Creates a wrapper that splits the incoming VAAs into 2 channels (pyth to non pyth) in order
	// to be able to process them in a differentiated way
	vaaGossipConsumerSplitter := processor.NewVAAGossipSplitterConsumer(vaaGossipConsumer.Push, cfg.VaasWorkersSize, logger, processor.WithSize(cfg.VaasChannelSize))
	vaaQueueConsumer.Start(rootCtx)
	vaaGossipConsumerSplitter.Start(rootCtx)

	// start fly http server.
	healthChecks := []healthcheck.Check{healthObservations, healthVaas, healthEvents, builder.CheckGuardian(guardianCheck)}
	pprofEnabled := cfg.PprofEnabled
	server := server.NewServer(cfg.ApiPort, guardianCheck, logger, repository, pprofEnabled, alertClient, healthChecks...)
	server.Start()

	// VAA handler
	vaaHandler := gossip.NewVaaHandler(p2pNetworkConfig, metrics, channels.SignedInChannel, vaaGossipConsumerSplitter.Push, guardianCheck, logger)
	vaaHandler.Start(rootCtx)

	// Heartbeats handler
	hearbeatsHandler := gossip.NewHeartbeatsHandler(channels.HeartbeatChannel, repository, guardianCheck, metrics, logger)
	hearbeatsHandler.Start(rootCtx)

	// Governor config handler
	governorConfigHandler := gossip.NewGovernorConfigHandler(channels.GovConfigChannel, repository, guardianCheck, metrics, logger)
	governorConfigHandler.Start(rootCtx)

	// Governor status handler
	governorStatusHandler := gossip.NewGovernorStatusHandler(channels.GovStatusChannel, repository, guardianCheck, metrics, logger)
	governorStatusHandler.Start(rootCtx)

	// Load p2p private key
	var priv crypto.PrivKey
	priv, err = common.GetOrCreateNodeKey(logger, nodeKeyPath)
	if err != nil {
		logger.Fatal("Failed to load node key", zap.Error(err))
	}

	runParams, errRunParams := p2p.NewRunParams(
		p2pNetworkConfig.P2pBootstrap,
		p2pNetworkConfig.P2pNetworkID,
		priv,
		gst,
		rootCtxCancel,
		p2p.WithSignedObservationListener(channels.ObsvChannel),
		p2p.WithSignedObservationBatchListener(channels.BatchObsvC),
		p2p.WithSignedVAAListener(channels.SignedInChannel),
		p2p.WithObservationRequestListener(channels.ObsvReqChannel),
		p2p.WithChainGovernorConfigListener(channels.GovConfigChannel),
		p2p.WithChainGovernorStatusListener(channels.GovStatusChannel),
	)
	if errRunParams != nil {
		logger.Fatal("failed to create run params", zap.Error(errRunParams))
	}

	// Run supervisor.
	supervisor.New(rootCtx, logger, func(ctx context.Context) error {
		components := p2p.DefaultComponents()
		components.Port = cfg.P2pPort
		components.WarnChannelOverflow = true
		if err := supervisor.Run(ctx, "p2p",
			p2p.Run(
				runParams,
			)); err != nil {
			return err
		}

		logger.Info("Started internal services")

		<-ctx.Done()
		return nil
	},
		// It's safer to crash and restart the process in case we encounter a panic,
		// rather than attempting to reschedule the runnable.
		supervisor.WithPropagatePanic)

	<-rootCtx.Done()

	// TODO: wait for things to shut down gracefully
	vaaGossipConsumerSplitter.Close()
	observationGossipConsumer.Close()
	server.Stop()

	logger.Info("Closing MongoDB connection...")
	db.DisconnectWithTimeout(10 * time.Second)
}

func discardMessages[T any](ctx context.Context, obsvReqC chan T) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-obsvReqC:
			}
		}
	}()
}
