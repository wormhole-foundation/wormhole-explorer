package portalanalytic

import "github.com/wormhole-foundation/wormhole-explorer/common/storage"

// PortalAnalyticDoc is a portal analytic document.
type PortalAnalyticdDoc struct {
	ID    string         `bson:"_id" json:"id"`
	From  string         `bson:"from" json:"from"`
	To    string         `bson:"to" json:"to"`
	Value storage.Uint64 `bson:"value" json:"value"`
}
