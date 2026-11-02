package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Client struct {
	S3            *s3.Client
	Bucket        string
	PublicBaseURL string
}

func NewR2Client(ctx context.Context) (*R2Client, error) {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	bucket := os.Getenv("R2_BUCKET_NAME")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	endpoint := os.Getenv("R2_ENDPOINT")
	publicUrl := os.Getenv("PUBLIC_URL")

	// Validate configuration
	if accountID == "" ||
		bucket == "" ||
		accessKey == "" ||
		secretKey == "" ||
		endpoint == "" {
		return nil, fmt.Errorf("missing required R2 environment variables")
	}

	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("auto"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				accessKey,
				secretKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS SDK config: %w", err)
	}

	// Create S3 client configured for Cloudflare R2
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &R2Client{
		S3:            s3Client,
		Bucket:        bucket,
		PublicBaseURL: publicUrl,
	}, nil
}

// Helper function to check if we have connected to bucket or not
func (r *R2Client) CheckBucket(ctx context.Context) error {
	_, err := r.S3.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(r.Bucket),
	})

	if err != nil {
		return fmt.Errorf("failed to access R2 bucket: %w", err)
	}

	return nil
}

func (r *R2Client) UploadObject(ctx context.Context, key string, body io.Reader, contentLength int64, contentType string) error {

	_, err := r.S3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(r.Bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(contentLength),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("failed to upload object to R2: %w", err)
	}

	return nil
}

func (r *R2Client) Delete(ctx context.Context, keys ...string) error {
	var firstErr error
	for _, key := range keys {
		_, err := r.S3.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(r.Bucket),
			Key:    aws.String(key),
		})
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (r *R2Client) GetObjectURL(
	ctx context.Context,
	key string,
	visibility string,
) (string, error) {

	if visibility == "PUBLIC" {
		print("Base URL", r.PublicBaseURL)
		return strings.TrimRight(r.PublicBaseURL, "/") +
			"/" + strings.TrimLeft(key, "/"), nil
	}

	presignClient := s3.NewPresignClient(r.S3)

	result, err := presignClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(r.Bucket),
			Key:    aws.String(key),
		},
		s3.WithPresignExpires(15*time.Minute),
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return result.URL, nil
}
