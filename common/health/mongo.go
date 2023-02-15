package health

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoStatus struct {
	Ok          int32             `bson:"ok"`
	Host        string            `bson:"host"`
	Version     string            `bson:"version"`
	Process     string            `bson:"process"`
	Pid         int32             `bson:"pid"`
	Uptime      int32             `bson:"uptime"`
	Connections *mongoConnections `bson:"connections"`
}

// mongoConnections represents a mongo server connection.
type mongoConnections struct {
	Current      int32 `bson:"current"`
	Available    int32 `bson:"available"`
	TotalCreated int32 `bson:"totalCreated"`
}

func Mongo(db *mongo.Database) Check {
	return func(ctx context.Context) error {
		command := bson.D{{Key: "serverStatus", Value: 1}}
		result := db.RunCommand(ctx, command)
		if result.Err() != nil {
			return errors.WithStack(result.Err())
		}

		var mongoStatus mongoStatus
		err := result.Decode(&mongoStatus)
		if err != nil {
			return errors.WithStack(err)
		}
		// check mongo server status
		mongoStatusCheck := (mongoStatus.Ok == 1 && mongoStatus.Pid > 0 && mongoStatus.Uptime > 0)
		if !mongoStatusCheck {
			return fmt.Errorf("mongo server not ready (Ok = %v, Pid = %v, Uptime = %v)", mongoStatus.Ok, mongoStatus.Pid, mongoStatus.Uptime)
		}

		// check mongo connections
		if mongoStatus.Connections.Available <= 0 {
			return fmt.Errorf("mongo server without available connections (availableConection = %v)", mongoStatus.Connections.Available)
		}
		return nil
	}
}
