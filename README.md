# Image Gallery

A web application built with Go that allows users to add, update, delete, and list images on a main page. The application stores image files in AWS S3 and metadata in AWS DynamoDB.

## Features

- Image upload with metadata (title, description) and image preview
- Image listing with gallery view and responsive design
- Image detail view with metadata display
- Edit image metadata
- Delete images with confirmation
- Environment configuration via .env files
- Integration with AWS S3 for image storage
- Integration with AWS DynamoDB for metadata storage
- Bootstrap 5 UI with modern design and icons
- Responsive layout for mobile and desktop
- Type-safe HTML templates with github.com/a-h/templ

## Prerequisites

- Go 1.24 or higher
- AWS account with S3 and DynamoDB access
- AWS credentials configured locally
- templ command-line tool (for template generation)

## Environment Configuration

The application supports loading configuration from environment variables or a `.env` file. 

1. Copy the example .env file:
   ```
   cp .env.example .env
   ```

2. Edit the .env file with your specific configuration:
   ```
   # AWS Configuration
   S3_BUCKET_NAME=your-s3-bucket-name
   DYNAMODB_TABLE_NAME=your-dynamodb-table-name
   
   # Server Configuration
   PORT=8080
   
   # AWS Region (optional)
   # AWS_REGION=us-east-1
   
   # AWS Credentials (optional)
   # AWS_ACCESS_KEY_ID=your_access_key
   # AWS_SECRET_ACCESS_KEY=your_secret_key
   ```

3. The application will automatically load the `.env` file at startup

You can also set these environment variables directly in your shell instead of using the .env file.

## AWS Setup

You can set up the required AWS resources (S3 bucket and DynamoDB table) using the provided script:

```
./scripts/setup-aws.sh
```

The script uses values from your `.env` file for the resource names, or falls back to default values if no `.env` file is found.

Alternatively, you can create the resources manually:

1. Create an S3 bucket:
   ```
   aws s3 mb s3://image-gallery-bucket
   ```

2. Set S3 bucket permissions to allow public read:
   ```
   aws s3api put-bucket-policy --bucket image-gallery-bucket --policy '{
     "Version": "2012-10-17",
     "Statement": [
       {
         "Sid": "PublicRead",
         "Effect": "Allow",
         "Principal": "*",
         "Action": ["s3:GetObject"],
         "Resource": ["arn:aws:s3:::image-gallery-bucket/*"]
       }
     ]
   }'
   ```

3. Create a DynamoDB table:
   ```
   aws dynamodb create-table \
     --table-name image-gallery-table \
     --attribute-definitions AttributeName=id,AttributeType=S \
     --key-schema AttributeName=id,KeyType=HASH \
     --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
   ```

## Building and Running

1. Install dependencies:
   ```
   make deps
   ```

2. Build the application:
   ```
   make build
   ```

3. Run the application:
   ```
   make run
   ```

The application will be available at `http://localhost:8080`.

## Development

- `make templ`: Generate templ components
- `make dev`: Build and run the application
- `make test`: Run tests
- `make clean`: Clean built files