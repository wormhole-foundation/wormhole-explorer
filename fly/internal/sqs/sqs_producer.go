package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_sqs "github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Producer struct {
	api *aws_sqs.Client
	url string
}

// New instances of a client to connect SQS.
func NewProducer(cfg aws.Config, url string) (*Producer, error) {
	return &Producer{
		api: aws_sqs.NewFromConfig(cfg),
		url: url,
	}, nil
}

// SendMessage sends messages to SQS.
func (p *Producer) SendMessage(ctx context.Context, groupID, deduplicationID, body string) error {
	_, err := p.api.SendMessage(
		ctx,
		&aws_sqs.SendMessageInput{
			MessageGroupId:         aws.String(groupID),
			MessageDeduplicationId: aws.String(deduplicationID),
			MessageBody:            aws.String(body),
			QueueUrl:               aws.String(p.url),
		})
	return err
}
