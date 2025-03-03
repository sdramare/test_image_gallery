package services

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	
	"image_gallery/internal/models"
)

// DynamoDBService handles operations with AWS DynamoDB
type DynamoDBService struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBService creates a new DynamoDB service
func NewDynamoDBService(client *dynamodb.Client, tableName string) *DynamoDBService {
	return &DynamoDBService{
		client:    client,
		tableName: tableName,
	}
}

// Verify that DynamoDBService implements DatabaseService
var _ DatabaseService = (*DynamoDBService)(nil)

// SaveImage saves image metadata to DynamoDB
func (d *DynamoDBService) SaveImage(ctx context.Context, image models.Image) error {
	item, err := attributevalue.MarshalMap(image)
	if err != nil {
		return err
	}

	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      item,
	})

	return err
}

// GetImage retrieves an image by ID
func (d *DynamoDBService) GetImage(ctx context.Context, id string) (models.Image, error) {
	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return models.Image{}, err
	}

	if result.Item == nil {
		return models.Image{}, errors.New("image not found")
	}

	var image models.Image
	err = attributevalue.UnmarshalMap(result.Item, &image)
	if err != nil {
		return models.Image{}, err
	}

	return image, nil
}

// ListImages retrieves all images
func (d *DynamoDBService) ListImages(ctx context.Context) ([]models.Image, error) {
	result, err := d.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(d.tableName),
	})
	if err != nil {
		return nil, err
	}

	var images []models.Image
	err = attributevalue.UnmarshalListOfMaps(result.Items, &images)
	if err != nil {
		return nil, err
	}

	return images, nil
}

// DeleteImage removes image metadata from DynamoDB
func (d *DynamoDBService) DeleteImage(ctx context.Context, id string) error {
	_, err := d.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})

	return err
}