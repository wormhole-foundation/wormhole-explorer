package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/pipeline/topic"
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
	defaultRetries int
}

func NewTxHashHandler(repository IRepository, pushFunc topic.PushFunc, logger *zap.Logger, quit chan bool) *TxHashHandler {
	return &TxHashHandler{
		logger:         logger,
		repository:     repository,
		fixItems:       map[string]ItemTuple{},
		inputQueue:     make(chan topic.Event, 100),
		sleepTime:      2 * time.Second,
		pushFunc:       pushFunc,
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
				Retries: 3,
				Event:   event,
			}
		default:
			// no lock needed. the map is never updated while iterating.
			for vaa, item := range t.fixItems {
				if item.Retries > 0 {
					txHash, err := t.handleEmptyVaaTxHash(ctx, vaa)
					if err != nil {
						t.logger.Error("Error while trying to fix vaa txhash", zap.Int("retries_count", item.Retries), zap.Error(err))
						item.Retries = item.Retries - 1
						t.fixItems[vaa] = item
					} else {
						t.logger.Info("Vaa txhash fixed", zap.String("vaaID", vaa), zap.String("txHash", txHash))
						item.Event.TxHash = txHash
						t.pushFunc(ctx, &item.Event)
						delete(t.fixItems, vaa)

					}
				} else {
					t.logger.Error("Vaa txhash fix failed", zap.String("vaaID", vaa))
					// publish the event to the topic anyway
					t.pushFunc(ctx, &item.Event)
					delete(t.fixItems, vaa)
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
		return "", err
	}
	return vaaIdTxHash.TxHash, nil
}
