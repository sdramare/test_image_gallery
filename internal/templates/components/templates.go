package components

import (
	"context"
	"image_gallery/internal/models"
	"net/http"
)

// RenderListPage renders the list page with the given images
func RenderListPage(w http.ResponseWriter, images []models.Image) error {
	return Layout(List(images)).Render(context.Background(), w)
}

// RenderViewPage renders the view page for a single image
func RenderViewPage(w http.ResponseWriter, image models.Image) error {
	return Layout(View(image)).Render(context.Background(), w)
}

// RenderUploadPage renders the upload form
func RenderUploadPage(w http.ResponseWriter) error {
	return Layout(Upload()).Render(context.Background(), w)
}

// RenderEditPage renders the edit form for an image
func RenderEditPage(w http.ResponseWriter, image models.Image) error {
	return Layout(Edit(image)).Render(context.Background(), w)
}