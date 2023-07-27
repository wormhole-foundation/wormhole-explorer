package dbhelpers

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

// Session is a plain-old-data struct that represents a handle to a MongoDB database.
type Session struct {
	Client   *mongo.Client
	Database *mongo.Database
	logger   *zap.Logger
}

// Connect to a MongoDB database.
func Connect(
	ctx context.Context,
	logger *zap.Logger,
	uri string,
	databaseName string,
) (*Session, error) {

	// Create a timed sub-context for the connection attempt
	const connectTimeout = 10 * time.Second
	subContext, cancelFunc := context.WithTimeout(ctx, connectTimeout)
	defer cancelFunc()

	// Connect to MongoDB
	client, err := mongo.Connect(subContext, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to make sure we're actually connected
	//
	// This can detect a misconfuiguration error when a service is being initialized,
	// rather than waiting for the first query to fail in the service's processing loop.
	err = client.Ping(subContext, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB database: %w", err)
	}

	// Populate the result struct and return
	db := &Session{
		Client:   client,
		Database: client.Database(databaseName),
	}
	return db, nil
}

// Disconnect from a MongoDB database.
func (s *Session) DisconnectWithTimeout(timeout time.Duration) error {

	// Create a timed sub-context for the disconnection attempt
	subContext, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()

	// Attempt to disconnect
	err := s.Client.Disconnect(subContext)
	if err != nil {
		s.logger.Warn(
			"failed to disconnect from MongoDB",
			zap.Duration("timeout", timeout),
			zap.Error(err),
		)
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	return nil
}
