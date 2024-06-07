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

func GetTrackIDForDuplicatedVAA(vaaID string) string {
	uuid := uuid.New()
	return fmt.Sprintf("fly-duplicated-vaa-%s-%s", vaaID, uuid.String())
}

func GetTrackIDForGovernorStatus(nodeName string, timestamp int64) string {
	uuid := uuid.New()
	return fmt.Sprintf("fly-governor-status-%s-%v-%s", nodeName, timestamp, uuid.String())
}
