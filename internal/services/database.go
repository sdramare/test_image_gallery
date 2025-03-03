package services

import (
	"context"
	
	"image_gallery/internal/models"
)

// DatabaseService defines the common interface for database services
type DatabaseService interface {
	// SaveImage saves image metadata to database
	SaveImage(ctx context.Context, image models.Image) error
	
	// GetImage retrieves an image by ID
	GetImage(ctx context.Context, id string) (models.Image, error)
	
	// ListImages retrieves all images
	ListImages(ctx context.Context) ([]models.Image, error)
	
	// DeleteImage removes image metadata from database
	DeleteImage(ctx context.Context, id string) error
}