package domain

import (
	"encoding/base64"
	"fmt"

	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

// CreateUniqueVaaID creates a unique VAA ID based on the message ID and the signing digest.
func CreateUniqueVaaID(vaa *sdk.VAA) string {
	hash := base64.StdEncoding.EncodeToString(vaa.SigningDigest().Bytes())
	return fmt.Sprintf("%s/%s", vaa.MessageID(), hash)
}

// CreateUniqueVaaIDByObservation creates a unique VAA ID based on the message ID and the observation hash.
func CreateUniqueVaaIDByObservation(obs *gossipv1.SignedObservation) string {
	hash := base64.StdEncoding.EncodeToString(obs.Hash)
	return fmt.Sprintf("%s/%s", obs.MessageId, hash)
}
