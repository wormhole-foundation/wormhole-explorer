package builder

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/wormhole-foundation/wormhole-explorer/fly/config"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/sqs"
)

func NewAwsConfig(ctx context.Context, config *config.Configuration) (aws.Config, error) {
	if config.IsLocal {
		return *aws.NewConfig(), nil
	}
	awsSecretId := config.Aws.AwsAccessKeyID
	awsSecretKey := config.Aws.AwsSecretAccessKey
	if awsSecretId != "" && awsSecretKey != "" {
		credentials := credentials.NewStaticCredentialsProvider(awsSecretId, awsSecretKey, "")
		customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			awsEndpoint := config.Aws.AwsEndpoint
			if awsEndpoint != "" {
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           awsEndpoint,
					SigningRegion: region,
				}, nil
			}

			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})

		awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(config.Aws.AwsRegion),
			awsconfig.WithEndpointResolver(customResolver),
			awsconfig.WithCredentialsProvider(credentials),
		)
		return awsCfg, err
	}

	return awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(config.Aws.AwsRegion))
}

func NewSQSProducer(awsConfig aws.Config, sqsURL string) (*sqs.Producer, error) {

	return sqs.NewProducer(awsConfig, sqsURL)
}

func NewSQSConsumer(sqsURL string, ctx context.Context, cfg *config.Configuration) (*sqs.Consumer, error) {

	awsConfig, err := NewAwsConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return sqs.NewConsumer(awsConfig, sqsURL,
		sqs.WithMaxMessages(10),
		sqs.WithVisibilityTimeout(120))
}
