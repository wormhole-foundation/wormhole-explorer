package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/utils"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/repository/vaa"

	"github.com/wormhole-foundation/wormhole-explorer/common/events"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

// VaaEvent represents a vaa data to be handled by the pipeline.
type VaaEvent struct {
	ID               string      `json:"id"`
	VaaID            string      `json:"vaaId"`
	ChainID          sdk.ChainID `json:"emitterChain"`
	EmitterAddress   string      `json:"emitterAddr"`
	Sequence         string      `json:"sequence"`
	GuardianSetIndex uint32      `json:"guardianSetIndex"`
	Vaa              []byte      `json:"vaas"`
	IndexedAt        time.Time   `json:"indexedAt"`
	Timestamp        *time.Time  `json:"timestamp"`
	UpdatedAt        *time.Time  `json:"updatedAt"`
	TxHash           string      `json:"txHash"`
	Version          uint16      `json:"version"`
	Revision         uint16      `json:"revision"`
	Overwrite        bool        `json:"overwrite"`
}

// VaaConverter converts a message from a VAAEvent.
func NewVaaConverter(_ *zap.Logger) ConverterFunc {

	return func(ctx context.Context, msg string) (*Event, error) {
		// unmarshal message to vaaEvent
		var vaaEvent VaaEvent
		err := json.Unmarshal([]byte(msg), &vaaEvent)
		if err != nil {
			return nil, err
		}

		return &Event{
			Source:         "pipeline",
			TrackID:        fmt.Sprintf("pipeline-%s", vaaEvent.ID),
			Type:           SourceChainEvent,
			ID:             vaaEvent.ID, // digest
			VaaID:          vaaEvent.VaaID,
			ChainID:        vaaEvent.ChainID,
			EmitterAddress: vaaEvent.EmitterAddress,
			Sequence:       vaaEvent.Sequence,
			Timestamp:      vaaEvent.Timestamp,
			Vaa:            vaaEvent.Vaa,
			IsVaaSigned:    true,
			TxHash:         vaaEvent.TxHash,
			Overwrite:      vaaEvent.Overwrite,
		}, nil
	}
}

