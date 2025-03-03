package services

import (
	"bytes"
	"context"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalStorageService(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "image_gallery_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after the test

	// Create a local storage service
	service, err := NewLocalStorageService(tempDir)
	if err != nil {
		t.Fatalf("Failed to create local storage service: %v", err)
	}

	// Test GetBucketName
	t.Run("GetBucketName", func(t *testing.T) {
		if service.GetBucketName() != tempDir {
			t.Errorf("Expected bucket name %s, got %s", tempDir, service.GetBucketName())
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

		// Verify the file was created
		filePath := filepath.Join(tempDir, "test.jpg")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", filePath)
		}

		// Verify the content
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		if !bytes.Equal(content, imageContent) {
			t.Errorf("File content does not match expected")
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
		filePath := filepath.Join(tempDir, "test2.jpg")
		if err := os.WriteFile(filePath, imageContent, 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

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
		filePath := filepath.Join(tempDir, "delete.jpg")
		if err := os.WriteFile(filePath, imageContent, 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		ctx := context.Background()
		err := service.DeleteImage(ctx, "delete.jpg")
		if err != nil {
			t.Fatalf("Failed to delete image: %v", err)
		}

		// Verify the file was deleted
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Errorf("Expected file %s to be deleted", filePath)
		}
	})
}

// mockMultipartFile implements multipart.File for testing
type mockMultipartFile struct {
	*bytes.Reader
}

func (m *mockMultipartFile) Close() error {
	return nil
}

func (m *mockMultipartFile) ReadAt(p []byte, off int64) (n int, err error) {
	return m.Reader.ReadAt(p, off)
}

// Helper function to create a multipart file for testing
func createMultipartFile(t *testing.T, content []byte) multipart.File {
	t.Helper()
	
	// Create a reader for the content
	reader := bytes.NewReader(content)
	
	// Return a mock multipart file
	return &mockMultipartFile{Reader: reader}
}