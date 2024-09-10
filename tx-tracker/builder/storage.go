package builder

import (
	"context"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/db"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/health"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/consumer"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/repository/vaa"
	"go.uber.org/zap"
)

type StorageLayerParams struct {
	DbLayer         config.DbLayer
	MongodbUri      string
	MongodbDatabase string
	DbUrl           string
	DbLogEnabled    bool
}

type StorageLayer struct {
	mongoDB       *dbutil.Session
	postgresDB    *db.DB
	repository    consumer.Repository
	vaaRepository vaa.VAARepository
}

func NewStorageLayer(ctx context.Context, params StorageLayerParams, logger *zap.Logger) (*StorageLayer, error) {
	var storageLayer StorageLayer
	var mongoDb *dbutil.Session
	var postgresDb *db.DB
	var err error
	switch params.DbLayer {
	case config.DbLayerMongo:
		mongoDb, err = dbutil.Connect(ctx, logger, params.MongodbUri, params.MongodbDatabase, false)
		if err != nil {
			return nil, err
		}
		storageLayer.mongoDB = mongoDb
		vaaRepository := vaa.NewMongoVaaRepository(mongoDb.Database, logger)
		storageLayer.repository = consumer.NewMongoRepository(logger, mongoDb.Database, vaaRepository)
		storageLayer.vaaRepository = vaaRepository
	case config.DbLayerPostgresql:
		postgresDb, err = newPostgresDatabase(ctx, params.DbLogEnabled, params.DbUrl, logger)
		if err != nil {
			return nil, err
		}
		vaaRepository := vaa.NewVaaRepositoryPostreSQL(postgresDb, logger)
		storageLayer.postgresDB = postgresDb
		storageLayer.repository = consumer.NewPostgreSQLRepository(postgresDb, vaaRepository)
		storageLayer.vaaRepository = vaaRepository
	case config.DbLayerDual:
		mongoDb, err = dbutil.Connect(ctx, logger, params.MongodbUri, params.MongodbDatabase, false)
		if err != nil {
			return nil, err
		}
		storageLayer.mongoDB = mongoDb
		mongoVaaRepository := vaa.NewMongoVaaRepository(mongoDb.Database, logger)
		mongoRepository := consumer.NewMongoRepository(logger, mongoDb.Database, mongoVaaRepository)
		postgresDb, err = newPostgresDatabase(ctx, params.DbLogEnabled, params.DbUrl, logger)
		if err != nil {
			return nil, err
		}
		storageLayer.postgresDB = postgresDb
		postgresVaaRepository := vaa.NewVaaRepositoryPostreSQL(postgresDb, logger)
		postgresRepository := consumer.NewPostgreSQLRepository(postgresDb, postgresVaaRepository)
		// create dual vaa repository
		storageLayer.vaaRepository = vaa.NewDualVaaRepository(mongoVaaRepository, postgresVaaRepository)
		// create dual repository
		storageLayer.repository = consumer.NewDualRepository(mongoRepository, postgresRepository)
	default:
		return nil, fmt.Errorf("unknown db layer: %s", params.DbLayer)
	}
	return &storageLayer, nil
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

func (s *StorageLayer) Repository() consumer.Repository {
	return s.repository
}

func (s *StorageLayer) VaaRepository() vaa.VAARepository {
	return s.vaaRepository
}

func newPostgresDatabase(ctx context.Context,
	dbLogEnabled bool, dbUrl string,
	logger *zap.Logger) (*db.DB, error) {
	var option db.Option
	if dbLogEnabled {
		option = db.WithTracer(logger)
	}
	return db.NewDB(ctx, dbUrl, option)
}
