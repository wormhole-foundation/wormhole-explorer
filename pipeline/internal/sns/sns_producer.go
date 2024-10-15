package sns

import (
	"context"
	"fmt"
	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_sns "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
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
func (p *Producer) SendMessage(ctx context.Context, chainId sdk.ChainID, groupID, deduplicationID, body string) error {
	attrs := map[string]types.MessageAttributeValue{
		"chainId": {
			DataType:    aws.String("String"),
			StringValue: aws.String(fmt.Sprintf("%d", chainId)),
		},
	}
	_, err := p.api.Publish(ctx,
		&aws_sns.PublishInput{
			MessageGroupId:         aws.String(groupID),
			MessageDeduplicationId: aws.String(deduplicationID),
			Message:                aws.String(body),
			TopicArn:               aws.String(p.url),
			MessageAttributes:      attrs,
		})
	return err
}
