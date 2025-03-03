package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	
	"image_gallery/internal/models"
)

// Define custom errors for our mock
var (
	ErrMissingKey = errors.New("missing required key 'id'")
	ErrInvalidKeyType = errors.New("ID must be a string")
)

// mockDynamoDBClient is a mock implementation of the DynamoDB client for testing
type mockDynamoDBClient struct {
	items map[string]map[string]types.AttributeValue
}

// PutItem mocks the DynamoDB PutItem operation
func (m *mockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	// Extract the table name
	tableName := aws.ToString(params.TableName)
	
	// Initialize the items map if not already done
	if m.items == nil {
		m.items = make(map[string]map[string]types.AttributeValue)
	}
	
	// Extract the item ID (we assume the primary key is called "id")
	idAttr, ok := params.Item["id"]
	if !ok {
		return nil, ErrMissingKey
	}
	
	// Convert the ID to a string (we assume it's a string value)
	idVal, ok := idAttr.(*types.AttributeValueMemberS)
	if !ok {
		return nil, ErrInvalidKeyType
	}
	
	// Store the item
	m.items[tableName+"/"+idVal.Value] = params.Item
	
	return &dynamodb.PutItemOutput{}, nil
}

// GetItem mocks the DynamoDB GetItem operation
func (m *mockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	// Extract the table name
	tableName := aws.ToString(params.TableName)
	
	// Extract the item ID (we assume the primary key is called "id")
	idAttr, ok := params.Key["id"]
	if !ok {
		return nil, ErrMissingKey
	}
	
	// Convert the ID to a string (we assume it's a string value)
	idVal, ok := idAttr.(*types.AttributeValueMemberS)
	if !ok {
		return nil, ErrInvalidKeyType
	}
	
	// Check if the item exists
	item, ok := m.items[tableName+"/"+idVal.Value]
	if !ok {
		// Item not found, return an empty result (not an error)
		return &dynamodb.GetItemOutput{}, nil
	}
	
	// Return the item
	return &dynamodb.GetItemOutput{
		Item: item,
	}, nil
}

// Scan mocks the DynamoDB Scan operation
func (m *mockDynamoDBClient) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	// Extract the table name
	tableName := aws.ToString(params.TableName)
	
	// Collect all items for the table
	var items []map[string]types.AttributeValue
	prefix := tableName + "/"
	
	for key, item := range m.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			items = append(items, item)
		}
	}
	
	// Return the items
	return &dynamodb.ScanOutput{
		Items: items,
	}, nil
}

// DeleteItem mocks the DynamoDB DeleteItem operation
func (m *mockDynamoDBClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	// Extract the table name
	tableName := aws.ToString(params.TableName)
	
	// Extract the item ID (we assume the primary key is called "id")
	idAttr, ok := params.Key["id"]
	if !ok {
		return nil, ErrMissingKey
	}
	
	// Convert the ID to a string (we assume it's a string value)
	idVal, ok := idAttr.(*types.AttributeValueMemberS)
	if !ok {
		return nil, ErrInvalidKeyType
	}
	
	// Delete the item
	delete(m.items, tableName+"/"+idVal.Value)
	
	return &dynamodb.DeleteItemOutput{}, nil
}

// dynamoDBClientInterface defines the interface for DynamoDB client
type dynamoDBClientInterface interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

// Ensure mockDynamoDBClient implements the interface
var _ dynamoDBClientInterface = (*mockDynamoDBClient)(nil)

// TestDynamoDBServiceImpl is a custom implementation of DynamoDBService for testing
type TestDynamoDBServiceImpl struct {
	client    dynamoDBClientInterface
	tableName string
}

func (d *TestDynamoDBServiceImpl) SaveImage(ctx context.Context, image models.Image) error {
	// Marshal the image to DynamoDB attributes
	item, err := createMockImageItem(image)
	if err != nil {
		return err
	}

	// Save to DynamoDB
	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      item,
	})

	return err
}

func (d *TestDynamoDBServiceImpl) GetImage(ctx context.Context, id string) (models.Image, error) {
	// Create the key for DynamoDB
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: id},
	}

	// Get the item from DynamoDB
	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key:       key,
	})
	if err != nil {
		return models.Image{}, err
	}

	// Check if the item was found
	if result.Item == nil || len(result.Item) == 0 {
		return models.Image{}, errors.New("image not found")
	}

	// Convert the result to an Image
	return mockItemToImage(result.Item)
}

func (d *TestDynamoDBServiceImpl) ListImages(ctx context.Context) ([]models.Image, error) {
	// Scan all items in the table
	result, err := d.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(d.tableName),
	})
	if err != nil {
		return nil, err
	}

	// Convert the results to Images
	images := make([]models.Image, 0, len(result.Items))
	for _, item := range result.Items {
		image, err := mockItemToImage(item)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	return images, nil
}

