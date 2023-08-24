package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/sns"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly/topic"
	"go.uber.org/zap"
)

// newAwsConfig creates a new AWS config from the given configuration.
func newAwsConfig(ctx context.Context, cfg WorkerConfiguration) (aws.Config, error) {
	region := cfg.AwsRegion
	if region == "" {
		return aws.Config{}, fmt.Errorf("AWS_REGION is required")
	}
	awsSecretId := cfg.AwsAccessKeyId
	awsSecretKey := cfg.AwsSecretKey
	if awsSecretId != "" && awsSecretKey != "" {
		credentials := credentials.NewStaticCredentialsProvider(awsSecretId, awsSecretKey, "")
		customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			if cfg.AwsEndpoint != "" {
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           cfg.AwsEndpoint,
					SigningRegion: region,
				}, nil
			}

			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})

		awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(region),
			awsconfig.WithEndpointResolver(customResolver),
			awsconfig.WithCredentialsProvider(credentials),
		)
		return awsCfg, err
	}

	return awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
}

// newSNSProducer creates a new SNS producer from the given configuration.
func newSNSProducer(ctx context.Context, cfg WorkerConfiguration, alertClient alert.AlertClient, metricsClient metrics.Metrics, logger *zap.Logger) (*topic.SNSProducer, error) {
	if cfg.AwsSnsURL == "" {
		return nil, fmt.Errorf("AWS_SNS_URL is required")
	}

	awsConfig, err := newAwsConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	snsProducer, err := sns.NewProducer(awsConfig, cfg.AwsSnsURL)
	if err != nil {
		return nil, err
	}

	return topic.NewSNSProducer(snsProducer, alertClient, metricsClient, logger), nil
}

// newVAATopicProducerFunc creates a new VAA topic producer function from the given configuration.
func newVAATopicProducerFunc(ctx context.Context, cfg WorkerConfiguration, alertClient alert.AlertClient, metricsClient metrics.Metrics, logger *zap.Logger) (topic.PushFunc, error) {
	if !cfg.NotifyEnabled {
		return func(context.Context, *topic.NotificationEvent) error {
			return nil
		}, nil
	}

	snsProducer, err := newSNSProducer(ctx, cfg, alertClient, metricsClient, logger)
	if err != nil {
		logger.Fatal("could not create vaa topic producer ", zap.Error(err))
		return nil, err
	}

	return snsProducer.Push, nil
}
