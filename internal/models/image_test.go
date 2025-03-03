package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestImageSerialization(t *testing.T) {
	// Create a test image
	now := time.Now().Truncate(time.Second) // Truncate to remove monotonic clock
	testImage := Image{
		ID:          "test-id-1",
		Title:       "Test Image",
		Description: "A test image",
		S3Key:       "test1.jpg",
		ContentType: "image/jpeg",
		Size:        12345,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Test JSON serialization
	t.Run("JSON_Marshal", func(t *testing.T) {
		// Marshal the image to JSON
		data, err := json.Marshal(testImage)
		if err != nil {
			t.Fatalf("Failed to marshal image to JSON: %v", err)
		}

		// Unmarshal back to an image
		var unmarshaledImage Image
		if err := json.Unmarshal(data, &unmarshaledImage); err != nil {
			t.Fatalf("Failed to unmarshal image from JSON: %v", err)
		}

		// Verify fields match
		if unmarshaledImage.ID != testImage.ID {
			t.Errorf("Expected ID %s, got %s", testImage.ID, unmarshaledImage.ID)
		}
		if unmarshaledImage.Title != testImage.Title {
			t.Errorf("Expected Title %s, got %s", testImage.Title, unmarshaledImage.Title)
		}
		if unmarshaledImage.Description != testImage.Description {
			t.Errorf("Expected Description %s, got %s", testImage.Description, unmarshaledImage.Description)
		}
		if unmarshaledImage.S3Key != testImage.S3Key {
			t.Errorf("Expected S3Key %s, got %s", testImage.S3Key, unmarshaledImage.S3Key)
		}
		if unmarshaledImage.ContentType != testImage.ContentType {
			t.Errorf("Expected ContentType %s, got %s", testImage.ContentType, unmarshaledImage.ContentType)
		}
		if unmarshaledImage.Size != testImage.Size {
			t.Errorf("Expected Size %d, got %d", testImage.Size, unmarshaledImage.Size)
		}
		
		// Verify time fields are equal with a small tolerance
		if !unmarshaledImage.CreatedAt.Equal(testImage.CreatedAt) {
			t.Errorf("Expected CreatedAt %v, got %v", testImage.CreatedAt, unmarshaledImage.CreatedAt)
		}
		if !unmarshaledImage.UpdatedAt.Equal(testImage.UpdatedAt) {
			t.Errorf("Expected UpdatedAt %v, got %v", testImage.UpdatedAt, unmarshaledImage.UpdatedAt)
		}
	})
}