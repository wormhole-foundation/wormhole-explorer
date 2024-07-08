package migration

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// MigrateSourceChainTx is the job to migrate vaa txHash and timestamp from vaa collection to globalTx collection.
type MigrateNativeTxHash struct {
	db          *mongo.Database
	pageSize    int
	collections struct {
		observations *mongo.Collection
	}
	logger *zap.Logger
}

// NewMigrationNativeTxHash creates a new migration job.
func NewMigrationNativeTxHash(
	db *mongo.Database,
	pageSize int,
	logger *zap.Logger) *MigrateNativeTxHash {
	return &MigrateNativeTxHash{
		db:       db,
		pageSize: pageSize,
		collections: struct {
			observations *mongo.Collection
		}{
			observations: db.Collection(repository.Observations),
		},
		logger: logger}
}

// GlobalTransaction represents a global transaction.
type Observation struct {
	ID        string      `bson:"_id" json:"id"`
	ChainID   vaa.ChainID `bson:"emitterChain"`
	TxHash    []byte      `bson:"txHash"`
	IndexedAt time.Time   `bson:"indexedAt"`
}

func (m *MigrateNativeTxHash) Run(ctx context.Context) error {
	return m.runMigration(ctx)
}

// Run runs the migration job.
func (m *MigrateNativeTxHash) runMigration(ctx context.Context) error {
	var updated atomic.Uint64
	var total atomic.Uint64
	var wg sync.WaitGroup
	workerLimit := m.pageSize
	jobs := make(chan Observation, workerLimit)

	for i := 1; i <= workerLimit; i++ {
		wg.Add(1)
		go updateNativeTxHash(ctx, &wg, jobs, m.collections.observations, &updated, m.logger)
	}

	indexedAt := time.Now()
	for {
		observations, err := m.getObservationsToMigrate(ctx, int64(m.pageSize), indexedAt)
		if err != nil {
			m.logger.Error("failed to get observations", zap.Error(err))
			break
		}

		if len(observations) == 0 {
			break
		}
		total.Add(uint64(len(observations)))
		for _, v := range observations {
			jobs <- v
			indexedAt = v.IndexedAt
		}
		m.logger.Info("migrating observations",
			zap.String("indexedAt", indexedAt.Format(time.RFC3339)),
			zap.Uint64("total", total.Load()),
			zap.Uint64("updated", updated.Load()))
	}
	close(jobs)
	wg.Wait()

	return nil
}

func updateNativeTxHash(ctx context.Context, wg *sync.WaitGroup, jobs <-chan Observation, collection *mongo.Collection, updated *atomic.Uint64, logger *zap.Logger) {
	defer wg.Done()
	for v := range jobs {
		var ignoreNativeTxHash bool
		var nativeTxHash string

		// if not exist txhash, skip
		if v.TxHash == nil || len(v.TxHash) == 0 {
			logger.Warn("txHash is nil", zap.String("id", v.ID))
			ignoreNativeTxHash = true
		}

		if !ignoreNativeTxHash {
			txHash, err := domain.EncodeTrxHashByChainID(v.ChainID, v.TxHash)
			if err != nil {
				logger.Error("failed to encode transaction hash", zap.Error(err), zap.String("id", v.ID))
			} else {
				nativeTxHash = txHash
			}
		}

		// update observations
		update := bson.D{
			{Key: "$set", Value: bson.D{{Key: "nativeTxHash", Value: nativeTxHash}}},
		}

		result, err := collection.UpdateByID(ctx, v.ID, update, &options.UpdateOptions{Upsert: &[]bool{true}[0]})
		if err != nil {
			logger.Error("failed to update observation", zap.Error(err), zap.String("id", v.ID))
			break
		}
		if result.ModifiedCount == 1 {
			updated.Add(1)
			logger.Debug("updated nativeTxHash observation", zap.String("id", v.ID))
		} else {
			logger.Info("nativeTxHash in observation already exists", zap.String("id", v.ID))
		}
	}
}

func (m *MigrateNativeTxHash) getObservationsToMigrate(ctx context.Context, pageSize int64, lessThan time.Time) ([]Observation, error) {

	limit := pageSize
	sort := bson.D{{Key: "indexedAt", Value: -1}}

	solanaAndAptosAndWormchainIds := []sdk.ChainID{sdk.ChainIDSolana, sdk.ChainIDAptos, sdk.ChainIDWormchain}
	filter := bson.D{
		{Key: "emitterChain", Value: bson.M{"$nin": solanaAndAptosAndWormchainIds}},
		{Key: "nativeTxHash", Value: bson.M{"$exists": false}},
		{Key: "indexedAt", Value: bson.M{"$lte": lessThan}},
	}

	cur, err := m.collections.observations.Find(ctx, filter, &options.FindOptions{Limit: &limit, Sort: sort})
	if err != nil {
		return []Observation{}, err
	}

	var observations []Observation
	if err := cur.All(ctx, &observations); err != nil {
		return []Observation{}, err
	}

	return observations, nil
}
