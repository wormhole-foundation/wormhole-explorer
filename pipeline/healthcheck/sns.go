package healthcheck

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
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
