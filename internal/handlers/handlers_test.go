package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"io"
	"image_gallery/internal/models"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// MockStorageService implements the StorageService interface for testing
type MockStorageService struct {
	images map[string][]byte // Map of image keys to image content
}

func NewMockStorageService() *MockStorageService {
	return &MockStorageService{
		images: make(map[string][]byte),
	}
}

func (m *MockStorageService) UploadImage(_ context.Context, key string, fileContent multipart.File, _ string) error {
	// Read the file content
	content, err := io.ReadAll(fileContent)
	if err != nil {
		return err
	}

	// Store the content
	m.images[key] = content
	return nil
}

func (m *MockStorageService) GetImageURL(_ context.Context, key string) (string, error) {
	return "/images/" + key, nil
}

func (m *MockStorageService) GetImage(_ context.Context, key string) ([]byte, string, error) {
	content, ok := m.images[key]
	if !ok {
		return nil, "", os.ErrNotExist
	}
	return content, "image/jpeg", nil
}

func (m *MockStorageService) DeleteImage(_ context.Context, key string) error {
	delete(m.images, key)
	return nil
}

func (m *MockStorageService) GetBucketName() string {
	return "mock-bucket"
}

// MockDatabaseService implements the DatabaseService interface for testing
type MockDatabaseService struct {
	images map[string]models.Image
}

func NewMockDatabaseService() *MockDatabaseService {
	return &MockDatabaseService{
		images: make(map[string]models.Image),
	}
}

func (m *MockDatabaseService) SaveImage(_ context.Context, image models.Image) error {
	m.images[image.ID] = image
	return nil
}

func (m *MockDatabaseService) GetImage(_ context.Context, id string) (models.Image, error) {
	image, ok := m.images[id]
	if !ok {
		return models.Image{}, os.ErrNotExist
	}
	return image, nil
}

func (m *MockDatabaseService) ListImages(_ context.Context) ([]models.Image, error) {
	images := make([]models.Image, 0, len(m.images))
	for _, img := range m.images {
		images = append(images, img)
	}
	return images, nil
}

func (m *MockDatabaseService) DeleteImage(_ context.Context, id string) error {
	delete(m.images, id)
	return nil
}

// createMockTemplates creates mock templates for testing
func createMockTemplates() *template.Template {
	// Create a simple template with the required templates
	tmpl := template.New("templates")
	
	// Add a content template
	template.Must(tmpl.New("content").Parse(`
		{{range .}}
		<div>{{.Title}}</div>
		<div>{{.Description}}</div>
		{{end}}
	`))
	
	// Add a layout template
	template.Must(tmpl.New("layout").Parse(`
		<html>
		<body>
		{{if eq .PageName "list"}}
			{{template "content" .Data}}
		{{else}}
			Unknown page
		{{end}}
		</body>
		</html>
	`))
	
	// Add view, upload, and edit templates
	template.Must(tmpl.New("view-content").Parse(`View: {{.Title}}`))
	template.Must(tmpl.New("upload-content").Parse(`Upload form`))
	template.Must(tmpl.New("edit-content").Parse(`Edit: {{.Title}}`))
	
	return tmpl
}

func TestListImages(t *testing.T) {
	// Set up mock services
	mockStorage := NewMockStorageService()
	mockDB := NewMockDatabaseService()

	// Add some test images to the mock DB
	now := time.Now().Truncate(time.Second)
	testImages := []models.Image{
		{
			ID:          "test-id-1",
			Title:       "Test Image 1",
			Description: "A test image 1",
			S3Key:       "test1.jpg",
			ContentType: "image/jpeg",
			Size:        12345,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "test-id-2",
			Title:       "Test Image 2",
			Description: "A test image 2",
			S3Key:       "test2.jpg",
			ContentType: "image/jpeg",
			Size:        67890,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	for _, img := range testImages {
		mockDB.SaveImage(context.Background(), img)
	}

	// Create a handler with a mock template
	handler := &ImageHandler{
		storageService:  mockStorage,
		databaseService: mockDB,
		templates:       createMockTemplates(),
	}

	// Test HTML output
	t.Run("ListImages_HTML", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		rr := httptest.NewRecorder()
		handler.ListImages(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Check that the response contains the image titles
		for _, img := range testImages {
			if !strings.Contains(rr.Body.String(), img.Title) {
				t.Errorf("Response body does not contain image title: %s", img.Title)
			}
		}
	})

	// Test JSON output
	t.Run("ListImages_JSON", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.ListImages(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Parse the response
		var responseImages []models.Image
		if err := json.Unmarshal(rr.Body.Bytes(), &responseImages); err != nil {
			t.Fatalf("Failed to parse response JSON: %v", err)
		}

		// Check that all test images are in the response
		if len(responseImages) != len(testImages) {
			t.Errorf("Expected %d images, got %d", len(testImages), len(responseImages))
		}

		// Create a map of image IDs for easy lookup
		imageMap := make(map[string]models.Image)
		for _, img := range responseImages {
			imageMap[img.ID] = img
		}

		// Check that all test images are in the response
		for _, img := range testImages {
			responseImg, ok := imageMap[img.ID]
			if !ok {
				t.Errorf("Image with ID %s not found in response", img.ID)
				continue
			}

			if responseImg.Title != img.Title {
				t.Errorf("Expected title %s, got %s", img.Title, responseImg.Title)
			}
		}
	})
}

func TestServeImage(t *testing.T) {
	// Set up mock services
	mockStorage := NewMockStorageService()
	mockDB := NewMockDatabaseService()

	// Add a test image to the mock storage
	imageContent := []byte("test image content")
	mockStorage.images["test.jpg"] = imageContent

	// Create a handler with a mock template
	handler := &ImageHandler{
		storageService:  mockStorage,
		databaseService: mockDB,
		templates:       createMockTemplates(),
	}

	t.Run("ServeImage_Success", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/images/test.jpg", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.URL.Path = "/images/test.jpg" // Explicitly set for the ServeImage handler

		rr := httptest.NewRecorder()
		handler.ServeImage(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Check the content type
		contentType := rr.Header().Get("Content-Type")
		if contentType != "image/jpeg" {
			t.Errorf("Handler returned wrong content type: got %v want %v", contentType, "image/jpeg")
		}

		// Check the response body
		if !bytes.Equal(rr.Body.Bytes(), imageContent) {
			t.Errorf("Handler returned unexpected body")
		}
	})

	t.Run("ServeImage_NotFound", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/images/nonexistent.jpg", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.URL.Path = "/images/nonexistent.jpg"

		rr := httptest.NewRecorder()
		handler.ServeImage(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
		}
	})
}