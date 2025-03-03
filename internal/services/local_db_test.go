package services

import (
	"context"
	"os"
	"testing"
	"time"
	
	"image_gallery/internal/models"
)

func TestLocalDBService(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "image_gallery_db_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after the test

	// Create a local DB service
	service, err := NewLocalDBService(tempDir)
	if err != nil {
		t.Fatalf("Failed to create local DB service: %v", err)
	}

	// Test SaveImage and GetImage
	t.Run("SaveAndGetImage", func(t *testing.T) {
		// Create a test image
		now := time.Now().Truncate(time.Second) // Truncate to remove monotonic clock
		testImage := models.Image{
			ID:          "test-id-1",
			Title:       "Test Image",
			Description: "A test image",
			S3Key:       "test1.jpg",
			ContentType: "image/jpeg",
			Size:        12345,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		ctx := context.Background()
		
		// Save the image
		err := service.SaveImage(ctx, testImage)
		if err != nil {
			t.Fatalf("Failed to save image: %v", err)
		}

		// Get the image
		retrievedImage, err := service.GetImage(ctx, testImage.ID)
		if err != nil {
			t.Fatalf("Failed to get image: %v", err)
		}

		// Compare fields
		if retrievedImage.ID != testImage.ID {
			t.Errorf("Expected ID %s, got %s", testImage.ID, retrievedImage.ID)
		}
		if retrievedImage.Title != testImage.Title {
			t.Errorf("Expected Title %s, got %s", testImage.Title, retrievedImage.Title)
		}
		if retrievedImage.Description != testImage.Description {
			t.Errorf("Expected Description %s, got %s", testImage.Description, retrievedImage.Description)
		}
		if retrievedImage.S3Key != testImage.S3Key {
			t.Errorf("Expected S3Key %s, got %s", testImage.S3Key, retrievedImage.S3Key)
		}
		if retrievedImage.ContentType != testImage.ContentType {
			t.Errorf("Expected ContentType %s, got %s", testImage.ContentType, retrievedImage.ContentType)
		}
		if retrievedImage.Size != testImage.Size {
			t.Errorf("Expected Size %d, got %d", testImage.Size, retrievedImage.Size)
		}
		if !retrievedImage.CreatedAt.Equal(testImage.CreatedAt) {
			t.Errorf("Expected CreatedAt %v, got %v", testImage.CreatedAt, retrievedImage.CreatedAt)
		}
		if !retrievedImage.UpdatedAt.Equal(testImage.UpdatedAt) {
			t.Errorf("Expected UpdatedAt %v, got %v", testImage.UpdatedAt, retrievedImage.UpdatedAt)
		}
	})

	// Test ListImages
	t.Run("ListImages", func(t *testing.T) {
		// Create a few test images
		now := time.Now().Truncate(time.Second)
		testImages := []models.Image{
			{
				ID:          "test-id-2",
				Title:       "Test Image 2",
				Description: "A test image 2",
				S3Key:       "test2.jpg",
				ContentType: "image/jpeg",
				Size:        12345,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          "test-id-3",
				Title:       "Test Image 3",
				Description: "A test image 3",
				S3Key:       "test3.jpg",
				ContentType: "image/jpeg",
				Size:        67890,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		}

		ctx := context.Background()
		
		// Save the images
		for _, img := range testImages {
			err := service.SaveImage(ctx, img)
			if err != nil {
				t.Fatalf("Failed to save image: %v", err)
			}
		}

		// List the images
		retrievedImages, err := service.ListImages(ctx)
		if err != nil {
			t.Fatalf("Failed to list images: %v", err)
		}

		// Verify we got at least 3 images (from both test runs)
		if len(retrievedImages) < 3 {
			t.Errorf("Expected at least 3 images, got %d", len(retrievedImages))
		}

		// Check that our new images are in the list
		foundImages := make(map[string]bool)
		for _, img := range retrievedImages {
			foundImages[img.ID] = true
		}

		for _, img := range testImages {
			if !foundImages[img.ID] {
				t.Errorf("Image with ID %s not found in list", img.ID)
			}
		}
	})

	// Test DeleteImage
	t.Run("DeleteImage", func(t *testing.T) {
		// Create a test image
		now := time.Now().Truncate(time.Second)
		testImage := models.Image{
			ID:          "test-id-delete",
			Title:       "Delete Test",
			Description: "A test image to delete",
			S3Key:       "delete-test.jpg",
			ContentType: "image/jpeg",
			Size:        12345,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		ctx := context.Background()
		
		// Save the image
		err := service.SaveImage(ctx, testImage)
		if err != nil {
			t.Fatalf("Failed to save image: %v", err)
		}

		// Delete the image
		err = service.DeleteImage(ctx, testImage.ID)
		if err != nil {
			t.Fatalf("Failed to delete image: %v", err)
		}

		// Try to get the deleted image, should fail
		_, err = service.GetImage(ctx, testImage.ID)
		if err == nil {
			t.Errorf("Expected error when getting deleted image, got nil")
		}
	})

	// Test persistence across service restarts
	t.Run("Persistence", func(t *testing.T) {
		// Create a test image
		now := time.Now().Truncate(time.Second)
		testImage := models.Image{
			ID:          "test-id-persistence",
			Title:       "Persistence Test",
			Description: "A test image for persistence",
			S3Key:       "persistence-test.jpg",
			ContentType: "image/jpeg",
			Size:        12345,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		ctx := context.Background()
		
		// Save the image
		err := service.SaveImage(ctx, testImage)
		if err != nil {
			t.Fatalf("Failed to save image: %v", err)
		}

		// Create a new service instance (simulating restart)
		newService, err := NewLocalDBService(tempDir)
		if err != nil {
			t.Fatalf("Failed to create new local DB service: %v", err)
		}

		// Get the image from the new service
		retrievedImage, err := newService.GetImage(ctx, testImage.ID)
		if err != nil {
			t.Fatalf("Failed to get image after restart: %v", err)
		}

		// Verify it's the same image
		if retrievedImage.ID != testImage.ID {
			t.Errorf("Expected ID %s, got %s", testImage.ID, retrievedImage.ID)
		}
		if retrievedImage.Title != testImage.Title {
			t.Errorf("Expected Title %s, got %s", testImage.Title, retrievedImage.Title)
		}
	})
}