package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"image_gallery/internal/handlers"
	"image_gallery/internal/services"
)

// loadEnv loads environment variables from .env files
func loadEnv() {
	// Try to load from .env file
	if err := godotenv.Load(); err != nil {
		// Try to load from .env file in project root
		wd, err := os.Getwd()
		if err == nil {
			for {
				envPath := filepath.Join(wd, ".env")
				if _, err := os.Stat(envPath); err == nil {
					godotenv.Load(envPath)
					break
				}
				
				parent := filepath.Dir(wd)
				if parent == wd {
					break
				}
				wd = parent
			}
		}
	}
}

func main() {
	// Load environment variables from .env file
	loadEnv()
	
	// Get configuration from environment variables
	s3BucketName := getEnv("S3_BUCKET_NAME", "image-gallery-bucket")
	dynamoDBTableName := getEnv("DYNAMODB_TABLE_NAME", "image-gallery-table")
	port := getEnv("PORT", "8080")
	localStoragePath := getEnv("LOCAL_STORAGE_PATH", "./data/images")
	useLocalStorage := getEnv("USE_LOCAL_STORAGE", "true") == "true"

	var storageService services.StorageService
	var databaseService services.DatabaseService
	var err error

	if useLocalStorage {
		// Use local file storage instead of S3
		log.Println("Using local storage at:", localStoragePath)
		storageService, err = services.NewLocalStorageService(localStoragePath)
		if err != nil {
			log.Fatalf("Failed to initialize local storage: %v", err)
		}
		
		// Use local DB service
		log.Println("Using local database at:", localStoragePath)
		databaseService, err = services.NewLocalDBService(localStoragePath)
		if err != nil {
			log.Fatalf("Failed to initialize local database: %v", err)
		}
	} else {
		// Initialize AWS SDK configuration
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			log.Fatalf("Failed to load AWS configuration: %v", err)
		}

		// Create AWS service clients
		s3Client := s3.NewFromConfig(cfg)
		dynamoDBClient := dynamodb.NewFromConfig(cfg)

		// Create services
		storageService = services.NewS3Service(s3Client, s3BucketName)
		databaseService = services.NewDynamoDBService(dynamoDBClient, dynamoDBTableName)
	}

	// Create handlers
	imageHandler := handlers.NewImageHandler(storageService, databaseService)

	// Set up router
	router := mux.NewRouter()

	// Define routes
	router.HandleFunc("/", imageHandler.ListImages).Methods("GET")
	router.HandleFunc("/image/{id}", imageHandler.GetImage).Methods("GET")
	router.HandleFunc("/upload", imageHandler.UploadImageForm).Methods("GET")
	router.HandleFunc("/upload", imageHandler.UploadImage).Methods("POST")
	router.HandleFunc("/edit/{id}", imageHandler.EditImageForm).Methods("GET")
	router.HandleFunc("/update/{id}", imageHandler.UpdateImage).Methods("POST")
	router.HandleFunc("/delete/{id}", imageHandler.DeleteImage).Methods("POST")

	// Handle image proxy to S3
	router.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.HandlerFunc(imageHandler.ServeImage)))

	// Start server
	log.Printf("Server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

// getEnv gets an environment variable or returns the default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}