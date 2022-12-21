package infraestructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/wormhole-foundation/wormhole-explorer/parser/internal/sqs"
	"go.uber.org/zap"
)

// Service definition.
type Service struct {
	repo            *Repository
	consumer        *sqs.Consumer
	isQueueConsumer bool
	logger          *zap.Logger
}

// NewService create a new Service instance.
func NewService(dao *Repository, consumer *sqs.Consumer, isQueueConsumer bool, logger *zap.Logger) *Service {
	return &Service{repo: dao, consumer: consumer, isQueueConsumer: isQueueConsumer, logger: logger.With(zap.String("module", "Infraestructureervice"))}
}

// CheckIsReady check if the service is ready.
func (s *Service) CheckIsReady(ctx context.Context) (bool, error) {
	// check if mongodb is ready
	isMongoReady, err := s.CheckMongoServerStatus(ctx)
	if err != nil {
		return false, err
	}

	// check if aws sqs is ready
	isAwsSQSReady, err := s.CheckAwsSQS(ctx)
	if err != nil {
		return false, err
	}

	if !(isMongoReady && isAwsSQSReady) {
		return false, errors.New("error services not ready")
	}
	return true, nil
}

// CheckMongoServerStatus check mongodb status.
func (s *Service) CheckMongoServerStatus(ctx context.Context) (bool, error) {
	mongoStatus, err := s.repo.GetMongoStatus(ctx)
	if err != nil {
		return false, err
	}

	// check mongo server status
	mongoStatusCheck := (mongoStatus.Ok == 1 && mongoStatus.Pid > 0 && mongoStatus.Uptime > 0)
	if !mongoStatusCheck {
		return false, fmt.Errorf("mongo server not ready (Ok = %v, Pid = %v, Uptime = %v)", mongoStatus.Ok, mongoStatus.Pid, mongoStatus.Uptime)
	}

	// check mongo connections
	if mongoStatus.Connections.Available <= 0 {
		return false, fmt.Errorf("mongo server without available connections (availableConection = %v)", mongoStatus.Connections.Available)
	}
	return true, nil
}

// CheckAwsSQS check aws sqs status.
func (s *Service) CheckAwsSQS(ctx context.Context) (bool, error) {
	// vaa queue handle in memory [local enviroment]
	if !s.isQueueConsumer {
		return true, nil
	}
	// get queue attributes
	queueAttributes, err := s.consumer.GetQueueAttributes()
	if err != nil || queueAttributes == nil {
		return false, err
	}

	// check queue created
	createdTimestamp := queueAttributes.Attributes["CreatedTimestamp"]
	if createdTimestamp == nil {
		return false, errors.New("error createdTimestamp attributes does not exist")
	}
	return *createdTimestamp != "", nil
}
