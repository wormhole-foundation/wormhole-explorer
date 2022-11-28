package main

import (
	"context"
	"flag"

	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/wormhole-foundation/wormhole-explorer/fly/deduplicator"
	"github.com/wormhole-foundation/wormhole-explorer/fly/guardiansets"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/sqs"
	"github.com/wormhole-foundation/wormhole-explorer/fly/migration"
	"github.com/wormhole-foundation/wormhole-explorer/fly/processor"
	"github.com/wormhole-foundation/wormhole-explorer/fly/queue"
	"github.com/wormhole-foundation/wormhole-explorer/fly/server"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"

	"github.com/certusone/wormhole/node/pkg/common"
	"github.com/certusone/wormhole/node/pkg/p2p"
	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/certusone/wormhole/node/pkg/supervisor"
	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	eth_common "github.com/ethereum/go-ethereum/common"
	crypto2 "github.com/ethereum/go-ethereum/crypto"
	ipfslog "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"

	"github.com/joho/godotenv"
)

var (
	rootCtx       context.Context
	rootCtxCancel context.CancelFunc
)

var (
	p2pNetworkID string
	p2pPort      uint
	p2pBootstrap string
	nodeKeyPath  string
	logLevel     string
)

func getenv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("[%s] env is required", key)
	}
	return v, nil
}

// TODO refactor to another file/package
func newAwsSession() (*session.Session, error) {
	region, err := getenv("AWS_REGION")
	if err != nil {
		return nil, err
	}
	config := aws.NewConfig().WithRegion(region)
	awsSecretId, _ := getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey, _ := getenv("AWS_SECRET_ACCESS_KEY")
	if awsSecretId != "" && awsSecretKey != "" {
		config.WithCredentials(credentials.NewStaticCredentials(awsSecretId, awsSecretKey, ""))
	}
	if awsEndpoint, err := getenv("AWS_ENDPOINT"); err == nil {
		config.WithEndpoint(awsEndpoint)
	}

	return session.NewSession(config)
}

// TODO refactor to another file/package
func newSQSProducer() (*sqs.Producer, error) {
	sqsURL, err := getenv("SQS_URL")
	if err != nil {
		return nil, err
	}

	session, err := newAwsSession()
	if err != nil {
		return nil, err
	}

	return sqs.NewProducer(session, sqsURL)
}

// TODO refactor to another file/package
func newSQSConsumer() (*sqs.Consumer, error) {
	sqsURL, err := getenv("SQS_URL")
	if err != nil {
		return nil, err
	}

	session, err := newAwsSession()
	if err != nil {
		return nil, err
	}

	return sqs.NewConsumer(session, sqsURL,
		sqs.WithMaxMessages(10),
		sqs.WithVisibilityTimeout(120))
}

// TODO refactor to another file/package
func newCache() (cache.CacheInterface[bool], error) {
	c, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 10000,          // Num keys to track frequency of (1000).
		MaxCost:     10 * (1 << 20), // Maximum cost of cache (10 MB).
		BufferItems: 64,             // Number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}
	store := store.NewRistretto(c)
	return cache.New[bool](store), nil
}

// Creates two callbacks depending on whether the execution is local (memory queue) or not (SQS queue)
// callback to obtain queue messages from a queue
// callback to publish vaa non pyth messages to a sink
func newVAAConsumePublish(isLocal bool, logger *zap.Logger) (*sqs.Consumer, processor.VAAQueueConsumeFunc, processor.VAAPushFunc) {
	if isLocal {
		vaaQueue := queue.NewVAAInMemory()
		return nil, vaaQueue.Consume, vaaQueue.Publish
	}
	sqsProducer, err := newSQSProducer()
	if err != nil {
		logger.Fatal("could not create sqs producer", zap.Error(err))
	}

	sqsConsumer, err := newSQSConsumer()
	if err != nil {
		logger.Fatal("could not create sqs consumer", zap.Error(err))
	}

	vaaQueue := queue.NewVAASQS(sqsProducer, sqsConsumer, logger)
	return sqsConsumer, vaaQueue.Consume, vaaQueue.Publish
}

