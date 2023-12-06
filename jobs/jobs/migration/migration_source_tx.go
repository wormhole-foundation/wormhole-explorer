package migration

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	txtrackerProcessVaa "github.com/wormhole-foundation/wormhole-explorer/common/client/txtracker"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// MigrateSourceChainTx is the job to migrate vaa txHash and timestamp from vaa collection to globalTx collection.
type MigrateSourceChainTx struct {
	db                 *mongo.Database
	pageSize           int
	chainID            sdk.ChainID
	txTrackerAPIClient txtrackerProcessVaa.TxTrackerAPIClient
	sleepTime          time.Duration
	collections        struct {
		vaas               *mongo.Collection
		globalTransactions *mongo.Collection
	}
	logger *zap.Logger
}

// NewMigrationSourceChainTx creates a new migration job.
func NewMigrationSourceChainTx(
	db *mongo.Database,
	pageSize int,
	chainID sdk.ChainID,
	txTrackerAPIClient txtrackerProcessVaa.TxTrackerAPIClient,
	sleepTime time.Duration,
	logger *zap.Logger) *MigrateSourceChainTx {
	return &MigrateSourceChainTx{
		db:                 db,
		pageSize:           pageSize,
		chainID:            chainID,
		txTrackerAPIClient: txTrackerAPIClient,
		sleepTime:          sleepTime,
		collections: struct {
			vaas               *mongo.Collection
			globalTransactions *mongo.Collection
		}{
			vaas:               db.Collection("vaas"),
			globalTransactions: db.Collection("globalTransactions3"),
		},
		logger: logger}
}

// VAASourceChain defines the structure of vaa fields needed for migration.
type VAASourceChain struct {
	ID           string      `bson:"_id"`
	EmitterChain sdk.ChainID `bson:"emitterChain" json:"emitterChain"`
	Timestamp    *time.Time  `bson:"timestamp" json:"timestamp"`
	TxHash       *string     `bson:"txHash" json:"txHash,omitempty"`
}

// GlobalTransaction represents a global transaction.
type GlobalTransaction struct {
	ID       string    `bson:"_id" json:"id"`
	OriginTx *OriginTx `bson:"originTx" json:"originTx"`
}

// OriginTx represents a origin transaction.
type OriginTx struct {
	TxHash string `bson:"nativeTxHash" json:"txHash"`
	From   string `bson:"from" json:"from"`
	Status string `bson:"status" json:"status"`
}

func (m *MigrateSourceChainTx) Run(ctx context.Context) error {
	if m.chainID == sdk.ChainIDSolana || m.chainID == sdk.ChainIDAptos {
		return m.runComplexMigration(ctx)
	} else {
		return m.runMigration(ctx)
	}
}

// runComplexMigration runs the migration job for solana and aptos chains calling the txtracker endpoint.
func (m *MigrateSourceChainTx) runComplexMigration(ctx context.Context) error {
	if vaa.ChainIDSolana != m.chainID && vaa.ChainIDAptos != m.chainID {
		return errors.New("invalid chainID")
	}

	var page int64 = 0
	for {
		// get vaas to migrate by page and pageSize.
		vaas, err := m.getVaasToMigrate(ctx, m.chainID, page, int64(m.pageSize))
		if err != nil {
			m.logger.Error("failed to get vaas", zap.Error(err), zap.Int64("page", page))
			break
		}

		if len(vaas) == 0 {
			break
		}

		for _, v := range vaas {

			// check if global transaction exists and nested originTx exists
			filter := bson.D{
				{Key: "_id", Value: v.ID},
				{Key: "originTx", Value: bson.D{{Key: "$exists", Value: true}}},
			}

			var globalTransacations GlobalTransaction
			err := m.collections.globalTransactions.FindOne(ctx, filter).Decode(&globalTransacations)

			// if global transaction exists, skip
			if err == nil {
				m.logger.Info("global transaction already exists", zap.String("id", v.ID))
				continue
			}

			// if exist and error getting global transaction, log error
			if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
				m.logger.Error("failed to get global transaction", zap.Error(err), zap.String("id", v.ID))
				continue
			}

			// if not exist txhash, skip
			if v.TxHash == nil {
				m.logger.Error("txHash is nil", zap.String("id", v.ID))
				continue
			}

			_, err = m.txTrackerAPIClient.Process(v.ID)
			if err != nil {
				m.logger.Error("failed to process vaa", zap.Error(err), zap.String("id", v.ID))
				continue
			}
			time.Sleep(5 * time.Second)
		}
		page++
	}
	return nil
}

