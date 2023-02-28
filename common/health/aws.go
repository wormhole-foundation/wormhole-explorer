package health

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	aws_sqs_types "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func SNS(config aws.Config, url string) Check {
	api := sns.NewFromConfig(config)
	return func(ctx context.Context) error {
		params := &sns.GetTopicAttributesInput{
			TopicArn: aws.String(url),
		}
		_, err := api.GetTopicAttributes(ctx, params)
		if err != nil {
			return err
		}
		return nil
	}
}

func SQS(config aws.Config, url string) Check {
	api := sqs.NewFromConfig(config)
	return func(ctx context.Context) error {
		params := &sqs.GetQueueAttributesInput{
			QueueUrl: aws.String(url),
			AttributeNames: []aws_sqs_types.QueueAttributeName{
				aws_sqs_types.QueueAttributeNameCreatedTimestamp,
			},
		}
		queueAttributes, err := api.GetQueueAttributes(ctx, params)
		if err != nil {
			return err
		}
		if queueAttributes == nil {
			return errors.New("queue attributes can not be empty")
		}
		createdTimestamp := queueAttributes.Attributes["CreatedTimestamp"]
		if createdTimestamp == "" {
			return errors.New("queue attribute [createdTimestamp] does not exist")
		}
		return nil
	}
}
