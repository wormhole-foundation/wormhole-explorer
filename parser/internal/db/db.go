package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// Database definition.
type Database struct {
	Database *mongo.Database
	client   *mongo.Client
}

// New connects to DB and returns a client that will disconnect when the passed in context is cancelled.
func New(appCtx context.Context, log *zap.Logger, uri, databaseName string) (*Database, error) {
	cli, err := mongo.Connect(appCtx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return &Database{client: cli, Database: cli.Database(databaseName)}, err
}

// Close closes the database connections.
func (d *Database) Close() error {
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
	return d.client.Disconnect(ctx)
}
