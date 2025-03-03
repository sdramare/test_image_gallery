.PHONY: build run clean test templ

# Generate templ templates
templ:
	templ generate ./internal/templates/components

# Build the application
build: templ
	go build -o bin/server ./cmd/server

# Run the application
run:
	go run ./cmd/server/main.go

# Clean built files
clean:
	rm -rf bin/

# Run tests
test:
	go test ./...

# Download dependencies
deps:
	go mod download

# Build and run application
dev: build run