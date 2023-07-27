package main

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbhelpers"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type TxHashEncondingConfig struct {
	LogLevel      string `env:"LOG_LEVEL,required"`
	MongoURI      string `env:"MONGODB_URI,required"`
	MongoDatabase string `env:"MONGODB_DATABASE,required"`
	ChainID       uint16 `env:"CHAIN_ID,required"`
	PageSize      int64  `env:"PAGE_SIZE,required"`
}

func RunTxHashEncoding(cfg TxHashEncondingConfig) {
	ctx := context.Background()
	logger := logger.New("wormhole-fly", logger.WithLevel(cfg.LogLevel))

	db, err := dbhelpers.Connect(ctx, logger, cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		logger.Fatal("could not connect to DB", zap.Error(err))
	}
	defer db.DisconnectWithTimeout(10 * time.Second)

	repository := storage.NewRepository(alert.NewDummyClient(), metrics.NewDummyMetrics(), db.Database, logger)

	workerTxHashEncoding(ctx, logger, repository, vaa.ChainID(cfg.ChainID), cfg.PageSize)
}

func workerTxHashEncoding(ctx context.Context, logger *zap.Logger, repo *storage.Repository, chainID vaa.ChainID, pageSize int64) {

	log := logger.With(zap.String("chainID", chainID.String()))
	log.Info("Processing chain")
	page := int64(0)
	for {
		log.Info("Processing page", zap.Int64("page", page))

		vaas, err := repo.FindVaaByChainID(ctx, chainID, page, pageSize)
		if err != nil {
			log.Error("Failed to get vaas", zap.Error(err))
			break
		}

		if len(vaas) == 0 {
			log.Info("Empty page", zap.Int64("page", page))
			break
		}
		for _, v := range vaas {
			l := log.With(zap.String("vaaId", v.ID), zap.String("txHash", v.TxHash))
			// check if txHash is already processed
			if v.OriginTxHash != nil && *v.OriginTxHash != "" {
				l.Debug("Already processed")
				continue
			}
			// check if txHash is not a hexadecimal, ignore it
			if len(v.TxHash) != 64 && len(v.TxHash) != 66 {
				l.Debug("txHash is not hexadecimal, ignore it")
				continue
			}

			hexTxHash, err := hex.DecodeString(v.TxHash)
			// txHash is not hex
			if err != nil {
				l.Error("txHash can not decode to hexadecimal", zap.Error(err))
				continue
			}

			newTxHash, err := domain.EncodeTrxHashByChainID(chainID, hexTxHash)
			if err != nil {
				l.Error("Failed to encode txHash", zap.String("vaaId", v.ID), zap.String("txHash", v.TxHash), zap.Error(err))
				continue
			}
			err = repo.ReplaceVaaTxHash(ctx, v.ID, v.TxHash, newTxHash)
			if err != nil {
				l.Error("replacing txHash", zap.Error(err))
				continue
			}
			l.Debug("Processing vaa")
		}
		page++
	}
}
