package infrastructure

import (
	"context"
	"errors"
	"fmt"

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
	isInfluxReady, err := s.CheckInfluxServerStatus(ctx)
	if err != nil {
		return false, err
	}

	// check if aws sqs is ready
	isAwsSQSReady, err := s.CheckAwsSQS(ctx)
	if err != nil {
		return false, err
	}

	if !(isInfluxReady && isAwsSQSReady) {
		return false, errors.New("error services not ready")
	}
	return true, nil
}

// CheckInfluxServerStatus check influxdb server status.
func (s *Service) CheckInfluxServerStatus(ctx context.Context) (bool, error) {
	influxStatus, err := s.repo.GetInfluxStatus(ctx)
	if err != nil {
		return false, err
	}

	// check mongo server status
	influxDbStatusCheck := influxStatus.Message == "ready for queries and writes" && influxStatus.Status == "pass"
	if !influxDbStatusCheck {
		return false, fmt.Errorf("influx server not ready (Message = %s, Status = %s)", influxStatus.Message, influxStatus.Status)
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
	queueAttributes, err := s.consumer.GetQueueAttributes(ctx)
	if err != nil || queueAttributes == nil {
		return false, err
	}

	// check queue created
	createdTimestamp := queueAttributes.Attributes["CreatedTimestamp"]
	if createdTimestamp == "" {
		return false, errors.New("error createdTimestamp attributes does not exist")
	}
	return createdTimestamp != "", nil
}
