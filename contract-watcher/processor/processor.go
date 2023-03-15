package processor

import (
	"context"

	"github.com/wormhole-foundation/wormhole-explorer/contract-watcher/watcher"
	"go.uber.org/zap"
)

type Processor struct {
	watchers []watcher.ContractWatcher
	logger   *zap.Logger
}

func NewProcessor(watchers []watcher.ContractWatcher, logger *zap.Logger) *Processor {
	return &Processor{watchers: watchers, logger: logger}
}

func (p *Processor) Start(ctx context.Context) {
	for _, watcher := range p.watchers {
		go watcher.Start(ctx)
	}
}

func (p *Processor) Close() {
	for _, watcher := range p.watchers {
		watcher.Close()
	}
}
