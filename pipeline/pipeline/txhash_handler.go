package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	pipelineAlert "github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/alert"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/topic"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.uber.org/zap"
)

type ItemTuple struct {
	Retries int
	Event   topic.Event
}

type TxHashHandler struct {
	logger         *zap.Logger
	repository     IRepository
	fixItems       map[string]ItemTuple
	inputQueue     chan topic.Event
	quit           chan bool
	sleepTime      time.Duration
	pushFunc       topic.PushFunc
	alertClient    alert.AlertClient
	metrics        metrics.Metrics
	defaultRetries int
}

// NewTxHashHandler creates a new TxHashHandler.
func NewTxHashHandler(repository IRepository, pushFunc topic.PushFunc, alertClient alert.AlertClient, metrics metrics.Metrics, logger *zap.Logger, quit chan bool) *TxHashHandler {
	return &TxHashHandler{
		logger:         logger,
		repository:     repository,
		fixItems:       map[string]ItemTuple{},
		inputQueue:     make(chan topic.Event, 100),
		sleepTime:      2 * time.Second,
		pushFunc:       pushFunc,
		alertClient:    alertClient,
		metrics:        metrics,
		defaultRetries: 3,
	}
}

// Add a new element to the fixItems array
func (t *TxHashHandler) AddVaaFixItem(event topic.Event) {
	t.inputQueue <- event
}

func (t *TxHashHandler) Run(ctx context.Context) {
	t.logger.Info("TxHashHandler started")
	for {
		select {
		case <-t.quit:
			t.logger.Info("stopping txhash handler")
			return
		case event := <-t.inputQueue:
			t.fixItems[event.ID] = ItemTuple{
				Retries: 5,
				Event:   event,
			}
		default:
			// no lock needed. the map is never updated while iterating.
			for vaaID, item := range t.fixItems {
				if item.Retries > 0 {
					vaa, err := sdk.Unmarshal(item.Event.Vaa)
					if err != nil {
						t.logger.Error("Error unmarshalling vaa", zap.Error(err), zap.String("vaaId", vaaID))
						delete(t.fixItems, vaaID)
						continue
					}
					uniqueVaaID := domain.CreateUniqueVaaID(vaa)
					txHash, err := t.handleEmptyVaaTxHash(ctx, uniqueVaaID)
					if err != nil {
						t.logger.Error("Error while trying to fix vaa txhash", zap.Int("retries_count", item.Retries), zap.Error(err))
						item.Retries = item.Retries - 1
						t.fixItems[vaaID] = item
					} else {
						t.logger.Info("Vaa txhash fixed", zap.String("vaaID", vaaID), zap.String("txHash", txHash))
						item.Event.TxHash = txHash
						t.pushFunc(ctx, &item.Event)
						delete(t.fixItems, vaaID)
						// increment metrics vaa with txhash fixed
						t.metrics.IncVaaWithTxHashFixed(uint16(item.Event.ChainID))
					}
				} else {
					t.logger.Error("Vaa txhash fix failed", zap.String("vaaID", vaaID))
					// publish the event to the topic anyway
					t.pushFunc(ctx, &item.Event)
					delete(t.fixItems, vaaID)
				}
			}
		}
		time.Sleep(t.sleepTime)
	}
}

// handleEmptyVaaTxHash tries to get the txhash for the vaa with the given id.
func (p *TxHashHandler) handleEmptyVaaTxHash(ctx context.Context, id string) (string, error) {
	vaaIdTxHash, err := p.repository.GetVaaIdTxHash(ctx, id)
	if err != nil {
		return "", err
	}

	if vaaIdTxHash.TxHash == "" {
		return "", fmt.Errorf("txhash for vaa (%s) is empty", id)
	}

	err = p.repository.UpdateVaaDocTxHash(ctx, id, vaaIdTxHash.TxHash)
	if err != nil {
		// Alert error updating vaa txhash.
		alertContext := alert.AlertContext{
			Details: map[string]string{
				"vaaID":  id,
				"txHash": vaaIdTxHash.TxHash,
			},
			Error: err,
		}
		p.alertClient.CreateAndSend(ctx, pipelineAlert.ErrorUpdateVaaTxHash, alertContext)
		return "", err
	}
	return vaaIdTxHash.TxHash, nil
}
