package heartbeats

import "time"

// HeartbeatDoc represent an heartbeat document.
type HeartbeatDoc struct {
	ID            string             `bson:"_id" json:"id"`
	BootTimestamp int64              `bson:"boottimestamp" json:"bootTimestamp"`
	Counter       int64              `bson:"counter" json:"counter"`
	Features      []string           `bson:"features" json:"features"`
	IndexedAt     *time.Time         `bson:"indexedAt" json:"indexedAt"`
	GuardianAddr  string             `bson:"guardianaddr" json:"guardianAddr"`
	NodeName      string             `bson:"nodename" json:"nodeName"`
	Timestamp     int64              `bson:"timestamp" json:"timestamp"`
	UpdatedAt     *time.Time         `bson:"updatedAt" json:"updatedAt"`
	Version       string             `bson:"version" json:"version"`
	Networks      []HeartbeatNetwork `bson:"networks" json:"networks"`
}

type heartbeatSQL struct {
	ID            string             `db:"id"`
	BootTimestamp *time.Time         `db:"boot_timestamp"`
	Counter       int64              `db:"counter"`
	Features      []string           `db:"feature"`
	IndexedAt     *time.Time         `db:"created_at"`
	GuardianAddr  string             `db:"guardianaddr"`
	NodeName      string             `db:"guardian_name"`
	Timestamp     *time.Time         `db:"timestamp"`
	UpdatedAt     *time.Time         `db:"updated_at"`
	Version       string             `db:"version"`
	Networks      []HeartbeatNetwork `db:"networks"`
}

// HeartbeatNetwork definition.
type HeartbeatNetwork struct {
	ID              int64  `db:"id" bson:"id" json:"id"`
	Height          int64  `db:"height" bson:"height" json:"height"`
	ContractAddress string `db:"contractaddress" bson:"contractaddress" json:"contractAddress"`
	ErrorCount      int64  `db:"errorcount" bson:"errorcount" json:"errorCount"`
}
