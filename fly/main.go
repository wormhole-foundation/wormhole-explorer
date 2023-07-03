package main

import (
	"context"
	"flag"
	"strconv"
	"strings"

	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/go-redis/redis/v8"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
	"github.com/wormhole-foundation/wormhole-explorer/fly/deduplicator"
	"github.com/wormhole-foundation/wormhole-explorer/fly/guardiansets"
	flyAlert "github.com/wormhole-foundation/wormhole-explorer/fly/internal/alert"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/health"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/sqs"
	"github.com/wormhole-foundation/wormhole-explorer/fly/migration"
	"github.com/wormhole-foundation/wormhole-explorer/fly/notifier"
	"github.com/wormhole-foundation/wormhole-explorer/fly/processor"
	"github.com/wormhole-foundation/wormhole-explorer/fly/queue"
	"github.com/wormhole-foundation/wormhole-explorer/fly/server"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"
	"google.golang.org/protobuf/proto"

	"github.com/certusone/wormhole/node/pkg/common"
	"github.com/certusone/wormhole/node/pkg/p2p"
	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/certusone/wormhole/node/pkg/supervisor"
	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	eth_common "github.com/ethereum/go-ethereum/common"
	crypto2 "github.com/ethereum/go-ethereum/crypto"
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
	nodeKeyPath string
	logLevel    string
)

func getenv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("[%s] env is required", key)
	}
	return v, nil
}

// TODO refactor to another file/package
func newAwsConfig(ctx context.Context) (aws.Config, error) {
	region, err := getenv("AWS_REGION")
	if err != nil {
		return *aws.NewConfig(), err
	}
	awsSecretId, _ := getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey, _ := getenv("AWS_SECRET_ACCESS_KEY")
	if awsSecretId != "" && awsSecretKey != "" {
		credentials := credentials.NewStaticCredentialsProvider(awsSecretId, awsSecretKey, "")
		customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			awsEndpoint, _ := getenv("AWS_ENDPOINT")
			if awsEndpoint != "" {
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           awsEndpoint,
					SigningRegion: region,
				}, nil
			}

			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})

		awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(region),
			awsconfig.WithEndpointResolver(customResolver),
			awsconfig.WithCredentialsProvider(credentials),
		)
		return awsCfg, err
	}

	return awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
}

// TODO refactor to another file/package
func newSQSProducer(ctx context.Context) (*sqs.Producer, error) {
	sqsURL, err := getenv("SQS_URL")
	if err != nil {
		return nil, err
	}

	awsConfig, err := newAwsConfig(ctx)
	if err != nil {
		return nil, err
	}

	return sqs.NewProducer(awsConfig, sqsURL)
}