func main() {
	// Node's main lifecycle context.

	rootCtx, rootCtxCancel = context.WithCancel(context.Background())
	defer rootCtxCancel()
	// main
	p2pNetworkID = "/wormhole/mainnet/2"
	p2pBootstrap = "/dns4/wormhole-mainnet-v2-bootstrap.certus.one/udp/8999/quic/p2p/12D3KooWQp644DK27fd3d4Km3jr7gHiuJJ5ZGmy8hH4py7fP4FP7"
	// devnet
	// p2pNetworkID = "/wormhole/dev"
	// p2pBootstrap = "/dns4/guardian-0.guardian/udp/8999/quic/p2p/12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw"
	p2pPort = 8999
	nodeKeyPath = "/tmp/node.key"
	logLevel = "warn"
	common.SetRestrictiveUmask()

	lvl, err := ipfslog.LevelFromString(logLevel)
	if err != nil {
		fmt.Println("Invalid log level")
		os.Exit(1)
	}

	logger := ipfslog.Logger("wormhole-fly").Desugar()

	ipfslog.SetAllLoggers(lvl)

	isLocal := flag.Bool("local", false, "a bool")
	flag.Parse()

	// Verify flags
	if nodeKeyPath == "" {
		logger.Fatal("Please specify --nodeKey")
	}
	if p2pBootstrap == "" {
		logger.Fatal("Please specify --bootstrap")
	}

	//TODO: use a configuration structure to obtain the configuration
	// Setup DB
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found")
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		logger.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	db, err := storage.GetDB(rootCtx, logger, uri, "wormhole")
	if err != nil {
		logger.Fatal("could not connect to DB", zap.Error(err))
	}

	// Run the database migration.
	err = migration.Run(db)
	if err != nil {
		logger.Fatal("error running migration", zap.Error(err))
	}

	repository := storage.NewRepository(db, logger)

	// Outbound gossip message queue
	sendC := make(chan []byte)

	// Inbound observations
	obsvC := make(chan *gossipv1.SignedObservation, 50)

	// Inbound observation requests
	obsvReqC := make(chan *gossipv1.ObservationRequest, 50)

	// Inbound signed VAAs
	signedInC := make(chan *gossipv1.SignedVAAWithQuorum, 50)

	// Heartbeat updates
	heartbeatC := make(chan *gossipv1.Heartbeat, 50)

	// Guardian set state managed by processor
	gst := common.NewGuardianSetState(heartbeatC)

	// Governor cfg
	govConfigC := make(chan *gossipv1.SignedChainGovernorConfig, 50)

	// Governor status
	govStatusC := make(chan *gossipv1.SignedChainGovernorStatus, 50)
	// Bootstrap guardian set, otherwise heartbeats would be skipped
	// TODO: fetch this and probably figure out how to update it live
	gs := guardiansets.GetLatest()
	gst.Set(&gs)

	// Ignore observation requests
	// Note: without this, the whole program hangs on observation requests
	discardMessages(rootCtx, obsvReqC)

	// Log observations
	go func() {
		for {
			select {
			case <-rootCtx.Done():
				return
			case o := <-obsvC:
				ok := verifyObservation(logger, o, gst.Get())
				if !ok {
					logger.Error("Could not verify observation", zap.String("id", o.MessageId))
					continue
				}
				err := repository.UpsertObservation(o)
				if err != nil {
					logger.Error("Error inserting observation", zap.Error(err))
				}
			}
		}
	}()

	// Log signed VAAs
	cache, err := newCache()
	if err != nil {
		logger.Fatal("could not create cache", zap.Error(err))
	}
	// Creates a deduplicator to discard VAA messages that were processed previously
	deduplicator := deduplicator.New(cache, logger)
	// Creates two callbacks
	sqsConsumer, vaaQueueConsume, nonPythVaaPublish := newVAAConsumePublish(isLocal != nil && *isLocal, logger)
	// Creates a instance to consume VAA messages from Gossip network and handle the messages
	// When recive a message, the message filter by deduplicator
	// if VAA is from pyhnet should be saved directly to repository
	// if VAA is from non pyhnet should be publish with nonPythVaaPublish
	vaaGossipConsumer := processor.NewVAAGossipConsumer(gst, deduplicator, nonPythVaaPublish, repository.UpsertVaa, logger)
	// Creates a instance to consume VAA messages (non pyth) from a queue and store in a storage
	vaaQueueConsumer := processor.NewVAAQueueConsumer(vaaQueueConsume, repository, logger)
	// Creates a wrapper that splits the incoming VAAs into 2 channels (pyth to non pyth) in order
	// to be able to process them in a differentiated way
	vaaGossipConsumerSplitter := processor.NewVAAGossipSplitterConsumer(vaaGossipConsumer.Push, logger)
	vaaQueueConsumer.Start(rootCtx)
	vaaGossipConsumerSplitter.Start(rootCtx)

	// start fly http server.
	server := server.NewServer(logger, repository, sqsConsumer, *isLocal)
	server.Start()

	go func() {
		for {
			select {
			case <-rootCtx.Done():
				return
			case sVaa := <-signedInC:
				v, err := vaa.Unmarshal(sVaa.Vaa)
				if err != nil {
					logger.Error("Error unmarshalling vaa", zap.Error(err))
					continue
				}
				// Push an incoming VAA to be processed
				if err := vaaGossipConsumerSplitter.Push(rootCtx, v, sVaa.Vaa); err != nil {
					logger.Error("Error inserting vaa", zap.Error(err))
				}
			}
		}
	}()

	// Log heartbeats
	go func() {
		for {
			select {
			case <-rootCtx.Done():
				return
			case hb := <-heartbeatC:
				err := repository.UpsertHeartbeat(hb)
				if err != nil {
					logger.Error("Error inserting heartbeat", zap.Error(err))
				}
			}
		}
	}()

	// Log govConfigs
	go func() {
		for {
			select {
			case <-rootCtx.Done():
				return
			case govConfig := <-govConfigC:
				err := repository.UpsertGovernorConfig(govConfig)
				if err != nil {
					logger.Error("Error inserting gov config", zap.Error(err))
				}
			}
		}
	}()

	// Log govStatus
	go func() {
		for {
			select {
			case <-rootCtx.Done():
				return
			case govStatus := <-govStatusC:
				err := repository.UpsertGovernorStatus(govStatus)
				if err != nil {
					logger.Error("Error inserting gov status", zap.Error(err))
				}
			}
		}
	}()

	// Load p2p private key
	var priv crypto.PrivKey
	priv, err = common.GetOrCreateNodeKey(logger, nodeKeyPath)
	if err != nil {
		logger.Fatal("Failed to load node key", zap.Error(err))
	}

	// Run supervisor.
	supervisor.New(rootCtx, logger, func(ctx context.Context) error {
		if err := supervisor.Run(ctx, "p2p", p2p.Run(obsvC, obsvReqC, nil, sendC, signedInC, priv, nil, gst, p2pPort, p2pNetworkID, p2pBootstrap, "", false, rootCtxCancel, nil, govConfigC, govStatusC)); err != nil {
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
	server.Stop()
}

func verifyObservation(logger *zap.Logger, obs *gossipv1.SignedObservation, gs *common.GuardianSet) bool {
	pk, err := crypto2.Ecrecover(obs.GetHash(), obs.GetSignature())
	if err != nil {
		return false
	}

	theirAddr := eth_common.BytesToAddress(obs.GetAddr())
	signerAddr := eth_common.BytesToAddress(crypto2.Keccak256(pk[1:])[12:])
	if theirAddr != signerAddr {
		logger.Error("error validating observation, signer addr and addr don't match",
			zap.String("id", obs.MessageId),
			zap.String("obs_addr", theirAddr.Hex()),
			zap.String("signer_addr", signerAddr.Hex()),
		)
		return false
	}

	_, isFromGuardian := gs.KeyIndex(theirAddr)
	if !isFromGuardian {
		logger.Error("error validating observation, signer not in guardian set",
			zap.String("id", obs.MessageId),
			zap.String("obs_addr", theirAddr.Hex()),
		)
	}
	return isFromGuardian
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
