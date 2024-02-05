package queue

import (
	"context"
	"sync"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/sqs"
	"go.uber.org/zap"
)

type sqsConsumerMessage[T any] struct {
	data      T
	consumer  *sqs.Consumer
	id        *string
	logger    *zap.Logger
	expiredAt time.Time
	wg        *sync.WaitGroup
	ctx       context.Context
}

func (m *sqsConsumerMessage[T]) Data() T {
	return m.data
}

func (m *sqsConsumerMessage[T]) Done(ctx context.Context) {
	if err := m.consumer.DeleteMessage(ctx, m.id); err != nil {
		m.logger.Error("Error deleting message from SQS", zap.Error(err))
	}
	m.wg.Done()
}

func (m *sqsConsumerMessage[T]) Failed() {
	m.wg.Done()
}

func (m *sqsConsumerMessage[T]) IsExpired() bool {
	return m.expiredAt.Before(time.Now())
}

type memoryConsumerMessageQueue[T any] struct {
	data T
}

func (m *memoryConsumerMessageQueue[T]) Data() T {
	return m.data
}

func (m *memoryConsumerMessageQueue[T]) Done(_ context.Context) {}

func (m *memoryConsumerMessageQueue[T]) Failed() {}

func (m *memoryConsumerMessageQueue[T]) IsExpired() bool {
	return false
}
