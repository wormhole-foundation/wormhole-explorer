package track

import (
	"fmt"

	"github.com/google/uuid"
)

// GetTrackID returns a unique track id for the pipeline.
func GetTrackID(vaaID string) string {
	uuid := uuid.New()
	return fmt.Sprintf("gossip-signed-vaa-%s-%s", vaaID, uuid.String())
}
