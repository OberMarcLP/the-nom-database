package services

import (
	"testing"
)

func TestObjectURL(t *testing.T) {
	tests := []struct {
		name     string
		service  *S3Service
		key      string
		expected string
	}{
		{
			name:     "AWSVirtualHostedStyle",
			service:  &S3Service{bucketName: "photos", region: "eu-central-1"},
			key:      "menu_photos/abc.jpg",
			expected: "https://photos.s3.eu-central-1.amazonaws.com/menu_photos/abc.jpg",
		},
		{
			name:     "CustomEndpointPathStyle",
			service:  &S3Service{bucketName: "photos", endpoint: "https://minio.internal:9000"},
			key:      "menu_photos/abc.jpg",
			expected: "https://minio.internal:9000/photos/menu_photos/abc.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.service.objectURL(tt.key); got != tt.expected {
				t.Errorf("objectURL() = %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestInitS3_CustomEndpoint(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "test-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test-secret")
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("S3_BUCKET_NAME", "test-bucket")
	t.Setenv("S3_ENDPOINT", "https://minio.internal:9000/")

	prev := s3Service
	t.Cleanup(func() { s3Service = prev })

	if err := InitS3(); err != nil {
		t.Fatalf("InitS3() returned error: %v", err)
	}

	// Trailing slash must be trimmed so objectURL doesn't produce double slashes
	if s3Service.endpoint != "https://minio.internal:9000" {
		t.Errorf("expected trimmed endpoint, got %q", s3Service.endpoint)
	}
}
