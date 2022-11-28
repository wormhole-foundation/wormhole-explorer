package server

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/sqs"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"
)

// Controller definition.
type Controller struct {
	repository *storage.Repository
	consumer   *sqs.Consumer
	isLocal    bool
}

// NewController creates a Controller instance.
func NewController(repo *storage.Repository, consumer *sqs.Consumer, isLocal bool) *Controller {
	return &Controller{repository: repo, consumer: consumer, isLocal: isLocal}
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
	mongoStatus := c.checkMongoStatus(ctx.Context())
	if !mongoStatus {
		return ctx.Status(fiber.StatusInternalServerError).JSON(struct {
			Ready string `json:"ready"`
		}{Ready: "NO"})
	}
	// check aws SQS is ready.
	queueStatus := c.checkQueueStatus(ctx.Context())
	if !queueStatus {
		return ctx.Status(fiber.StatusInternalServerError).JSON(struct {
			Ready string `json:"ready"`
		}{Ready: "NO"})
	}

	// return success response.
	return ctx.Status(fiber.StatusOK).JSON(struct {
		Ready string `json:"ready"`
	}{Ready: "OK"})
}

func (c *Controller) checkMongoStatus(ctx context.Context) bool {
	mongoStatus, err := c.repository.GetMongoStatus(ctx)
	if err != nil {
		return false
	}

	// check mongo server status
	mongoStatusCheck := (mongoStatus.Ok == 1 && mongoStatus.Pid > 0 && mongoStatus.Uptime > 0)
	if !mongoStatusCheck {
		return false
	}

	// check mongo connections
	if mongoStatus.Connections.Available <= 0 {
		return false
	}
	return true
}

func (c *Controller) checkQueueStatus(ctx context.Context) bool {
	// vaa queue handle in memory [local enviroment]
	if c.isLocal {
		return true
	}
	// get queue attributes
	queueAttributes, err := c.consumer.GetQueueAttributes()
	if err != nil || queueAttributes == nil {
		return false
	}

	// check queue created
	createdTimestamp := queueAttributes.Attributes["CreatedTimestamp"]
	if createdTimestamp == nil {
		return false
	}
	return *createdTimestamp != ""
}
