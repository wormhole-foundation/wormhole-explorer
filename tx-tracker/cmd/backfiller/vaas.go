package backfiller

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/common/pool"
	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/consumer"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/metrics"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/ratelimit"
	"go.uber.org/zap"
)

type VaasBackfiller struct {
	P2pNetwork        string
	LogLevel          string
	MongoURI          string
	MongoDatabase     string
	RequestsPerMinute int64
	StartTime         string
	EndTime           string
	EmitterChainID    *sdk.ChainID
	EmitterAddress    *string
	Overwrite         bool
	DisableDBUpsert   bool
	PageSize          int64
	NumWorkers        int
	RpcProvidersPath  string
}

type vaasBackfillerParams struct {
	logger                      *zap.Logger
	rpcPool                     map[sdk.ChainID]*pool.Pool
	wormchainRpcPool            map[sdk.ChainID]*pool.Pool
	repository                  *consumer.Repository
	queue                       chan *repository.VaaDoc
	wg                          *sync.WaitGroup
	processedDocumentsSuccess   *atomic.Uint64
	processedDocumentsWithError *atomic.Uint64
	p2pNetwork                  string
	overwrite                   bool
	disableDBUpsert             bool
	limiter                     ratelimit.Limiter
}

func RunByVaas(backfillerConfig *VaasBackfiller) {

	ctx := context.Background()

	// Load config
	cfg, err := config.NewRpcProviderSettingJson(backfillerConfig.RpcProvidersPath)
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	// create rpc pool
	rpcPool, wormchainRpcPool, err := newRpcPool(cfg)
	if err != nil {
		log.Fatal("Failed to initialize rpc pool: ", zap.Error(err))
	}

	logger := logger.New("wormhole-explorer-tx-tracker", logger.WithLevel(backfillerConfig.LogLevel))

	logger.Info("Starting wormhole-explorer-tx-tracker as vaas backfiller ...")

	startTime, err := time.Parse(time.RFC3339, backfillerConfig.StartTime)
	if err != nil {
		logger.Fatal("failed to parse start time", zap.Error(err))
	}

	endTime := time.Now()
	if backfillerConfig.EndTime != "" {
		endTime, err = time.Parse(time.RFC3339, backfillerConfig.EndTime)
		if err != nil {
			logger.Fatal("Failed to parse end time", zap.Error(err))
		}
	}

	if startTime.After(endTime) {
		logger.Fatal("Start time should be before end time",
			zap.String("start_time", startTime.Format(time.RFC3339)),
			zap.String("end_time", endTime.Format(time.RFC3339)))
	}

	//setup DB connection
	db, err := dbutil.Connect(ctx, logger, backfillerConfig.MongoURI, backfillerConfig.MongoDatabase, false)
	if err != nil {
		logger.Fatal("failed to connect MongoDB", zap.Error(err))
	}

	postreSQLDB, err := consumer.NewPostgreSQLRepository(ctx, "{postresql_url}")
	if err != nil {
		log.Fatal("Failed to initialize PostgreSQL client: ", err)
	}

	// create a vaa repository.
	vaaRepository := repository.NewVaaRepository(db.Database, logger)
	// create a consumer repository.
	globalTrxRepository := consumer.NewRepository(logger, db.Database)

	query := repository.VaaQuery{
		StartTime:      &startTime,
		EndTime:        &endTime,
		EmitterChainID: backfillerConfig.EmitterChainID,
		EmitterAddress: backfillerConfig.EmitterAddress,
	}

	limiter := ratelimit.New(int(backfillerConfig.RequestsPerMinute), ratelimit.Per(time.Minute))

	pagination := repository.Pagination{
		Page:     0,
		PageSize: backfillerConfig.PageSize,
		SortAsc:  true,
	}

	queue := make(chan *repository.VaaDoc, 5*backfillerConfig.PageSize)

	var quantityProduced, quantityConsumedWithError, quantityConsumedSuccess atomic.Uint64

	go getVaas(ctx, logger, pagination, query, vaaRepository, queue, &quantityProduced)

	var wg sync.WaitGroup
	wg.Add(backfillerConfig.NumWorkers)

	for i := 0; i < backfillerConfig.NumWorkers; i++ {
		p := vaasBackfillerParams{
			wg:                          &wg,
			logger:                      logger.With(zap.Int("worker", i)),
			rpcPool:                     rpcPool,
			queue:                       queue,
			wormchainRpcPool:            wormchainRpcPool,
			repository:                  globalTrxRepository,
			p2pNetwork:                  backfillerConfig.P2pNetwork,
			limiter:                     limiter,
			overwrite:                   backfillerConfig.Overwrite,
			disableDBUpsert:             backfillerConfig.DisableDBUpsert,
			processedDocumentsSuccess:   &quantityConsumedSuccess,
			processedDocumentsWithError: &quantityConsumedWithError,
		}
		go processVaa(ctx, &p, postreSQLDB)
	}

	logger.Info("Waiting for all workers to finish...")
	wg.Wait()

	logger.Info("closing MongoDB connection...")
	db.DisconnectWithTimeout(10 * time.Second)

	logger.Info("Finish wormhole-explorer-tx-tracker as vaas backfiller",
		zap.Uint64("produced", quantityProduced.Load()),
		zap.Uint64("consumer_success", quantityConsumedSuccess.Load()),
		zap.Uint64("consumed_error", quantityConsumedWithError.Load()))
}

