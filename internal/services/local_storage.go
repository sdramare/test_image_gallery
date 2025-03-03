package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// LocalStorageService is a local implementation of S3Service for development
type LocalStorageService struct {
	storagePath string
}

// NewLocalStorageService creates a new local storage service
func NewLocalStorageService(storagePath string) (*LocalStorageService, error) {
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorageService{
		storagePath: storagePath,
	}, nil
}

// GetBucketName returns the storage path
func (s *LocalStorageService) GetBucketName() string {
	return s.storagePath
}

// Verify that LocalStorageService implements StorageService
var _ StorageService = (*LocalStorageService)(nil)

// UploadImage saves an image to local storage
func (s *LocalStorageService) UploadImage(ctx context.Context, key string, fileContent multipart.File, contentType string) error {
	// Create destination file
	filePath := filepath.Join(s.storagePath, key)
	
	// Create directories if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Create destination file
	destFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer destFile.Close()
	
	// Copy content
	_, err = io.Copy(destFile, fileContent)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}
	
	return nil
}

// GetImageURL generates a URL to access the image
func (s *LocalStorageService) GetImageURL(ctx context.Context, key string) (string, error) {
	return "/images/" + key, nil
}

// DeleteImage removes an image from local storage
func (s *LocalStorageService) DeleteImage(ctx context.Context, key string) error {
	filePath := filepath.Join(s.storagePath, key)
	return os.Remove(filePath)
}

// GetImage gets an image from local storage
func (s *LocalStorageService) GetImage(ctx context.Context, key string) ([]byte, string, error) {
	filePath := filepath.Join(s.storagePath, key)
	
	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}
	
	// Determine content type based on file extension
	contentType := "image/jpeg" // Default
	ext := filepath.Ext(key)
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}
	
	return content, contentType, nil
}