package components

import (
	"bytes"
	"context"
	"image_gallery/internal/models"
	"strings"
	"testing"
	"time"
)

func TestListComponent(t *testing.T) {
	// Create test images
	now := time.Now().Truncate(time.Second)
	images := []models.Image{
		{
			ID:          "test-id-1",
			Title:       "Test Image 1",
			Description: "A test image 1",
			S3Key:       "/images/test1.jpg",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	// Render the component
	var buf bytes.Buffer
	err := List(images).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render list component: %v", err)
	}

	// Check that the output contains expected content
	output := buf.String()
	if !strings.Contains(output, "Test Image 1") {
		t.Errorf("List component output does not contain image title")
	}
	if !strings.Contains(output, "A test image 1") {
		t.Errorf("List component output does not contain image description")
	}
	if !strings.Contains(output, "/images/test1.jpg") {
		t.Errorf("List component output does not contain image URL")
	}
}

func TestViewComponent(t *testing.T) {
	// Create test image
	now := time.Now().Truncate(time.Second)
	image := models.Image{
		ID:          "test-id-1",
		Title:       "Test Image 1",
		Description: "A test image 1",
		S3Key:       "/images/test1.jpg",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Render the component
	var buf bytes.Buffer
	err := View(image).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render view component: %v", err)
	}

	// Check that the output contains expected content
	output := buf.String()
	if !strings.Contains(output, "Test Image 1") {
		t.Errorf("View component output does not contain image title")
	}
	if !strings.Contains(output, "A test image 1") {
		t.Errorf("View component output does not contain image description")
	}
	if !strings.Contains(output, "/images/test1.jpg") {
		t.Errorf("View component output does not contain image URL")
	}
}

func TestEditComponent(t *testing.T) {
	// Create test image
	now := time.Now().Truncate(time.Second)
	image := models.Image{
		ID:          "test-id-1",
		Title:       "Test Image 1",
		Description: "A test image 1", 
		S3Key:       "/images/test1.jpg",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Render the component
	var buf bytes.Buffer
	err := Edit(image).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render edit component: %v", err)
	}

	// Check that the output contains expected content
	output := buf.String()
	if !strings.Contains(output, "Test Image 1") {
		t.Errorf("Edit component output does not contain image title")
	}
	if !strings.Contains(output, "A test image 1") {
		t.Errorf("Edit component output does not contain image description")
	}
	if !strings.Contains(output, "/images/test1.jpg") {
		t.Errorf("Edit component output does not contain image URL")
	}
}

func TestUploadComponent(t *testing.T) {
	// Render the component
	var buf bytes.Buffer
	err := Upload().Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Failed to render upload component: %v", err)
	}

	// Check that the output contains expected content
	output := buf.String()
	if !strings.Contains(output, "Upload New Image") {
		t.Errorf("Upload component output does not contain expected title")
	}
	if !strings.Contains(output, "enctype=\"multipart/form-data\"") {
		t.Errorf("Upload component output does not contain multipart form")
	}
}