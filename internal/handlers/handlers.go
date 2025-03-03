package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"image_gallery/internal/models"
	"image_gallery/internal/services"
)

// TemplateData contains data to be passed to templates
type TemplateData struct {
	PageName string      // Name of the template to render
	Data     interface{} // Data to pass to the template
}

// ImageHandler handles HTTP requests for images
type ImageHandler struct {
	storageService  services.StorageService
	databaseService services.DatabaseService
	templates       *template.Template
}

// NewImageHandler creates a new image handler
func NewImageHandler(storageService services.StorageService, databaseService services.DatabaseService) *ImageHandler {
	// Load templates with proper error handling
	templ, err := loadTemplates()
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}
	
	return &ImageHandler{
		storageService:  storageService,
		databaseService: databaseService,
		templates:       templ,
	}
}

// loadTemplates loads all HTML templates
func loadTemplates() (*template.Template, error) {
	// Create a template with layout.html as the base template
	templates := template.New("templates")
	
	// Add a function to format times
	templates = templates.Funcs(template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format("Jan 2, 2006 at 15:04")
		},
	})
	
	// Parse all templates
	return templates.ParseGlob("internal/templates/*.html")
}

// generateID creates a random ID for images
func generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return time.Now().Format("20060102150405")
	}
	return hex.EncodeToString(bytes)
}

// ListImages displays all images
func (h *ImageHandler) ListImages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	images, err := h.databaseService.ListImages(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch images", http.StatusInternalServerError)
		return
	}

	// Get URLs for each image
	for i := range images {
		url, err := h.storageService.GetImageURL(ctx, images[i].S3Key)
		if err == nil {
			images[i].S3Key = url
		}
	}

	// For API requests
	if r.Header.Get("Content-Type") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(images)
		return
	}

	// For web page requests
	templateData := TemplateData{
		PageName: "list",
		Data:     images,
	}
	
	if err := h.templates.ExecuteTemplate(w, "layout", templateData); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetImage gets a single image
func (h *ImageHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	ctx := r.Context()

	image, err := h.databaseService.GetImage(ctx, id)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	url, err := h.storageService.GetImageURL(ctx, image.S3Key)
	if err == nil {
		image.S3Key = url
	}

	// For API requests
	if r.Header.Get("Content-Type") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(image)
		return
	}

	// For web page requests
	templateData := TemplateData{
		PageName: "view",
		Data:     image,
	}
	
	if err := h.templates.ExecuteTemplate(w, "layout", templateData); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// UploadImageForm displays the form to upload an image
func (h *ImageHandler) UploadImageForm(w http.ResponseWriter, r *http.Request) {
	templateData := TemplateData{
		PageName: "upload",
		Data:     nil,
	}
	
	if err := h.templates.ExecuteTemplate(w, "layout", templateData); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// EditImageForm displays the form to edit an image
func (h *ImageHandler) EditImageForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	ctx := r.Context()

	image, err := h.databaseService.GetImage(ctx, id)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	templateData := TemplateData{
		PageName: "edit",
		Data:     image,
	}
	
	if err := h.templates.ExecuteTemplate(w, "layout", templateData); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// UploadImage handles image upload
func (h *ImageHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create image metadata
	id := generateID()
	s3Key := fmt.Sprintf("%s%s", id, filepath.Ext(handler.Filename))
	
	image := models.Image{
		ID:          id,
		Title:       r.FormValue("title"),
		Description: r.FormValue("description"),
		S3Key:       s3Key,
		ContentType: handler.Header.Get("Content-Type"),
		Size:        handler.Size,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Upload image to S3
	ctx := r.Context()
	err = h.storageService.UploadImage(ctx, s3Key, file, image.ContentType)
	if err != nil {
		http.Error(w, "Failed to upload image to S3", http.StatusInternalServerError)
		return
	}

	// Save metadata to DynamoDB
	err = h.databaseService.SaveImage(ctx, image)
	if err != nil {
		http.Error(w, "Failed to save image metadata", http.StatusInternalServerError)
		return
	}

	// Redirect to list page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// UpdateImage handles image update
func (h *ImageHandler) UpdateImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	ctx := r.Context()

	// Get existing image
	existingImage, err := h.databaseService.GetImage(ctx, id)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// Update fields
	existingImage.Title = r.FormValue("title")
	existingImage.Description = r.FormValue("description")
	existingImage.UpdatedAt = time.Now()

	// Save updated metadata to DynamoDB
	err = h.databaseService.SaveImage(ctx, existingImage)
	if err != nil {
		http.Error(w, "Failed to update image metadata", http.StatusInternalServerError)
		return
	}

	// Redirect to list page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ServeImage serves an image from S3
func (h *ImageHandler) ServeImage(w http.ResponseWriter, r *http.Request) {
	// Get the image key from the URL path
	imageKey := strings.TrimPrefix(r.URL.Path, "/images/")
	if imageKey == "" {
		http.Error(w, "Image key is required", http.StatusBadRequest)
		return
	}

	// Get the image from S3
	ctx := r.Context()
	content, contentType, err := h.storageService.GetImage(ctx, imageKey)
	if err != nil {
		log.Printf("Error getting image from S3: %v", err)
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// Set content type and write the image content
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 1 day
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// DeleteImage handles image deletion
func (h *ImageHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	ctx := r.Context()

	// Get existing image
	existingImage, err := h.databaseService.GetImage(ctx, id)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// Delete image from S3
	err = h.storageService.DeleteImage(ctx, existingImage.S3Key)
	if err != nil {
		http.Error(w, "Failed to delete image from S3", http.StatusInternalServerError)
		return
	}

	// Delete metadata from DynamoDB
	err = h.databaseService.DeleteImage(ctx, id)
	if err != nil {
		http.Error(w, "Failed to delete image metadata", http.StatusInternalServerError)
		return
	}

	// Redirect to list page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}