// TODO refactor to another file/package
func newSQSConsumer(ctx context.Context) (*sqs.Consumer, error) {
	sqsURL, err := getenv("SQS_URL")
	if err != nil {
		return nil, err
	}

	awsConfig, err := newAwsConfig(ctx)
	if err != nil {
		return nil, err
	}

	return sqs.NewConsumer(awsConfig, sqsURL,
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
func newVAAConsumePublish(ctx context.Context, isLocal bool, logger *zap.Logger) (*sqs.Consumer, processor.VAAQueueConsumeFunc, processor.VAAPushFunc) {
	if isLocal {
		vaaQueue := queue.NewVAAInMemory()
		return nil, vaaQueue.Consume, vaaQueue.Publish
	}
	sqsProducer, err := newSQSProducer(ctx)
	if err != nil {
		logger.Fatal("could not create sqs producer", zap.Error(err))
	}

	sqsConsumer, err := newSQSConsumer(ctx)
	if err != nil {
		logger.Fatal("could not create sqs consumer", zap.Error(err))
	}

	vaaQueue := queue.NewVAASQS(sqsProducer, sqsConsumer, logger)
	return sqsConsumer, vaaQueue.Consume, vaaQueue.Publish
}

func newVAANotifierFunc(isLocal bool, logger *zap.Logger) processor.VAANotifyFunc {
	if isLocal {
		return func(context.Context, *vaa.VAA, []byte) error {
			return nil
		}
	}

	redisUri, err := getenv("REDIS_URI")
	if err != nil {
		logger.Fatal("could not create vaa notifier ", zap.Error(err))
	}

	redisPrefix, err := getenv("REDIS_PREFIX")
	if err != nil {
		logger.Fatal("could not create vaa notifier ", zap.Error(err))
	}

	logger.Info("using redis notifier", zap.String("prefix", redisPrefix))
	client := redis.NewClient(&redis.Options{Addr: redisUri})

	return notifier.NewLastSequenceNotifier(client, redisPrefix).Notify
}

func newAlertClient() (alert.AlertClient, error) {
	alertConfig, err := config.GetAlertConfig()
	if err != nil {
		return nil, err
	}
	if !alertConfig.Enabled {
		return alert.NewDummyClient(), nil
	}
	return alert.NewAlertService(alertConfig, flyAlert.LoadAlerts)
}

func newMetrics(enviroment string) metrics.Metrics {
	metricsEnabled := config.GetMetricsEnabled()
	if !metricsEnabled {
		return metrics.NewDummyMetrics()
	}
	return metrics.NewPrometheusMetrics(enviroment)
}

func main() {
	//TODO: use a configuration structure to obtain the configuration
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}

	// Node's main lifecycle context.
	rootCtx, rootCtxCancel = context.WithCancel(context.Background())
	defer rootCtxCancel()

	// get p2p values to connect p2p network
	p2pNetworkConfig, err := config.GetP2pNetwork()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	nodeKeyPath = "/tmp/node.key"
	logLevel = "warn"
	common.SetRestrictiveUmask()

	logger := logger.New("wormhole-fly", logger.WithLevel(logLevel))

	isLocal := flag.Bool("local", false, "a bool")
	flag.Parse()

	// Verify flags
	if nodeKeyPath == "" {
		logger.Fatal("Please specify --nodeKey")
	}
	if p2pNetworkConfig.P2pBootstrap == "" {
		logger.Fatal("Please specify --bootstrap")
	}

	// get Alert client
	alertClient, err := newAlertClient()
	if err != nil {
		logger.Fatal("could not create alert client", zap.Error(err))
	}

	// new metrics client
	metrics := newMetrics(config.GetEnviroment())

	// Setup DB
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		logger.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	databaseName := os.Getenv("MONGODB_DATABASE")
	if databaseName == "" {
		logger.Fatal("You must set your 'MONGODB_DATABASE' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	db, err := storage.GetDB(rootCtx, logger, uri, databaseName)
	if err != nil {
		logger.Fatal("could not connect to DB", zap.Error(err))
	}

	// Run the database migration.
	err = migration.Run(db)
	if err != nil {
		logger.Fatal("error running migration", zap.Error(err))
	}

	repository := storage.NewRepository(alertClient, metrics, db, logger)

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
	guardianSetHistory := guardiansets.GetByEnv(p2pNetworkConfig.Enviroment)
	gsLastet := guardianSetHistory.GetLatest()
	gst.Set(&gsLastet)

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
				metrics.IncObservationTotal()
				ok := verifyObservation(logger, o, gst.Get())
				if !ok {
					logger.Error("Could not verify observation", zap.String("id", o.MessageId))
					continue
				}

				// get chainID from observationID.
				chainID, err := getObservationChainID(logger, o)
				if err != nil {
					logger.Error("Error getting chainID", zap.Error(err))
					continue
				}
				metrics.IncObservationFromGossipNetwork(chainID)

				// apply filter observations by env.
				if filterObservationByEnv(o, p2pNetworkConfig.Enviroment) {
					continue
				}

				metrics.IncObservationUnfiltered(chainID)

				err = repository.UpsertObservation(o)
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
	isLocalFlag := isLocal != nil && *isLocal
	// Creates a deduplicator to discard VAA messages that were processed previously
	deduplicator := deduplicator.New(cache, logger)
	// Creates two callbacks
	sqsConsumer, vaaQueueConsume, nonPythVaaPublish := newVAAConsumePublish(rootCtx, isLocalFlag, logger)
	// Create a vaa notifier
	notifierFunc := newVAANotifierFunc(isLocalFlag, logger)
	// Creates a instance to consume VAA messages from Gossip network and handle the messages
	// When recive a message, the message filter by deduplicator
	// if VAA is from pyhnet should be saved directly to repository
	// if VAA is from non pyhnet should be publish with nonPythVaaPublish
	vaaGossipConsumer := processor.NewVAAGossipConsumer(&guardianSetHistory, deduplicator, nonPythVaaPublish, repository.UpsertVaa, metrics, logger)
	// Creates a instance to consume VAA messages (non pyth) from a queue and store in a storage
	vaaQueueConsumer := processor.NewVAAQueueConsumer(vaaQueueConsume, repository, notifierFunc, metrics, logger)
	// Creates a wrapper that splits the incoming VAAs into 2 channels (pyth to non pyth) in order
	// to be able to process them in a differentiated way
	vaaGossipConsumerSplitter := processor.NewVAAGossipSplitterConsumer(vaaGossipConsumer.Push, logger)
	vaaQueueConsumer.Start(rootCtx)
	vaaGossipConsumerSplitter.Start(rootCtx)

	// start fly http server.
	pprofEnabled := config.GetPprofEnabled()
	maxHealthTimeSeconds := config.GetMaxHealthTimeSeconds()
	guardianCheck := health.NewGuardianCheck(maxHealthTimeSeconds)
	server := server.NewServer(guardianCheck, logger, repository, sqsConsumer, *isLocal, pprofEnabled)
	server.Start()

	go func() {
		for {
			select {
			case <-rootCtx.Done():
				return
			case sVaa := <-signedInC:
				metrics.IncVaaTotal()
				v, err := vaa.Unmarshal(sVaa.Vaa)
				if err != nil {
					logger.Error("Error unmarshalling vaa", zap.Error(err))
					continue
				}

				metrics.IncVaaFromGossipNetwork(v.EmitterChain)
				// apply filter observations by env.
				if filterVaasByEnv(v, p2pNetworkConfig.Enviroment) {
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
	go func(guardianCheck *health.GuardianCheck) {
		for {
			select {
			case <-rootCtx.Done():
				return
			case hb := <-heartbeatC:
				metrics.IncHeartbeatFromGossipNetwork(hb.NodeName)
				err := repository.UpsertHeartbeat(hb)
				if err != nil {
					logger.Error("Error inserting heartbeat", zap.Error(err))
				} else {
					metrics.IncHeartbeatInserted(hb.NodeName)
				}
				guardianCheck.Ping(rootCtx)
			}
		}
	}(guardianCheck)

	// Log govConfigs
	go func() {
		for {
			select {
			case <-rootCtx.Done():
				return
			case govConfig := <-govConfigC:
				nodeName, err := getGovernorConfigNodeName(govConfig)
				if err != nil {
					logger.Error("Error getting gov config node name", zap.Error(err))
					continue
				}
				metrics.IncGovernorConfigFromGossipNetwork(nodeName)

				err = repository.UpsertGovernorConfig(govConfig)
				if err != nil {
					logger.Error("Error inserting gov config", zap.Error(err))
				} else {
					metrics.IncGovernorConfigInserted(nodeName)
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
				nodeName, err := getGovernorStatusNodeName(govStatus)
				if err != nil {
					logger.Error("Error getting gov status node name", zap.Error(err))
					continue
				}
				metrics.IncGovernorStatusFromGossipNetwork(nodeName)
				err = repository.UpsertGovernorStatus(govStatus)
				if err != nil {
					logger.Error("Error inserting gov status", zap.Error(err))
				} else {
					metrics.IncGovernorStatusInserted(nodeName)
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
		if err := supervisor.Run(ctx, "p2p",
			p2p.Run(obsvC, obsvReqC, nil, sendC, signedInC, priv, nil, gst, p2pNetworkConfig.P2pNetworkID, p2pNetworkConfig.P2pBootstrap, "", false, rootCtxCancel, nil, nil, govConfigC, govStatusC, nil)); err != nil {
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

// getGovernorConfigNodeName get node name from governor config.
func getGovernorConfigNodeName(govConfig *gossipv1.SignedChainGovernorConfig) (string, error) {
	var gCfg gossipv1.ChainGovernorConfig
	err := proto.Unmarshal(govConfig.Config, &gCfg)
	if err != nil {
		return "", err
	}
	return gCfg.NodeName, nil
}

// getGovernorStatusNodeName get node name from governor status.
func getGovernorStatusNodeName(govStatus *gossipv1.SignedChainGovernorStatus) (string, error) {
	var gStatus gossipv1.ChainGovernorStatus
	err := proto.Unmarshal(govStatus.Status, &gStatus)
	if err != nil {
		return "", err
	}
	return gStatus.NodeName, nil
}

// getObservationChainID get chainID from observationID.
func getObservationChainID(logger *zap.Logger, obs *gossipv1.SignedObservation) (vaa.ChainID, error) {
	vaaID := strings.Split(obs.MessageId, "/")
	chainIDStr := vaaID[0]
	chainID, err := strconv.ParseUint(chainIDStr, 10, 16)
	if err != nil {
		logger.Error("Error parsing chainId", zap.Error(err))
		return 0, err
	}
	return vaa.ChainID(chainID), nil
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

// filterObservation filter observation by enviroment.
func filterObservationByEnv(o *gossipv1.SignedObservation, enviroment string) bool {
	if enviroment == domain.P2pTestNet {
		// filter pyth message in test enviroment (for solana and pyth chain).
		if strings.Contains((o.GetMessageId()), "1/f346195ac02f37d60d4db8ffa6ef74cb1be3550047543a4a9ee9acf4d78697b0") ||
			strings.HasPrefix("26/", o.GetMessageId()) {
			return true
		}
	}
	return false
}

// filterVaasByEnv filter vaa by enviroment.
func filterVaasByEnv(v *vaa.VAA, enviroment string) bool {
	if enviroment == domain.P2pTestNet {
		vaaFromSolana := v.EmitterChain == vaa.ChainIDSolana
		addressToFilter := strings.ToLower(v.EmitterAddress.String()) == "f346195ac02f37d60d4db8ffa6ef74cb1be3550047543a4a9ee9acf4d78697b0"
		isPyth := v.EmitterChain == vaa.ChainIDPythNet
		if (vaaFromSolana && addressToFilter) || isPyth {
			return true
		}
	}
	return false
}