// Run runs the migration job.
func (m *MigrateSourceChainTx) runMigration(ctx context.Context) error {
	var page int64 = 0
	var wg sync.WaitGroup
	workerLimit := m.pageSize
	jobs := make(chan VAASourceChain, workerLimit)

	for i := 1; i <= workerLimit; i++ {
		wg.Add(1)
		go worker(ctx, &wg, jobs, m.collections.globalTransactions, m.logger)
	}

	for {

		// get vaas to migrate by page and pageSize.
		vaas, err := m.getVaasToMigrate(ctx, m.chainID, page, int64(m.pageSize))
		if err != nil {
			m.logger.Error("failed to get vaas", zap.Error(err), zap.Int64("page", page))
			break
		}

		if len(vaas) == 0 {
			break
		}

		for _, v := range vaas {
			jobs <- v
		}

		page++
	}
	close(jobs)
	wg.Wait()

	return nil
}

func worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan VAASourceChain, collection *mongo.Collection, logger *zap.Logger) {
	defer wg.Done()
	for v := range jobs {
		if v.EmitterChain == sdk.ChainIDSolana || v.EmitterChain == sdk.ChainIDAptos {
			logger.Debug("skip migration", zap.String("id", v.ID), zap.String("chain", v.EmitterChain.String()))
			continue
		}

		// check if global transaction exists and nested originTx exists
		filter := bson.D{
			{Key: "_id", Value: v.ID},
			{Key: "originTx", Value: bson.D{{Key: "$exists", Value: true}}},
		}

		var globalTransacations GlobalTransaction
		err := collection.FindOne(ctx, filter).Decode(&globalTransacations)

		// if global transaction exists, skip
		if err == nil {
			logger.Info("global transaction already exists", zap.String("id", v.ID))
			continue
		}

		// if exist and error getting global transaction, log error
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			logger.Error("failed to get global transaction", zap.Error(err), zap.String("id", v.ID))
			continue
		}

		// if not exist txhash, skip
		if v.TxHash == nil {
			logger.Error("txHash is nil", zap.String("id", v.ID))
			continue
		}

		// set txHash format by chain
		var txHash string
		switch v.EmitterChain {
		case sdk.ChainIDAcala,
			sdk.ChainIDArbitrum,
			sdk.ChainIDAvalanche,
			sdk.ChainIDBase,
			sdk.ChainIDBSC,
			sdk.ChainIDCelo,
			sdk.ChainIDEthereum,
			sdk.ChainIDFantom,
			sdk.ChainIDKarura,
			sdk.ChainIDKlaytn,
			sdk.ChainIDMoonbeam,
			sdk.ChainIDOasis,
			sdk.ChainIDOptimism,
			sdk.ChainIDPolygon:
			txHash = txHashLowerCaseWith0x(*v.TxHash)
		default:
			txHash = *v.TxHash
		}

		// update global transaction
		update := bson.D{
			{Key: "$set", Value: bson.D{{Key: "originTx.timestamp", Value: v.Timestamp}}},
			{Key: "$set", Value: bson.D{{Key: "originTx.nativeTxHash", Value: txHash}}},
			{Key: "$set", Value: bson.D{{Key: "originTx.status", Value: "confirmed"}}},
		}

		opts := options.Update().SetUpsert(true)
		result, err := collection.UpdateByID(ctx, v.ID, update, opts)
		if err != nil {
			logger.Error("failed to update global transaction", zap.Error(err), zap.String("id", v.ID))
			break
		}
		if result.UpsertedCount == 1 {
			logger.Info("inserted global transaction", zap.String("id", v.ID))
		} else {
			logger.Debug("global transaction already exists", zap.String("id", v.ID))
		}
	}
}

func txHashLowerCaseWith0x(v string) string {
	if strings.HasPrefix(v, "0x") {
		return strings.ToLower(v)
	}
	return "0x" + strings.ToLower(v)
}

func (m *MigrateSourceChainTx) getVaasToMigrate(ctx context.Context, chainID vaa.ChainID, page int64, pageSize int64) ([]VAASourceChain, error) {
	// define query parameter
	var filter bson.D
	if chainID != vaa.ChainIDUnset {
		filter = bson.D{{Key: "emitterChain", Value: chainID}}
	}

	skip := page * pageSize
	limit := pageSize
	sort := bson.D{{Key: "timestamp", Value: 1}}

	// define projection
	projection := bson.D{
		{Key: "_id", Value: 1},
		{Key: "emitterChain", Value: 1},
		{Key: "timestamp", Value: 1},
		{Key: "txHash", Value: 1},
	}

	// define options
	opts := options.
		Find().
		SetProjection(projection).
		SetLimit(limit).
		SetSkip(skip).
		SetSort(sort)

	// find vaas
	cur, err := m.collections.vaas.Find(ctx, filter, opts)
	if err != nil {
		return []VAASourceChain{}, err
	}

	// decode vaas
	vaas := make([]VAASourceChain, pageSize)
	if err := cur.All(ctx, &vaas); err != nil {
		return []VAASourceChain{}, err
	}

	return vaas, nil
}
