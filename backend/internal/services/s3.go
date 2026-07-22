package services

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
	"github.com/nomdb/backend/internal/logger"
)

type S3Service struct {
	client     *s3.Client
	bucketName string
	region     string
	endpoint   string
}

var s3Service *S3Service

// InitS3 initializes the S3 service with credentials
func InitS3() error {
	logger.Info("☁️  Initializing AWS S3 service...")

	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsRegion := os.Getenv("AWS_REGION")
	bucketName := os.Getenv("S3_BUCKET_NAME")
	s3Endpoint := os.Getenv("S3_ENDPOINT")

	if awsAccessKey == "" || awsSecretKey == "" || awsRegion == "" || bucketName == "" {
		logger.Warn("⚠️  AWS S3 not configured - using local storage for photos")
		return fmt.Errorf("AWS credentials or bucket name not configured. Set AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_REGION, and S3_BUCKET_NAME")
	}

	logger.Debug("Loading AWS config for region: %s, bucket: %s", awsRegion, bucketName)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			awsAccessKey,
			awsSecretKey,
			"",
		)),
	)
	if err != nil {
		logger.Error("❌ Failed to load AWS config: %v", err)
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Service = &S3Service{
		client: s3.NewFromConfig(cfg, func(o *s3.Options) {
			if s3Endpoint != "" {
				// S3-compatible storage (MinIO, R2, Hetzner, ...): path-style
				// addressing avoids per-bucket DNS on custom endpoints
				o.BaseEndpoint = aws.String(s3Endpoint)
				o.UsePathStyle = true
			}
		}),
		bucketName: bucketName,
		region:     awsRegion,
		endpoint:   strings.TrimRight(s3Endpoint, "/"),
	}

	if s3Endpoint != "" {
		logger.Info("✅ S3 service initialized (endpoint: %s, bucket: %s, region: %s)", s3Endpoint, bucketName, awsRegion)
	} else {
		logger.Info("✅ AWS S3 service initialized (bucket: %s, region: %s)", bucketName, awsRegion)
	}
	return nil
}

// GetS3Service returns the initialized S3 service
func GetS3Service() *S3Service {
	return s3Service
}

// UploadFile uploads a file to S3 and returns the URL
func (s *S3Service) UploadFile(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	logger.Debug("📤 Uploading file to S3: %s (type: %s)", key, contentType)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
		ACL:         "private", // Use private ACL for security
	})
	if err != nil {
		logger.Error("❌ Failed to upload file to S3: %v", err)
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Return the S3 URL
	url := s.objectURL(key)
	logger.Info("✅ File uploaded to S3: %s", key)
	return url, nil
}

// objectURL builds the object URL: path-style on a custom endpoint,
// otherwise the standard AWS virtual-hosted-style URL
func (s *S3Service) objectURL(key string) string {
	if s.endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucketName, key)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucketName, s.region, key)
}

// DeleteFile deletes a file from S3
func (s *S3Service) DeleteFile(ctx context.Context, key string) error {
	logger.Debug("🗑️  Deleting file from S3: %s", key)

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		logger.Error("❌ Failed to delete file from S3: %v", err)
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	logger.Info("✅ File deleted from S3: %s", key)
	return nil
}

// GetPresignedURL generates a presigned URL for private file access
func (s *S3Service) GetPresignedURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

// IsConfigured returns true if S3 is properly configured
func IsS3Configured() bool {
	return os.Getenv("AWS_ACCESS_KEY_ID") != "" &&
		os.Getenv("AWS_SECRET_ACCESS_KEY") != "" &&
		os.Getenv("AWS_REGION") != "" &&
		os.Getenv("S3_BUCKET_NAME") != ""
}
