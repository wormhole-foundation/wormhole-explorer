package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/internal/repository/vaa"
	"strconv"
	"time"

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
func NewVaaConverter(log *zap.Logger) ConverterFunc {

	return func(ctx context.Context, msg string) (*Event, error) {
		// unmarshal message to vaaEvent
		var vaaEvent VaaEvent
		err := json.Unmarshal([]byte(msg), &vaaEvent)
		if err != nil {
			return nil, err
		}
		//"{\"Message\":\"{\\\"id\\\":\\\"3104/aeb534c45c3049d380b9d9b966f9895f53abd4301bfaff407fa09dea8ae7a924/39762\\\",\\\"vaaId\\\":\\\"c13b999ecea3c39781d6bf3aa9ce5552ed9399da3a3f680826b982d21daec825\\\",\\\"emitterChain\\\":3104,\\\"emitterAddr\\\":\\\"aeb534c45c3049d380b9d9b966f9895f53abd4301bfaff407fa09dea8ae7a924\\\",\\\"sequence\\\":\\\"39762\\\",\\\"guardianSetIndex\\\":4,\\\"vaas\\\":\\\"AQAAAAQSAKIkJDDlTUmI9qqQwo/sGgpos2umMb92lo9H+GGLTx0zOkGyEZ4EuOXH6ZNq6Fp246jbQygvgyBjb5XISrOiofEAAaPaCSseY8NAGHHLCUGqzK81dOwty8b1N9xdp2RuE/fKGfD2K7xuVdDscVFaP8bewySlnTZpRcDJe0IsoyrIHAYAAmgZLrkWwsq7cQtdOPYuMlBropzL+vfxOkFbmwTA/HCkS+mZD22hNM480NGrxWjoP4OUJZg+hT4AfW2pqsxk0jQABBhuOpMdDR30sgvMQe/e3kerDAx8RMTH+Y9guvw6LB+IYL0ygq2/4qAR3D1plas2TyMgiw7Dn+IXRSnHkseAjagABZCGpe30kUgyonuUQ3OEI6MXV3hA/e82M9VW64+OA7F1bfycKumW1NvKvEqlD2JCa+3vnKN13znsLy8HhJP68qQBBpf0KANNhm9TVfEcPbkYr270w5FYpUUDbJzkQ3rhD9BIQz6gSqcCDV4dHx9BdTE6TmalEpUL/hae5zlxLmK5R0wBB168R0rt0zfacknPeC+AkNa3vWHOf2lksS4QVTyyfOx2UIFFF3kCv/Y/N+27wr2u4Vz3ZInGXI+iSrSqzPiw/DcACKMcg2adD+XV7Ypi/nxIS7lSQIsNORneLDj8iGi0tsvKUEz/L/3XzttKxBSLktTKlR5S1URWJoKFe84asI1vKiAACX/0rKlZwtgtakvMtBuEUg3mcpT/RboMBrp66SDdr8mwJlrreSiHXjH6P4ql8j9cAdYyxXvwUNeaR4YkbTXcZuQBCtuy6oO+y7irnz+OrTduf4wew+8c4I4oIWC1p6GvlwVCY0neiRSG0UdlCwj+cHHlnmNp8UfUfo8R1/sBhRtB1rsAC371KYbz7fpodHkYfncGIj3v4t2NiAJjuiF1s33GnchsNEo3/6hUENAldu+sYiy+0vu+sNxD6HH8C9JUvV4hUhQBDEFMWk9B0OozGswkKfTYfCGEbHL5V60GLpTEeBDFQcjWISoepHZoD81LuaO1III3kqLQOyNwe5sf4sLvR1D7XQ8BDW2o4bXdsPESNVsAq07an01qFkH9tAMiUa5gfkwI1UbCbgW2xzYrcj2AptimZ+vOvQk9/6ddRD+++CWxU9ZNwPEBDt+0Z4P5ZCWiv+o33wiAPtdScn6fX2IZOZQUiqiE8b28BTCQqruH61Vnjg0vUurlSjaHpYB8pnDZNsK5K9HaePgAD4vwVhDccui7/UFwKuRd6UbdRhbTL6+1WKCX7rN9H+lZdgfhYFR7V0dBUPBsFOC/CST7lb/b6NXRRaWzCRdzmoAAEMm8mock0LVoboaOcavLc29hdvDWHWrWv0YVNaBhgKQeOVqqgA7lRk1RwFi/BXIjealJskrtKCZTF0xhy/AGFaMBEeRF6CXijJK9L+Baf+Rw3smiPyjEAIN0Un+LpkJM7RdtA7jdSxLYBbHQVSzrVkhDXYFIqd/qChDMqrr/tb8Wl8IBEpWRvsuBGA7wbWU5083R0XXPr0uHjdcLjy5NbrEJZVR3Omx8VgcC4hk01PNk9HfQeuSxrBBRc6w4pq0bxgLG2VoBZqEfGQAAC7oMIK61NMRcMEnTgLnZuWb5iV9Tq9QwG/r/QH+gneqK56kkAAAAAAAAm1IAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA7msoAAAAAAAAAAAAAAAAAoLhpkcYhizbB0Z1KLp6wzjYG60gAAgAAAAAAAAAAAAAAAK+Wzh4ZUrIwjrNlFYZ+BB8ySQ70ABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==\\\",\\\"indexedAt\\\":\\\"2024-07-24T15:35:01.636Z\\\",\\\"timestamp\\\":\\\"2024-07-24T15:34:49.000Z\\\",\\\"updatedAt\\\":\\\"2024-07-24T15:41:59.038Z\\\",\\\"txHash\\\":\\\"534ab4c1dd03f1f3931deb2fb23331e7354d3e9a4f20dd60cf55d2c4a22d1c52\\\",\\\"version\\\":1,\\\"revision\\\":2,\\\"overwrite\\\":false}\"}"
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
			signedVaa, err := events.GetEventData[events.SignedVaa](&notification)
			if err != nil {
				log.Error("Error decoding signedVAA from notification event", zap.String("trackId", notification.TrackID), zap.Error(err))
				return nil, nil
			}

			return &Event{
				Source:         "chain-event",
				TrackID:        notification.TrackID,
				Type:           SourceChainEvent,
				ID:             getVAADigest(ctx, signedVaa.ID, vaaRepository, log, notification),
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
				ID:             vaa.HexDigest(),
				VaaID:          vaa.MessageID(),
				ChainID:        sdk.ChainID(plm.ChainID),
				EmitterAddress: plm.Attributes.Sender,
				Sequence:       strconv.FormatUint(plm.Attributes.Sequence, 10),
				Timestamp:      &plm.BlockTime,
				IsVaaSigned:    false,
				TxHash:         plm.TxHash,
			}, nil

		case events.EvmTransactionFoundType:
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
				ID:             vaa.HexDigest(),
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

			return &Event{
				Source:         "chain-event",
				TrackID:        notification.TrackID,
				Type:           TargetChainEvent,
				ID:             getVAADigest(ctx, vaa.MessageID(), vaaRepository, log, notification),
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

func getVAADigest(ctx context.Context, vaaID string, vaaRepository vaa.VAARepository, log *zap.Logger, notification events.NotificationEvent) string {
	res, errGetVaa := vaaRepository.GetVaa(ctx, vaaID)
	if errGetVaa != nil {
		log.Error("Error getting vaa from repository", zap.String("trackId", notification.TrackID), zap.String("vaaID", vaaID), zap.Error(errGetVaa))
		res = &vaa.VaaDoc{}
	}
	return res.ID
}
