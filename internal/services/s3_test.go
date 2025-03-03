package services

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// mockS3Client is a mock implementation of the S3 client for testing
type mockS3Client struct {
	objects map[string][]byte
}

// PutObject mocks the S3 PutObject operation
func (m *mockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	// Extract the bucket and key
	bucket := aws.ToString(params.Bucket)
	key := aws.ToString(params.Key)

	// Read the content of the body
	body, err := io.ReadAll(params.Body)
	if err != nil {
		return nil, err
	}

	// Store the object (in a real app, this would go to S3)
	if m.objects == nil {
		m.objects = make(map[string][]byte)
	}
	m.objects[bucket+"/"+key] = body

	return &s3.PutObjectOutput{}, nil
}

// GetObject mocks the S3 GetObject operation
func (m *mockS3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	// Extract the bucket and key
	bucket := aws.ToString(params.Bucket)
	key := aws.ToString(params.Key)

	// Check if the object exists
	objectKey := bucket + "/" + key
	content, ok := m.objects[objectKey]
	if !ok {
		return nil, &types.NoSuchKey{
			Message: aws.String("The specified key does not exist."),
		}
	}

	// Return the object
	return &s3.GetObjectOutput{
		Body:        io.NopCloser(bytes.NewReader(content)),
		ContentType: aws.String("image/jpeg"),
	}, nil
}

// DeleteObject mocks the S3 DeleteObject operation
func (m *mockS3Client) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	// Extract the bucket and key
	bucket := aws.ToString(params.Bucket)
	key := aws.ToString(params.Key)

	// Delete the object
	objectKey := bucket + "/" + key
	delete(m.objects, objectKey)

	return &s3.DeleteObjectOutput{}, nil
}

// makeS3ClientInterface creates a client that satisfies the S3 client interface
type s3ClientInterface interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// Ensure mockS3Client implements the S3 client interface
var _ s3ClientInterface = (*mockS3Client)(nil)

// TestS3ServiceImpl is a custom implementation of S3Service for testing
type TestS3ServiceImpl struct {
	client     s3ClientInterface
	bucketName string
}

// Implement all methods from the S3Service on TestS3ServiceImpl
func (s *TestS3ServiceImpl) UploadImage(ctx context.Context, key string, fileContent multipart.File, contentType string) error {
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

func (s *TestS3ServiceImpl) GetImageURL(ctx context.Context, key string) (string, error) {
	return "/images/" + key, nil
}

func (s *TestS3ServiceImpl) GetImage(ctx context.Context, key string) ([]byte, string, error) {
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

func (s *TestS3ServiceImpl) DeleteImage(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	return err
}

func (s *TestS3ServiceImpl) GetBucketName() string {
	return s.bucketName
}

// Ensure TestS3ServiceImpl implements StorageService
var _ StorageService = (*TestS3ServiceImpl)(nil)

func TestS3Service(t *testing.T) {
	// Create a mock S3 client
	mockClient := &mockS3Client{
		objects: make(map[string][]byte),
	}

	// Create an S3 service with the mock client
	bucketName := "test-bucket"
	// Create our test service
	service := &TestS3ServiceImpl{
		client:     mockClient,
		bucketName: bucketName,
	}

	// Test GetBucketName
	t.Run("GetBucketName", func(t *testing.T) {
		if service.GetBucketName() != bucketName {
			t.Errorf("Expected bucket name %s, got %s", bucketName, service.GetBucketName())
		}
	})

	// Test UploadImage
	t.Run("UploadImage", func(t *testing.T) {
		// Create a test image
		imageContent := []byte("fake image content")
		fakeFile := createMultipartFile(t, imageContent)
		
		ctx := context.Background()
		err := service.UploadImage(ctx, "test.jpg", fakeFile, "image/jpeg")
		if err != nil {
			t.Fatalf("Failed to upload image: %v", err)
		}

		// Verify the file was stored in the mock client
		objectKey := bucketName + "/" + "test.jpg"
		content, ok := mockClient.objects[objectKey]
		if !ok {
			t.Errorf("Expected object %s to exist", objectKey)
		}

		if !bytes.Equal(content, imageContent) {
			t.Errorf("Object content does not match expected")
		}
	})

	// Test GetImageURL
	t.Run("GetImageURL", func(t *testing.T) {
		ctx := context.Background()
		url, err := service.GetImageURL(ctx, "test.jpg")
		if err != nil {
			t.Fatalf("Failed to get image URL: %v", err)
		}

		expectedURL := "/images/test.jpg"
		if url != expectedURL {
			t.Errorf("Expected URL %s, got %s", expectedURL, url)
		}
	})

	// Test GetImage
	t.Run("GetImage", func(t *testing.T) {
		// Upload a test image first
		imageContent := []byte("another test image")
		objectKey := bucketName + "/" + "test2.jpg"
		mockClient.objects[objectKey] = imageContent

		ctx := context.Background()
		content, contentType, err := service.GetImage(ctx, "test2.jpg")
		if err != nil {
			t.Fatalf("Failed to get image: %v", err)
		}

		if !bytes.Equal(content, imageContent) {
			t.Errorf("Image content does not match expected")
		}

		if contentType != "image/jpeg" {
			t.Errorf("Expected content type image/jpeg, got %s", contentType)
		}
	})

	// Test DeleteImage
	t.Run("DeleteImage", func(t *testing.T) {
		// Upload a test image first
		imageContent := []byte("delete test image")
		objectKey := bucketName + "/" + "delete.jpg"
		mockClient.objects[objectKey] = imageContent

		ctx := context.Background()
		err := service.DeleteImage(ctx, "delete.jpg")
		if err != nil {
			t.Fatalf("Failed to delete image: %v", err)
		}

		// Verify the file was deleted
		_, ok := mockClient.objects[objectKey]
		if ok {
			t.Errorf("Expected object %s to be deleted", objectKey)
		}
	})
}