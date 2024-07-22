package sqs

import (
	"context"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_sqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	aws_sqs_types "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// ConsumerOption represents a consumer option function.
type ConsumerOption func(*Consumer)

// Consumer represents SQS consumer.
type Consumer struct {
	api               *aws_sqs.Client
	url               string
	maxMessages       int32
	visibilityTimeout int32
	waitTimeSeconds   int32
}

// New instances of a Consumer to consume SQS messages.
func NewConsumer(awsConfig aws.Config, url string, opts ...ConsumerOption) (*Consumer, error) {
	consumer := &Consumer{
		api:               aws_sqs.NewFromConfig(awsConfig),
		url:               url,
		maxMessages:       10,
		visibilityTimeout: 60,
		waitTimeSeconds:   20,
	}

	for _, opt := range opts {
		opt(consumer)
	}

	return consumer, nil
}

// WithMaxMessages allows to specify an maximum number of messages to return when setting a value.
func WithMaxMessages(v int32) ConsumerOption {
	return func(c *Consumer) {
		c.maxMessages = v
	}
}

// WithVisibilityTimeout allows to specify a visibility timeout when setting a value.
func WithVisibilityTimeout(v int32) ConsumerOption {
	return func(c *Consumer) {
		c.visibilityTimeout = v
	}
}

// WithWaitTimeSeconds allows to specify a wait time when setting a value.
func WithWaitTimeSeconds(v int32) ConsumerOption {
	return func(c *Consumer) {
		c.waitTimeSeconds = v
	}
}

// GetMessages retrieves messages from SQS.
func (c *Consumer) GetMessages(ctx context.Context) ([]aws_sqs_types.Message, error) {
	params := &aws_sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(c.url),
		MaxNumberOfMessages: c.maxMessages,
		AttributeNames: []aws_sqs_types.QueueAttributeName{
			aws_sqs_types.QueueAttributeNameAll,
		},
		MessageAttributeNames: []string{
			string(aws_sqs_types.QueueAttributeNameAll),
		},
		WaitTimeSeconds:   c.waitTimeSeconds,
		VisibilityTimeout: c.visibilityTimeout,
	}

	res, err := c.api.ReceiveMessage(ctx, params)
	if err != nil {
		return nil, err
	}

	return res.Messages, nil
}

// DeleteMessage deletes messages from SQS.
func (c *Consumer) DeleteMessage(ctx context.Context, id *string) error {
	params := &aws_sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.url),
		ReceiptHandle: id,
	}
	_, err := c.api.DeleteMessage(ctx, params)

	return err
}

// GetVisibilityTimeout returns visibility timeout.
func (c *Consumer) GetVisibilityTimeout() time.Duration {
	return time.Duration(int64(c.visibilityTimeout) * int64(time.Second))
}

// GetQueueAttributes get queue attributes.
func (c *Consumer) GetQueueAttributes(ctx context.Context) (*aws_sqs.GetQueueAttributesOutput, error) {
	params := &aws_sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(c.url),
		AttributeNames: []aws_sqs_types.QueueAttributeName{
			aws_sqs_types.QueueAttributeNameCreatedTimestamp,
		},
	}
	return c.api.GetQueueAttributes(ctx, params)
}

// GetQueueUrl returns queue url.
func (c *Consumer) GetQueueUrl() string {
	return c.url
}

func GetSentTimestamp(msg aws_sqs_types.Message) *time.Time {
	sentTimestampStr := msg.Attributes[string(aws_sqs_types.MessageSystemAttributeNameSentTimestamp)]
	if sentTimestampStr == "" {
		return nil
	}

	sentTimestampUInt, err := strconv.ParseUint(sentTimestampStr, 10, 64)
	if err != nil {
		return nil
	}
	sentTimestamp := time.Unix(0, int64(sentTimestampUInt)*int64(time.Millisecond))
	return &sentTimestamp
}
