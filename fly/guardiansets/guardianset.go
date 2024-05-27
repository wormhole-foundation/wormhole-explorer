package guardiansets

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/certusone/wormhole/node/pkg/common"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	flyAlert "github.com/wormhole-foundation/wormhole-explorer/fly/internal/alert"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type GuardianSetProvider interface {
	GetCurrentGuardianSetIndex(ctx context.Context) (uint32, error)
	GetGuardianSet(ctx context.Context, index uint32) (*common.GuardianSet, *time.Time, error)
	GetGuardianSetHistory(ctx context.Context) (*GuardianSetHistory, error)
	AddGuardianSet(ctx context.Context, gs *common.GuardianSet, et time.Time) error
}

// GuardianSetHistory contains information about all guardian sets for the current network (past and present).
type GuardianSetHistory struct {
	sync.RWMutex
	guardianSetsByIndex    []common.GuardianSet
	expirationTimesByIndex []time.Time
	alertClient            alert.AlertClient
}

// Verify takes a VAA as input and validates its guardian signatures.
func (h *GuardianSetHistory) Verify(ctx context.Context, vaa *sdk.VAA) error {
	idx := vaa.GuardianSetIndex

	h.RLock()
	lenGuardianSetsByIndex := uint32(len(h.guardianSetsByIndex))

	// Make sure the index exists
	if idx >= lenGuardianSetsByIndex {
		h.RUnlock()
		alertContext := alert.AlertContext{
			Details: map[string]string{
				"vaaID":               vaa.MessageID(),
				"vaaGuardianSetIndex": fmt.Sprint(vaa.GuardianSetIndex),
				"guardianSetIndex":    fmt.Sprint(lenGuardianSetsByIndex),
			},
		}
		_ = h.alertClient.CreateAndSend(ctx, flyAlert.GuardianSetUnknown, alertContext)
		return fmt.Errorf("guardian Set Index is out of bounds: got %d, max is %d",
			vaa.GuardianSetIndex,
			lenGuardianSetsByIndex,
		)
	}

	keysGuardianSets := h.guardianSetsByIndex[idx].Keys
	h.RUnlock()

	// Verify guardian signatures
	if vaa.VerifySignatures(keysGuardianSets) {
		return nil
	} else {
		return errors.New("VAA contains invalid signatures")
	}
}

// GetLatest returns the lastest guardian set.
func (h *GuardianSetHistory) GetLatest() common.GuardianSet {
	h.RLock()
	defer h.RUnlock()
	return h.guardianSetsByIndex[len(h.guardianSetsByIndex)-1]
}

func (h *GuardianSetHistory) Add(gs common.GuardianSet, t time.Time) {
	h.Lock()
	h.guardianSetsByIndex = append(h.guardianSetsByIndex, gs)
	h.expirationTimesByIndex = append(h.expirationTimesByIndex, t)
	h.Unlock()
}
