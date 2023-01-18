package sqs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	aws_sqs "github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

// Producer represents SQS producer.
type Producer struct {
	api sqsiface.SQSAPI
	url string
}

// NewProducer create a new instance of Producer.
func NewProducer(sess *session.Session, url string) (*Producer, error) {
	return &Producer{
		api: aws_sqs.New(sess),
		url: url,
	}, nil
}

// SendMessage sends messages to SQS.
func (p *Producer) SendMessage(groupID, deduplicationID, body string) error {
	_, err := p.api.SendMessage(
		&aws_sqs.SendMessageInput{
			MessageGroupId:         aws.String(groupID),
			MessageDeduplicationId: aws.String(deduplicationID),
			MessageBody:            aws.String(body),
			QueueUrl:               aws.String(p.url),
		})
	return err
}
