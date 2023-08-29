package producer

import (
	"context"
	"time"
)

// PushFunc is a function to push VAAEvent.
type PushFunc func(context.Context, *NotificationEvent) error

type NotificationEvent struct {
	TrackID string    `json:"trackId"`
	Source  string    `json:"source"`
	Type    string    `json:"type"`
	Payload SignedVaa `json:"payload"`
}

type SignedVaa struct {
	ID               string    `json:"id"`
	EmitterChain     uint16    `json:"emitterChain"`
	EmitterAddr      string    `json:"emitterAddr"`
	Sequence         uint64    `json:"sequence"`
	GuardianSetIndex uint32    `json:"guardianSetIndex"`
	Timestamp        time.Time `json:"timestamp"`
	Vaa              []byte    `json:"vaa"`
	TxHash           string    `json:"txHash"`
	Version          int       `json:"version"`
}

// NewComposite returns a PushFunc that calls all the given producers.
func NewComposite(producers ...PushFunc) PushFunc {
	return func(ctx context.Context, event *NotificationEvent) error {
		for _, producer := range producers {
			if err := producer(ctx, event); err != nil {
				return err
			}
		}
		return nil
	}
}
