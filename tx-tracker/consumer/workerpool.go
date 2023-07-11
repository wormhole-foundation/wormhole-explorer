package consumer

import (
	"context"
	"fmt"
	"sync"

	"github.com/wormhole-foundation/wormhole-explorer/txtracker/chains"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/config"
	"github.com/wormhole-foundation/wormhole-explorer/txtracker/queue"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
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

	// Do not process messages from PythNet
	if event.ChainID == sdk.ChainIDPythNet {
		if !msg.IsExpired() {
			w.logger.Debug("Deleting PythNet message", zap.String("vaaId", event.ID))
			msg.Done()
		} else {
			w.logger.Debug("Skipping expired PythNet message", zap.String("vaaId", event.ID))
		}
		return
	}

	// Skip non-processed, expired messages
	if msg.IsExpired() {
		w.logger.Warn("Message expired - skipping",
			zap.String("vaaId", event.ID),
			zap.Bool("isExpired", msg.IsExpired()),
		)
		return
	}

	// Process the VAA
	p := ProcessSourceTxParams{
		VaaId:     event.ID,
		ChainId:   event.ChainID,
		Emitter:   event.EmitterAddress,
		Sequence:  event.Sequence,
		TxHash:    event.TxHash,
		Overwrite: false, // avoid processing the same transaction twice
	}
	err := ProcessSourceTx(w.ctx, w.logger, w.rpcProviderSettings, w.repository, &p)

	// Log a message informing the processing status
	if err == chains.ErrChainNotSupported {
		w.logger.Info("Skipping VAA - chain not supported",
			zap.String("vaaId", event.ID),
		)
	} else if err == ErrAlreadyProcessed {
		w.logger.Warn("Message already processed - skipping",
			zap.String("vaaId", event.ID),
		)
	} else if err != nil {
		w.logger.Error("Failed to process originTx",
			zap.String("vaaId", event.ID),
			zap.Error(err),
		)
	} else {
		w.logger.Info("Updated originTx in the database",
			zap.String("id", event.ID),
		)
	}

	// Mark the message as done
	//
	// If the message is expired, it will be put back into the queue.
	if !msg.IsExpired() {
		msg.Done()
	}
}
