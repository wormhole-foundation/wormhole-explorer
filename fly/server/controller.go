package server

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/sqs"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"
	"go.uber.org/zap"
)

// Controller definition.
type Controller struct {
	repository *storage.Repository
	consumer   *sqs.Consumer
	isLocal    bool
	logger     *zap.Logger
}

// NewController creates a Controller instance.
func NewController(repo *storage.Repository, consumer *sqs.Consumer, isLocal bool, logger *zap.Logger) *Controller {
	return &Controller{repository: repo, consumer: consumer, isLocal: isLocal, logger: logger}
}

// HealthCheck handler for the endpoint /health.
func (c *Controller) HealthCheck(ctx *fiber.Ctx) error {
	return ctx.JSON(struct {
		Status string `json:"status"`
	}{Status: "OK"})
}

// ReadyCheck handler for the endpoint /ready
func (c *Controller) ReadyCheck(ctx *fiber.Ctx) error {
	// check mongo db is ready.
	mongoErr := c.checkMongoStatus(ctx.Context())
	if mongoErr != nil {
		c.logger.Error("Ready check failed", zap.Error(mongoErr))
		return ctx.Status(fiber.StatusInternalServerError).JSON(struct {
			Ready string `json:"ready"`
			Error string `json:"error"`
		}{Ready: "NO", Error: mongoErr.Error()})
	}
	// check aws SQS is ready.
	queueErr := c.checkQueueStatus(ctx.Context())
	if queueErr != nil {
		c.logger.Error("Ready check failed", zap.Error(queueErr))
		return ctx.Status(fiber.StatusInternalServerError).JSON(struct {
			Ready string `json:"ready"`
			Error string `json:"error"`
		}{Ready: "NO", Error: queueErr.Error()})
	}

	// return success response.
	return ctx.Status(fiber.StatusOK).JSON(struct {
		Ready string `json:"ready"`
	}{Ready: "OK"})
}

func (c *Controller) checkMongoStatus(ctx context.Context) error {
	mongoStatus, err := c.repository.GetMongoStatus(ctx)
	if err != nil {
		return err
	}

	// check mongo server status
	mongoStatusCheck := (mongoStatus.Ok == 1 && mongoStatus.Pid > 0 && mongoStatus.Uptime > 0)
	if !mongoStatusCheck {
		return errors.New("mongo invalid status")
	}

	// check mongo connections
	if mongoStatus.Connections.Available <= 0 {
		return errors.New("mongo hasn't available connections")
	}
	return nil
}

func (c *Controller) checkQueueStatus(ctx context.Context) error {
	// vaa queue handle in memory [local enviroment]
	if c.isLocal {
		return nil
	}
	// get queue attributes
	queueAttributes, err := c.consumer.GetQueueAttributes()
	if err != nil {
		return err
	}
	if queueAttributes == nil {
		return errors.New("can't get attributes for sqs")
	}

	// check queue created
	createdTimestamp := queueAttributes.Attributes["CreatedTimestamp"]
	if createdTimestamp == nil || *createdTimestamp == "" {
		return errors.New("sqs queue hasn't been created")
	}
	return nil
}
