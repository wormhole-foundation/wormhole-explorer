package builder

import (
	"context"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/analytics/config"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/storage"
	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"go.uber.org/zap"
)

type StorageLayer struct {
	mongoDB          *dbutil.Session
	postgresDB       *db.DB
	pricesRepository storage.PricesRepository
	vaaRepository    storage.VaaRepository
}

func NewStorageLayer(ctx context.Context, dbLayer string, mongodbURI string, mongodbDatabase string,
	dbURL string, dbLogEnable bool, metrics metrics.Metrics, logger *zap.Logger) (*StorageLayer, error) {
	var storageLayer StorageLayer
	var mongoDb *dbutil.Session
	var postgresDb *db.DB
	var err error
	switch dbLayer {
	case config.DbLayerMongo:
		mongoDb, err = dbutil.Connect(ctx, logger, mongodbURI, mongodbDatabase, false)
		if err != nil {
			return nil, err
		}
		storageLayer.mongoDB = mongoDb
		storageLayer.pricesRepository = storage.NewMongoPricesRepository(mongoDb.Database, metrics, logger)
		storageLayer.vaaRepository = storage.NewMongoVaaRepository(mongoDb.Database, logger)
	case config.DbLayerPostgres:
		postgresDb, err = newPostgresDatabase(ctx, dbURL, dbLogEnable, logger)
		if err != nil {
			return nil, err
		}
		storageLayer.postgresDB = postgresDb
		storageLayer.pricesRepository = storage.NewPostgresRepository(postgresDb, metrics, logger)
		storageLayer.vaaRepository = storage.NewPostgresVaaRepository(postgresDb, logger)
	case config.DbLayerDual:
		mongoDb, err = dbutil.Connect(ctx, logger, mongodbURI, mongodbDatabase, false)
		if err != nil {
			return nil, err
		}
		postgresDb, err = newPostgresDatabase(ctx, dbURL, dbLogEnable, logger)
		if err != nil {
			return nil, err
		}
		mongoPricesRepository := storage.NewMongoPricesRepository(mongoDb.Database, metrics, logger)
		postgresPricesRepository := storage.NewPostgresRepository(postgresDb, metrics, logger)
		mongoVaaRepository := storage.NewMongoVaaRepository(mongoDb.Database, logger)
		postgresVaaRepository := storage.NewPostgresVaaRepository(postgresDb, logger)

		storageLayer.mongoDB = mongoDb
		storageLayer.postgresDB = postgresDb
		storageLayer.vaaRepository = storage.NewVaaRepositoryComposite(mongoVaaRepository, postgresVaaRepository)
		storageLayer.pricesRepository = storage.NewPricesRepositoryComposite(mongoPricesRepository, postgresPricesRepository)
	default:
		return nil, fmt.Errorf("invalid db layer: %s", dbLayer)
	}

	return &storageLayer, nil
}

func newPostgresDatabase(ctx context.Context,
	dbURL string, dbLogEnable bool,
	logger *zap.Logger) (*db.DB, error) {

	// Enable database logging
	var options db.Option
	if dbLogEnable {
		options = db.WithTracer(logger)
	}

	return db.NewDB(ctx, dbURL, options)
}

func (s *StorageLayer) Close() {
	if s.mongoDB != nil {
		s.mongoDB.DisconnectWithTimeout(10 * time.Second)
	}
	if s.postgresDB != nil {
		s.postgresDB.Close()
	}
}

func (s *StorageLayer) HealthChecks() []health.Check {
	var checks []health.Check
	if s.mongoDB != nil {
		checks = append(checks, health.Mongo(s.mongoDB.Database))
	}
	if s.postgresDB != nil {
		checks = append(checks, health.Postgres(s.postgresDB))
	}
	return checks
}

func (s *StorageLayer) PricesRepository() storage.PricesRepository {
	if s.pricesRepository != nil {
		return s.pricesRepository
	}
	panic("no repository available")
}

func (s *StorageLayer) VaaRepository() storage.VaaRepository {
	if s.vaaRepository != nil {
		return s.vaaRepository
	}
	panic("no repository available")
}
