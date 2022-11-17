// Package db handle mongodb connections.
package db

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect create a new mongo db client for the options defined in the input param url.
func Connect(ctx context.Context, url string) (*mongo.Client, error) {
	cli, err := mongo.NewClient(options.Client().ApplyURI(url))
	if err != nil {
		return cli, err
	}
	err = cli.Connect(ctx)
	return cli, err
}
