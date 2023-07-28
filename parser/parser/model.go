package parser

import (
	"time"

	vaaPayloadParser "github.com/wormhole-foundation/wormhole-explorer/common/client/parser"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// ParsedVaaUpdate represent a parsed vaa update.
type ParsedVaaUpdate struct {
	ID                        string                                  `bson:"_id" json:"id"`
	EmitterChain              sdk.ChainID                             `bson:"emitterChain" json:"emitterChain"`
	EmitterAddr               string                                  `bson:"emitterAddr" json:"emitterAddr"`
	Sequence                  string                                  `bson:"sequence" json:"sequence"`
	AppIDs                    []string                                `bson:"appIds" json:"appIds"`
	ParsedPayload             interface{}                             `bson:"parsedPayload" json:"parsedPayload"`
	RawStandardizedProperties vaaPayloadParser.StandardizedProperties `bson:"rawStandardizedProperties" json:"rawStandardizedProperties"`
	StandardizedProperties    vaaPayloadParser.StandardizedProperties `bson:"standardizedProperties" json:"standardizedProperties"`
	UpdatedAt                 *time.Time                              `bson:"updatedAt" json:"updatedAt"`
	Timestamp                 time.Time                               `bson:"-" json:"-"`
}
