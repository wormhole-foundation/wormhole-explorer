package storage

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// GetDB connects to DB and returns a client that will disconnect when the passed in context is cancelled
func GetDB(appCtx context.Context, log *zap.Logger, uri, databaseName string) (*mongo.Database, error) {
	cli, err := mongo.Connect(appCtx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	go func() {
		<-appCtx.Done()
		ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
		err := cli.Disconnect(ctx)
		log.Error("error disconnecting from db", zap.Error(err))
	}()
	return cli.Database(databaseName), err
}
