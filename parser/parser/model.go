package parser

import (
	"time"

	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// VaaParserFunctions represent a vaaParserFunctions document.
type VaaParserFunctions struct {
	ID             string      `bson:"_id"`
	CreatedAt      *time.Time  `bson:"createdAt"`
	UpdatedAt      *time.Time  `bson:"updatedAt"`
	EmitterChain   vaa.ChainID `bson:"emitterChain"`
	EmitterAddress string      `bson:"emitterAddress"`
	ParserFunction string      `bson:"parserFunction"`
}
