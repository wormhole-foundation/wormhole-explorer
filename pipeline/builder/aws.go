package builder

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/internal/sns"
	"github.com/wormhole-foundation/wormhole-explorer/pipeline/topic"
	"go.uber.org/zap"
)

func NewAwsConfig(appCtx context.Context, region string, accessKeyID, secretAccessKey, endpoint string) (aws.Config, error) {

	if accessKeyID != "" && secretAccessKey != "" {
		credentials := credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")
		customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			if endpoint != "" {
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           endpoint,
					SigningRegion: region,
				}, nil
			}

			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})

		awsCfg, err := awsconfig.LoadDefaultConfig(appCtx,
			awsconfig.WithRegion(region),
			awsconfig.WithEndpointResolver(customResolver),
			awsconfig.WithCredentialsProvider(credentials),
		)
		return awsCfg, err
	}

	return awsconfig.LoadDefaultConfig(appCtx, awsconfig.WithRegion(region))
}

func NewTopicProducer(ctx context.Context, region, snsUrl string, accessKeyID, secretAccessKey, endpoint string,
	alertClient alert.AlertClient, metrics metrics.Metrics, logger *zap.Logger) (topic.PushFunc, error) {
	awsConfig, err := NewAwsConfig(ctx, region, accessKeyID, secretAccessKey, endpoint)
	if err != nil {
		return nil, err
	}

	snsProducer, err := sns.NewProducer(awsConfig, snsUrl)
	if err != nil {
		return nil, err
	}

	return topic.NewVAASNS(snsProducer, alertClient, metrics, logger).Publish, nil
}
