package queue

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/events"
	"go.uber.org/zap"
)

// VaaEvent represents a vaa data to be handle by the pipeline.
type VaaEvent struct {
	ID               string     `json:"id"`
	ChainID          uint16     `json:"emitterChain"`
	EmitterAddress   string     `json:"emitterAddr"`
	Sequence         string     `json:"sequence"`
	GuardianSetIndex uint32     `json:"guardianSetIndex"`
	Vaa              []byte     `json:"vaas"`
	IndexedAt        time.Time  `json:"indexedAt"`
	Timestamp        *time.Time `json:"timestamp"`
	UpdatedAt        *time.Time `json:"updatedAt"`
	TxHash           string     `json:"txHash"`
	Version          uint16     `json:"version"`
	Revision         uint16     `json:"revision"`
}

// VaaConverter converts a message from a VAAEvent.
func NewVaaConverter(log *zap.Logger) ConverterFunc {

	return func(msg string) (*Event, error) {
		// unmarshal message to vaaEvent
		var vaaEvent VaaEvent
		err := json.Unmarshal([]byte(msg), &vaaEvent)
		if err != nil {
			return nil, err
		}
		return &Event{
			Source:         "pipeline",
			TrackID:        fmt.Sprintf("pipeline-%s", vaaEvent.ID),
			ID:             vaaEvent.ID,
			ChainID:        vaaEvent.ChainID,
			EmitterAddress: vaaEvent.EmitterAddress,
			Sequence:       vaaEvent.Sequence,
			Vaa:            vaaEvent.Vaa,
			Timestamp:      vaaEvent.Timestamp,
			VaaIsSigned:    true,
		}, nil
	}
}

func NewNotificationEvent(log *zap.Logger) ConverterFunc {

	return func(msg string) (*Event, error) {
		// unmarshal message to NotificationEvent
		var notification events.NotificationEvent
		err := json.Unmarshal([]byte(msg), &notification)
		if err != nil {
			return nil, err
		}

		if notification.Event != events.SignedVaaType && notification.Event != events.LogMessagePublishedType {
			log.Debug("Skip event type", zap.String("trackId", notification.TrackID), zap.String("type", notification.Event))
			return nil, nil
		}

		switch notification.Event {
		case events.SignedVaaType:
			signedVaa, err := events.GetEventData[events.SignedVaa](&notification)
			if err != nil {
				log.Error("Error decoding signedVAA from notification event", zap.String("trackId", notification.TrackID), zap.Error(err))
				return nil, nil
			}

			return &Event{
				Source:         "chain-event",
				TrackID:        notification.TrackID,
				ID:             signedVaa.ID,
				ChainID:        signedVaa.EmitterChain,
				EmitterAddress: signedVaa.EmitterAddress,
				Sequence:       strconv.FormatUint(signedVaa.Sequence, 10),
				Timestamp:      &signedVaa.Timestamp,
				VaaIsSigned:    false,
			}, nil
		case events.LogMessagePublishedType:
			plm, err := events.GetEventData[events.LogMessagePublished](&notification)
			if err != nil {
				log.Error("Error decoding publishedLogMessage from notification event", zap.String("trackId", notification.TrackID), zap.Error(err))
				return nil, nil
			}

			vaa, err := events.CreateUnsignedVAA(&plm)
			if err != nil {
				log.Error("Error creating unsigned vaa", zap.String("trackId", notification.TrackID), zap.Error(err))
				return nil, err
			}

			vaaBytes, err := vaa.MarshalBinary()
			if err != nil {
				return nil, err
			}

			return &Event{
				Source:         "chain-event",
				TrackID:        notification.TrackID,
				ID:             vaa.MessageID(),
				ChainID:        plm.ChainID,
				EmitterAddress: plm.Attributes.Sender,
				Sequence:       strconv.FormatUint(plm.Attributes.Sequence, 10),
				Timestamp:      &plm.BlockTime,
				Vaa:            vaaBytes,
				VaaIsSigned:    false,
			}, nil
		}
		return nil, nil
	}
}
