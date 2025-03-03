package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"image_gallery/internal/models"
)

// LocalDBService is a local implementation of DynamoDBService for development
type LocalDBService struct {
	storagePath string
	images      map[string]models.Image
	mutex       sync.RWMutex
}

// Verify that LocalDBService implements DatabaseService
var _ DatabaseService = (*LocalDBService)(nil)

// NewLocalDBService creates a new local database service
func NewLocalDBService(storagePath string) (*LocalDBService, error) {
	// Create the full storage path
	dbPath := filepath.Join(storagePath, "db")
	
	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create DB storage directory: %w", err)
	}

	service := &LocalDBService{
		storagePath: dbPath,
		images:      make(map[string]models.Image),
	}

	// Try to load existing data
	if err := service.loadData(); err != nil {
		// It's okay if the file doesn't exist yet
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load existing DB: %w", err)
		}
	}

	return service, nil
}

// loadData loads image data from the local JSON file
func (d *LocalDBService) loadData() error {
	filePath := filepath.Join(d.storagePath, "images.json")
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var images []models.Image
	if err := json.Unmarshal(data, &images); err != nil {
		return fmt.Errorf("failed to parse images data: %w", err)
	}

	// Add all images to our map
	d.mutex.Lock()
	defer d.mutex.Unlock()
	
	for _, img := range images {
		d.images[img.ID] = img
	}

	return nil
}

// saveData saves image data to the local JSON file
func (d *LocalDBService) saveData() error {
	filePath := filepath.Join(d.storagePath, "images.json")
	
	// Convert map to slice
	d.mutex.RLock()
	images := make([]models.Image, 0, len(d.images))
	for _, img := range d.images {
		images = append(images, img)
	}
	d.mutex.RUnlock()
	
	// Marshal to JSON
	data, err := json.MarshalIndent(images, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal images data: %w", err)
	}
	
	// Save to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write images data: %w", err)
	}
	
	return nil
}

// SaveImage saves image metadata to local storage
func (d *LocalDBService) SaveImage(ctx context.Context, image models.Image) error {
	d.mutex.Lock()
	d.images[image.ID] = image
	d.mutex.Unlock()
	
	return d.saveData()
}

// GetImage retrieves an image by ID
func (d *LocalDBService) GetImage(ctx context.Context, id string) (models.Image, error) {
	d.mutex.RLock()
	image, exists := d.images[id]
	d.mutex.RUnlock()
	
	if !exists {
		return models.Image{}, fmt.Errorf("image not found")
	}
	
	return image, nil
}

// ListImages retrieves all images
func (d *LocalDBService) ListImages(ctx context.Context) ([]models.Image, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	
	images := make([]models.Image, 0, len(d.images))
	for _, img := range d.images {
		images = append(images, img)
	}
	
	// Sort by creation time (newest first) for newer images first
	// For this example, we'll skip sorting for simplicity
	
	return images, nil
}

// DeleteImage removes image metadata from local storage
func (d *LocalDBService) DeleteImage(ctx context.Context, id string) error {
	d.mutex.Lock()
	_, exists := d.images[id]
	if !exists {
		d.mutex.Unlock()
		return fmt.Errorf("image not found")
	}
	
	delete(d.images, id)
	d.mutex.Unlock()
	
	return d.saveData()
}