func getVaas(ctx context.Context, logger *zap.Logger, pagination repository.Pagination, query repository.VaaQuery,
	vaaRepository *repository.VaaRepository, queue chan *repository.VaaDoc, quantityProduced *atomic.Uint64) {
	defer close(queue)
	for {
		logger.Info("Processing page", zap.Any("pagination", pagination), zap.Any("query", query))

		vaas, err := vaaRepository.FindPage(ctx, query, pagination)
		if err != nil {
			logger.Error("Failed to get vaas", zap.Error(err))
			break
		}

		if len(vaas) == 0 {
			logger.Info("Empty page", zap.Int64("page", pagination.Page))
			break
		}

		for _, vaa := range vaas {
			queue <- vaa
			quantityProduced.Add(1)
		}

		pagination.Page++
	}
	for {
		select {
		case <-time.After(10 * time.Second):
			if len(queue) == 0 {
				logger.Info("Closing, queue is empty")
				return
			}
		case <-ctx.Done():
			logger.Info("Closing due to cancelled context")
			return
		}
	}
}

func processVaa(ctx context.Context, params *vaasBackfillerParams, postresqlDB consumer.PostgreSQLRepository) {
	// Main loop: fetch global txs and process them
	metrics := metrics.NewDummyMetrics()
	defer params.wg.Done()
	for {
		select {

		// Try to pop a globalTransaction from the queue
		case v, ok := <-params.queue:
			// If the channel was closed, exit immediately
			if !ok {
				params.logger.Info("Closing, channel was closed")
				return
			}

			params.limiter.Take()

			p := consumer.ProcessSourceTxParams{
				TrackID:         "backfiller",
				Timestamp:       v.Timestamp,
				VaaId:           v.ID,
				ChainId:         sdk.ChainID(v.ChainID),
				Emitter:         v.EmitterAddress,
				Sequence:        v.Sequence,
				TxHash:          v.TxHash,
				Overwrite:       params.overwrite,
				Vaa:             v.Vaa,
				IsVaaSigned:     true,
				Metrics:         metrics,
				DisableDBUpsert: params.disableDBUpsert,
			}
			_, err := consumer.ProcessSourceTx(ctx, params.logger, params.rpcPool, params.wormchainRpcPool, params.repository, &p, params.p2pNetwork, postresqlDB)
			if err != nil {
				if errors.Is(err, consumer.ErrAlreadyProcessed) {
					params.logger.Info("Source tx was already processed", zap.String("vaaId", v.ID))
					params.processedDocumentsSuccess.Add(1)
					continue
				}
				params.logger.Error("Failed to process source tx",
					zap.String("vaaId", v.ID),
					zap.Error(err),
				)
				params.processedDocumentsWithError.Add(1)
				continue
			} else {
				params.processedDocumentsSuccess.Add(1)
				params.logger.Info("Processed source tx", zap.String("vaaId", v.ID))
			}

			// If the context was cancelled, exit immediately
		case <-ctx.Done():
			params.logger.Info("Closing due to cancelled context")
			return
		}
	}
}

func newRpcPool(cfg *config.RpcProviderSettingsJson) (map[sdk.ChainID]*pool.Pool, map[sdk.ChainID]*pool.Pool, error) {

	if cfg == nil {
		return nil, nil, errors.New("rpc provider settings is nil")
	}

	rpcConfigMap, err := cfg.ToMap()
	if err != nil {
		return nil, nil, err
	}
	wormchainRpcConfigMap, err := cfg.WormchainToMap()
	if err != nil {
		return nil, nil, err
	}

	domains := []string{".network", ".cloud", ".com", ".io", ".build", ".team", ".dev", ".zone", ".org", ".net", ".in"}
	// convert rpc settings map to rpc pool
	convertFn := func(rpcConfig []config.RpcConfig) []pool.Config {
		poolConfigs := make([]pool.Config, 0, len(rpcConfig))
		for _, rpc := range rpcConfig {
			poolConfigs = append(poolConfigs, pool.Config{
				Id:                rpc.Url,
				Priority:          rpc.Priority,
				Description:       utils.FindSubstringBeforeDomains(rpc.Url, domains),
				RequestsPerMinute: rpc.RequestsPerMinute,
			})
		}
		return poolConfigs
	}

	// create rpc pool
	rpcPool := make(map[sdk.ChainID]*pool.Pool)
	for chainID, rpcConfig := range rpcConfigMap {
		rpcPool[chainID] = pool.NewPool(convertFn(rpcConfig))
	}

	// create wormchain rpc pool
	wormchainRpcPool := make(map[sdk.ChainID]*pool.Pool)
	for chainID, rpcConfig := range wormchainRpcConfigMap {
		wormchainRpcPool[chainID] = pool.NewPool(convertFn(rpcConfig))
	}

	return rpcPool, wormchainRpcPool, nil
}
