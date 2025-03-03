#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
else
  echo "No .env file found. Using default values."
  export S3_BUCKET_NAME="image-gallery-bucket"
  export DYNAMODB_TABLE_NAME="image-gallery-table"
fi

# Create S3 bucket
echo "Creating S3 bucket: $S3_BUCKET_NAME"
aws s3 mb s3://$S3_BUCKET_NAME

# Set S3 bucket permissions to allow public read
echo "Setting S3 bucket public-read permissions"
aws s3api put-bucket-policy --bucket $S3_BUCKET_NAME --policy '{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicRead",
      "Effect": "Allow",
      "Principal": "*",
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::'$S3_BUCKET_NAME'/*"]
    }
  ]
}'

# Create DynamoDB table
echo "Creating DynamoDB table: $DYNAMODB_TABLE_NAME"
aws dynamodb create-table \
  --table-name $DYNAMODB_TABLE_NAME \
  --attribute-definitions AttributeName=id,AttributeType=S \
  --key-schema AttributeName=id,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5

echo "AWS setup complete!"