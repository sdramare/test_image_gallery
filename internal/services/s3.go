package services

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Service handles operations with AWS S3
type S3Service struct {
	client     *s3.Client
	bucketName string
}

// NewS3Service creates a new S3 service
func NewS3Service(client *s3.Client, bucketName string) *S3Service {
	return &S3Service{
		client:     client,
		bucketName: bucketName,
	}
}

// GetBucketName returns the S3 bucket name
func (s *S3Service) GetBucketName() string {
	return s.bucketName
}

// Verify that S3Service implements StorageService
var _ StorageService = (*S3Service)(nil)

// UploadImage uploads an image to S3
func (s *S3Service) UploadImage(ctx context.Context, key string, fileContent multipart.File, contentType string) error {
	// Read file content
	buffer := new(bytes.Buffer)
	if _, err := io.Copy(buffer, fileContent); err != nil {
		return err
	}

	// Upload to S3
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buffer.Bytes()),
		ContentType: aws.String(contentType),
	})

	return err
}

// GetImageURL generates a presigned URL to access the image
func (s *S3Service) GetImageURL(ctx context.Context, key string) (string, error) {
	// For this example, we'll just return a direct path
	// In a real application, you would generate a presigned URL using AWS SDK
	return "/images/" + key, nil
}

// DeleteImage removes an image from S3
func (s *S3Service) DeleteImage(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	return err
}

// GetImage gets an image from S3
func (s *S3Service) GetImage(ctx context.Context, key string) ([]byte, string, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", err
	}
	defer result.Body.Close()

	// Read the content
	content, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, "", err
	}

	// Get the content type, default to "image/jpeg" if not specified
	contentType := "image/jpeg"
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	return content, contentType, nil
}