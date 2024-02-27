package s3

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Repository struct {
	uploader *manager.Uploader
	bucket   string
}

func NewS3Repository(awsConfig aws.Config, bucket string) *S3Repository {
	client := s3.NewFromConfig(awsConfig)

	return &S3Repository{
		uploader: manager.NewUploader(client),
		bucket:   bucket,
	}
}

func (r *S3Repository) Save(ctx context.Context, key string, body []byte) error {
	_, err := r.uploader.Upload(ctx,
		&s3.PutObjectInput{
			Bucket: &r.bucket,
			Key:    &key,
			Body:   bytes.NewReader(body),
		})
	return err
}
