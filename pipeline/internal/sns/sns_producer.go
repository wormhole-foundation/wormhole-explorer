package sns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_sns "github.com/aws/aws-sdk-go-v2/service/sns"
)

// Producer represents SNS producer.
type Producer struct {
	api *aws_sns.Client
	url string
}

func NewProducer(awsConfig aws.Config, url string) (*Producer, error) {
	return &Producer{
		api: aws_sns.NewFromConfig(awsConfig),
		url: url,
	}, nil
}

// SendMessage sends messages to SQS.
func (p *Producer) SendMessage(ctx context.Context, groupID, deduplicationID, body string) error {
	_, err := p.api.Publish(ctx,
		&aws_sns.PublishInput{
			MessageGroupId:         aws.String(groupID),
			MessageDeduplicationId: aws.String(deduplicationID),
			Message:                aws.String(body),
			TopicArn:               aws.String(p.url),
		})
	return err
}
