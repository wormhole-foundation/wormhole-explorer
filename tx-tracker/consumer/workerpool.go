package consumer

import (
	"context"
	"fmt"
	"sync"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/queue"
	"go.uber.org/zap"
)

const numWorkers = 500

// WorkerPool is an abstraction to process VAAs concurrently.
type WorkerPool struct {
	wg                  sync.WaitGroup
	chInput             chan queue.ConsumerMessage
	ctx                 context.Context
	logger              *zap.Logger
	rpcProviderSettings *config.RpcProviderSettings
	repository          *Repository
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(
	ctx context.Context,
	logger *zap.Logger,
	rpcProviderSettings *config.RpcProviderSettings,
	repository *Repository,
) *WorkerPool {

	w := WorkerPool{
		chInput:             make(chan queue.ConsumerMessage),
		ctx:                 ctx,
		logger:              logger,
		rpcProviderSettings: rpcProviderSettings,
		repository:          repository,
	}

	// Spawn worker goroutines
	for i := 0; i < numWorkers; i++ {
		w.wg.Add(1)
		go w.consumerLoop()
	}

	return &w
}

// Push sends a new item to the worker pool.
//
// This function will block until either a worker is available or the context is cancelled.
func (w *WorkerPool) Push(ctx context.Context, msg queue.ConsumerMessage) error {

	select {
	case w.chInput <- msg:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("failed to push message into worker pool due to calcelled context: %w", ctx.Err())
	}
}

// StopGracefully stops the worker pool gracefully.
//
// This function blocks until the consumer queue is empty.
func (w *WorkerPool) StopGracefully() {

	// Close the producer channel.
	// This will stop sending items to the workers.
	// After all items are consumed, the workers will exit.
	close(w.chInput)
	w.chInput = nil

	// Wait for all workers to finish gracefully
	w.wg.Wait()
}

// consumerLoop is the main loop of a worker.
//
// It will consume items from the input channel until the channel is closed or the context is cancelled.
func (w *WorkerPool) consumerLoop() {
	for {
		select {
		case msg, ok := <-w.chInput:
			if !ok {
				w.wg.Done()
				return
			}
			w.process(msg)

		case <-w.ctx.Done():
			w.wg.Done()
			return
		}
	}
}

// process consumes a single item from the input channel.
func (w *WorkerPool) process(msg queue.ConsumerMessage) {

	event := msg.Data()

	p := ProcessSourceTxParams{
		VaaId:    event.ID,
		ChainId:  event.ChainID,
		Emitter:  event.EmitterAddress,
		Sequence: event.Sequence,
		TxHash:   event.TxHash,
	}
	err := ProcessSourceTx(w.ctx, w.logger, w.rpcProviderSettings, w.repository, &p)

	if err == chains.ErrChainNotSupported {
		w.logger.Debug("Skipping VAA - chain not supported",
			zap.String("vaaId", event.ID),
		)
	} else if err != nil {
		w.logger.Error("Failed to upsert source transaction details",
			zap.String("vaaId", event.ID),
			zap.Error(err),
		)
	} else {
		w.logger.Info("Updated source transaction details in the database",
			zap.String("id", event.ID),
		)
	}

	msg.Done()
}
