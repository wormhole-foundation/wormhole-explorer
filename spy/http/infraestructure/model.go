package infraestructure

// MongoStatus represent a mongo server status.
type MongoStatus struct {
	Ok          int32             `bson:"ok"`
	Host        string            `bson:"host"`
	Version     string            `bson:"version"`
	Process     string            `bson:"process"`
	Pid         int32             `bson:"pid"`
	Uptime      int32             `bson:"uptime"`
	Connections *MongoConnections `bson:"connections"`
}

// MongoConnections represents a mongo server connection.
type MongoConnections struct {
	Current      int32 `bson:"current"`
	Available    int32 `bson:"available"`
	TotalCreated int32 `bson:"totalCreated"`
}
