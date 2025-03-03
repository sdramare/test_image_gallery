package services

import (
	"context"
	"mime/multipart"
)

// StorageService defines the common interface for storage services
type StorageService interface {
	// UploadImage uploads an image to storage
	UploadImage(ctx context.Context, key string, fileContent multipart.File, contentType string) error
	
	// GetImageURL generates a URL to access the image
	GetImageURL(ctx context.Context, key string) (string, error)
	
	// DeleteImage removes an image from storage
	DeleteImage(ctx context.Context, key string) error
	
	// GetImage gets an image from storage
	GetImage(ctx context.Context, key string) ([]byte, string, error)
	
	// GetBucketName returns the bucket name or storage path
	GetBucketName() string
}