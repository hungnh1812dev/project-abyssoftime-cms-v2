package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"

	"project-abyssoftime-cms-v2/api/internal/domain/repository"
)

// api is the subset of *s3.Client this adapter needs — small enough to
// fake directly in unit tests without hitting AWS.
type api interface {
	PutObject(ctx context.Context, params *awss3.PutObjectInput, optFns ...func(*awss3.Options)) (*awss3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *awss3.DeleteObjectInput, optFns ...func(*awss3.Options)) (*awss3.DeleteObjectOutput, error)
}

var _ repository.StorageAdapter = (*adapter)(nil)

type adapter struct {
	client api
	bucket string
	region string
}

// New builds an S3-backed StorageAdapter. Credentials are resolved by the
// AWS SDK's default chain (env vars, shared config, instance role, etc.)
// — never hard-coded.
func New(ctx context.Context, bucket, region string) (repository.StorageAdapter, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return &adapter{client: awss3.NewFromConfig(cfg), bucket: bucket, region: region}, nil
}

func (a *adapter) Upload(ctx context.Context, file io.Reader, filename string) (*repository.UploadResult, error) {
	_, err := a.client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		return nil, err
	}
	return &repository.UploadResult{
		URL:      fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", a.bucket, a.region, filename),
		PublicID: filename,
	}, nil
}

func (a *adapter) Delete(ctx context.Context, publicID string) error {
	_, err := a.client.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(publicID),
	})
	return err
}
