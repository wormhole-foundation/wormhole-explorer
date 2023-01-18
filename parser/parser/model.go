package parser

import (
	"time"
)

// ParsedVaaUpdate representa a parsedVaa document.
type ParsedVaaUpdate struct {
	ID           string      `bson:"_id"`
	EmitterChain uint16      `bson:"emitterChain"`
	EmitterAddr  string      `bson:"emitterAddr"`
	Sequence     string      `bson:"sequence"`
	AppID        string      `bson:"appId"`
	Result       interface{} `bson:"result"`
	UpdatedAt    *time.Time  `bson:"updatedAt"`
}
