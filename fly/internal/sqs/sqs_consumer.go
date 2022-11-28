package sqs

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	aws_sqs "github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

// ConsumerOption represents a consumer option function.
type ConsumerOption func(*Consumer)

// Consumer represents SQS consumer.
type Consumer struct {
	api               sqsiface.SQSAPI
	url               string
	maxMessages       *int64
	visibilityTimeout *int64
	waitTimeSeconds   *int64
}

// New instances of a Consumer to consume SQS messages.
func NewConsumer(sess *session.Session, url string, opts ...ConsumerOption) (*Consumer, error) {
	consumer := &Consumer{
		api:               aws_sqs.New(sess),
		url:               url,
		maxMessages:       aws.Int64(10),
		visibilityTimeout: aws.Int64(60),
		waitTimeSeconds:   aws.Int64(20),
	}

	for _, opt := range opts {
		opt(consumer)
	}

	return consumer, nil
}

// WithMaxMessages allows to specify an maximum number of messages to return when setting a value.
func WithMaxMessages(v int64) ConsumerOption {
	return func(c *Consumer) {
		c.maxMessages = aws.Int64(v)
	}
}

// WithVisibilityTimeout allows to specify a visibility timeout when setting a value.
func WithVisibilityTimeout(v int64) ConsumerOption {
	return func(c *Consumer) {
		c.visibilityTimeout = aws.Int64(v)
	}
}

// WithWaitTimeSeconds allows to specify a wait time when setting a value.
func WithWaitTimeSeconds(v int64) ConsumerOption {
	return func(c *Consumer) {
		c.waitTimeSeconds = aws.Int64(v)
	}
}

// GetMessages retrieves messages from SQS.
func (c *Consumer) GetMessages() ([]*aws_sqs.Message, error) {
	params := &aws_sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(c.url),
		MaxNumberOfMessages: c.maxMessages,
		AttributeNames: []*string{
			aws.String("All"),
		},
		MessageAttributeNames: []*string{
			aws.String("All"),
		},
		WaitTimeSeconds:   c.waitTimeSeconds,
		VisibilityTimeout: c.visibilityTimeout,
	}

	res, err := c.api.ReceiveMessage(params)
	if err != nil {
		return nil, err
	}

	return res.Messages, nil
}

// DeleteMessage deletes messages from SQS.
func (c *Consumer) DeleteMessage(msg *aws_sqs.Message) error {
	params := &aws_sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.url),
		ReceiptHandle: msg.ReceiptHandle,
	}
	_, err := c.api.DeleteMessage(params)

	return err
}

// GetVisibilityTimeout returns visibility timeout.
func (c *Consumer) GetVisibilityTimeout() time.Duration {
	return time.Duration(*c.visibilityTimeout * int64(time.Second))
}

// GetQueueAttributes get queue attributes.
func (c *Consumer) GetQueueAttributes() (*aws_sqs.GetQueueAttributesOutput, error) {
	params := &aws_sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(c.url),
		AttributeNames: []*string{
			aws.String("CreatedTimestamp"),
		},
	}
	return c.api.GetQueueAttributes(params)
}