func NewNotificationEvent(vaaRepository vaa.VAARepository, log *zap.Logger) ConverterFunc {

	return func(ctx context.Context, msg string) (*Event, error) {
		// unmarshal message to NotificationEvent
		var notification events.NotificationEvent
		err := json.Unmarshal([]byte(msg), &notification)
		if err != nil {
			return nil, err
		}

		switch notification.Event {
		case events.SignedVaaType,
			events.LogMessagePublishedType,
			events.EvmTransactionFoundType,
			events.TransferRedeemedType:
			//message is valid
		default:
			log.Debug("Skip event type", zap.String("trackId", notification.TrackID), zap.String("type", notification.Event))
			return nil, nil
		}

		switch notification.Event {
		case events.SignedVaaType:
			// add log to check if we can remove this event type.
			log.Error("Event SignedVaaType is not supported", zap.String("trackId", notification.TrackID))

			signedVaa, err := events.GetEventData[events.SignedVaa](&notification)
			if err != nil {
				log.Error("Error decoding signedVAA from notification event", zap.String("trackId", notification.TrackID), zap.Error(err))
				return nil, nil
			}

			digest, err := getVAADigest(ctx, signedVaa.ID, vaaRepository, log, notification)
			if err != nil {
				return nil, err
			}

			return &Event{
				Source:         "chain-event",
				TrackID:        notification.TrackID,
				Type:           SourceChainEvent,
				ID:             digest,
				VaaID:          signedVaa.ID,
				ChainID:        sdk.ChainID(signedVaa.EmitterChain),
				EmitterAddress: signedVaa.EmitterAddress,
				Sequence:       strconv.FormatUint(signedVaa.Sequence, 10),
				Timestamp:      &signedVaa.Timestamp,
				IsVaaSigned:    false,
				TxHash:         signedVaa.TxHash,
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

			return &Event{
				Source:         "chain-event",
				TrackID:        notification.TrackID,
				Type:           SourceChainEvent,
				ID:             utils.NormalizeHex(vaa.HexDigest()),
				VaaID:          vaa.MessageID(),
				ChainID:        sdk.ChainID(plm.ChainID),
				EmitterAddress: plm.Attributes.Sender,
				Sequence:       strconv.FormatUint(plm.Attributes.Sequence, 10),
				Timestamp:      &plm.BlockTime,
				IsVaaSigned:    false,
				TxHash:         plm.TxHash,
			}, nil
		case events.EvmTransactionFoundType:
			// add log to check if we can remove this event type.
			log.Error("Event EvmTransactionFoundType is not supported", zap.String("trackId", notification.TrackID))

			tr, err := events.GetEventData[events.EvmTransactionFound](&notification)
			if err != nil {
				log.Error("Error decoding transferRedeemed from notification event", zap.String("trackId", notification.TrackID), zap.Error(err))
				return nil, nil
			}
			address, err := sdk.StringToAddress(tr.Attributes.EmitterAddress)
			if err != nil {
				return nil, fmt.Errorf("error converting emitter address [%s]: %w", tr.Attributes.EmitterAddress, err)
			}
			vaa := sdk.VAA{
				EmitterChain:   sdk.ChainID(tr.Attributes.EmitterChain),
				EmitterAddress: address,
				Sequence:       tr.Attributes.Sequence,
			}

			if tr.Attributes.Name != events.EvmTransferRedeemedName {
				log.Warn("Skip event because it is not transfer-redeemed ", zap.String("trackId", notification.TrackID), zap.String("name", tr.Attributes.Name))
				return nil, nil
			}

			return &Event{
				Source:         "chain-event",
				TrackID:        notification.TrackID,
				Type:           TargetChainEvent,
				ID:             utils.NormalizeHex(vaa.HexDigest()),
				VaaID:          vaa.MessageID(),
				ChainID:        sdk.ChainID(tr.ChainID),
				EmitterAddress: tr.Attributes.EmitterAddress,
				Sequence:       strconv.FormatUint(tr.Attributes.Sequence, 10),
				Timestamp:      &tr.BlockTime,
				TxHash:         tr.TxHash,
				Attributes: &TargetChainAttributes{
					Emitter:     tr.Emitter,
					BlockHeight: tr.BlockHeight,
					TxHash:      tr.TxHash,
					From:        tr.Attributes.From,
					To:          tr.Attributes.To,
					Method:      tr.Attributes.Method,
					Status:      tr.Attributes.Status,
				},
			}, nil
		case events.TransferRedeemedType:
			tr, err := events.GetEventData[events.TransferRedeemed](&notification)
			if err != nil {
				log.Error("Error decoding transferRedeemed from notification event", zap.String("trackId", notification.TrackID), zap.Error(err))
				return nil, nil
			}
			address, err := sdk.StringToAddress(tr.Attributes.EmitterAddress)
			if err != nil {
				return nil, fmt.Errorf("error converting emitter address [%s]: %w", tr.Attributes.EmitterAddress, err)
			}

			vaa := sdk.VAA{
				EmitterChain:   sdk.ChainID(tr.Attributes.EmitterChain),
				EmitterAddress: address,
				Sequence:       tr.Attributes.Sequence,
				Timestamp:      tr.BlockTime,
			}

			vaaDigest, err := getVAADigest(ctx, vaa.MessageID(), vaaRepository, log, notification)
			if err != nil {
				return nil, err
			}

			return &Event{
				Source:         "chain-event",
				TrackID:        notification.TrackID,
				Type:           TargetChainEvent,
				ID:             vaaDigest,
				VaaID:          vaa.MessageID(),
				ChainID:        sdk.ChainID(tr.ChainID),
				EmitterAddress: tr.Attributes.EmitterAddress,
				Sequence:       strconv.FormatUint(tr.Attributes.Sequence, 10),
				Timestamp:      &tr.BlockTime,
				TxHash:         tr.TxHash,
				Attributes: &TargetChainAttributes{
					Emitter:           tr.Emitter,
					BlockHeight:       tr.BlockHeight,
					TxHash:            tr.TxHash,
					From:              tr.Attributes.From,
					To:                tr.Attributes.To,
					Method:            tr.Attributes.Method,
					Status:            tr.Attributes.Status,
					GasUsed:           tr.Attributes.GasUsed,
					EffectiveGasPrice: tr.Attributes.EffectiveGasPrice,
					Fee:               tr.Attributes.Fee,
				},
			}, nil
		}

		return nil, nil
	}
}

func getVAADigest(ctx context.Context, vaaID string, vaaRepository vaa.VAARepository, log *zap.Logger, notification events.NotificationEvent) (string, error) {
	res, errGetVaa := vaaRepository.GetVaa(ctx, vaaID)
	if errGetVaa != nil {
		log.Error("Error getting vaa from repository", zap.String("trackId", notification.TrackID), zap.String("vaaID", vaaID), zap.Error(errGetVaa))
		return "", errGetVaa
	}
	return res.ID, errGetVaa
}
