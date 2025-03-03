package models

import (
	"time"
)

// Image represents image metadata stored in DynamoDB
type Image struct {
	ID          string    `json:"id" dynamodbav:"id"`
	Title       string    `json:"title" dynamodbav:"title"`
	Description string    `json:"description" dynamodbav:"description"`
	S3Key       string    `json:"s3Key" dynamodbav:"s3Key"`
	ContentType string    `json:"contentType" dynamodbav:"contentType"`
	Size        int64     `json:"size" dynamodbav:"size"`
	CreatedAt   time.Time `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" dynamodbav:"updatedAt"`
}