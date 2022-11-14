package sqs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	aws_sqs "github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

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

func WithMaxMessages(v int64) ConsumerOption {
	return func(c *Consumer) {
		c.maxMessages = aws.Int64(v)
	}
}

func WithVisibilityTimeout(v int64) ConsumerOption {
	return func(c *Consumer) {
		c.visibilityTimeout = aws.Int64(v)
	}
}

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
