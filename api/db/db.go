package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect(ctx context.Context, url string) (*mongo.Client, error) {
	cli, err := mongo.NewClient(options.Client().ApplyURI(url))
	if err != nil {
		return cli, err
	}
	err = cli.Connect(ctx)
	return cli, err
}