func (d *TestDynamoDBServiceImpl) DeleteImage(ctx context.Context, id string) error {
	// Create the key for DynamoDB
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: id},
	}

	// Delete the item from DynamoDB
	_, err := d.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(d.tableName),
		Key:       key,
	})

	return err
}

// Helper functions for converting between models.Image and DynamoDB items
func createMockImageItem(image models.Image) (map[string]types.AttributeValue, error) {
	// In a real application, we would use attributevalue.MarshalMap
	// For the test, we'll create a simple map manually
	return map[string]types.AttributeValue{
		"id":          &types.AttributeValueMemberS{Value: image.ID},
		"title":       &types.AttributeValueMemberS{Value: image.Title},
		"description": &types.AttributeValueMemberS{Value: image.Description},
		"s3Key":       &types.AttributeValueMemberS{Value: image.S3Key},
		"contentType": &types.AttributeValueMemberS{Value: image.ContentType},
		"size":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", image.Size)},
		"createdAt":   &types.AttributeValueMemberS{Value: image.CreatedAt.Format(time.RFC3339)},
		"updatedAt":   &types.AttributeValueMemberS{Value: image.UpdatedAt.Format(time.RFC3339)},
	}, nil
}

func mockItemToImage(item map[string]types.AttributeValue) (models.Image, error) {
	// In a real application, we would use attributevalue.UnmarshalMap
	// For the test, we'll extract values manually
	image := models.Image{}

	// Extract ID
	if idAttr, ok := item["id"]; ok {
		if idVal, ok := idAttr.(*types.AttributeValueMemberS); ok {
			image.ID = idVal.Value
		}
	}

	// Extract Title
	if titleAttr, ok := item["title"]; ok {
		if titleVal, ok := titleAttr.(*types.AttributeValueMemberS); ok {
			image.Title = titleVal.Value
		}
	}

	// Extract Description
	if descAttr, ok := item["description"]; ok {
		if descVal, ok := descAttr.(*types.AttributeValueMemberS); ok {
			image.Description = descVal.Value
		}
	}

	// Extract S3Key
	if s3KeyAttr, ok := item["s3Key"]; ok {
		if s3KeyVal, ok := s3KeyAttr.(*types.AttributeValueMemberS); ok {
			image.S3Key = s3KeyVal.Value
		}
	}

	// Extract ContentType
	if contentTypeAttr, ok := item["contentType"]; ok {
		if contentTypeVal, ok := contentTypeAttr.(*types.AttributeValueMemberS); ok {
			image.ContentType = contentTypeVal.Value
		}
	}

	// Extract Size
	if sizeAttr, ok := item["size"]; ok {
		if sizeVal, ok := sizeAttr.(*types.AttributeValueMemberN); ok {
			size, err := parseInt64(sizeVal.Value)
			if err == nil {
				image.Size = size
			}
		}
	}

	// Extract CreatedAt
	if createdAtAttr, ok := item["createdAt"]; ok {
		if createdAtVal, ok := createdAtAttr.(*types.AttributeValueMemberS); ok {
			createdAt, err := time.Parse(time.RFC3339, createdAtVal.Value)
			if err == nil {
				image.CreatedAt = createdAt
			}
		}
	}

	// Extract UpdatedAt
	if updatedAtAttr, ok := item["updatedAt"]; ok {
		if updatedAtVal, ok := updatedAtAttr.(*types.AttributeValueMemberS); ok {
			updatedAt, err := time.Parse(time.RFC3339, updatedAtVal.Value)
			if err == nil {
				image.UpdatedAt = updatedAt
			}
		}
	}

	return image, nil
}

// Helper to parse int64
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// Ensure TestDynamoDBServiceImpl implements DatabaseService
var _ DatabaseService = (*TestDynamoDBServiceImpl)(nil)

func TestDynamoDBService(t *testing.T) {
	// Create a mock DynamoDB client
	mockClient := &mockDynamoDBClient{
		items: make(map[string]map[string]types.AttributeValue),
	}
	
	// Create a DynamoDB service with the mock client
	tableName := "test-table"
	// Create our test service
	service := &TestDynamoDBServiceImpl{
		client:    mockClient,
		tableName: tableName,
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
		
		// Note: Time comparison is tricky with DynamoDB since it may not preserve time precision
		// We'll use string comparison for dates to avoid issues with time zones and precision
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
		
		// Try to get the deleted image, should return a "not found" error
		_, err = service.GetImage(ctx, testImage.ID)
		if err == nil {
			t.Errorf("Expected error when getting deleted image, got nil")
		}
	})
}