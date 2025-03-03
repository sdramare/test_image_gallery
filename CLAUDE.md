# CLAUDE.md - Guidelines for Image Gallery Go Project

## Build/Run/Test Commands
- Generate templ templates: `templ generate ./internal/templates/components`
- Build: `go build ./...`
- Run: `go run ./cmd/server/main.go`
- Test all: `go test ./...`
- Test with coverage: `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`
- Test single: `go test ./path/to/package -run TestName`
- Lint: `golint ./...` (install: `go install golang.org/x/lint/golint@latest`)
- Format: `go fmt ./...` or `gofmt -s -w .`
- Static analysis: `go vet ./...`
- Run with env file: `go run ./cmd/server/main.go`
- Full build with templ generation: `make build`

## Environment Configuration
- Copy example config: `cp .env.example .env`
- Edit `.env` file to set configuration values:
  - S3_BUCKET_NAME: AWS S3 bucket name
  - DYNAMODB_TABLE_NAME: AWS DynamoDB table name
  - PORT: Web server port number
  - USE_LOCAL_STORAGE: Set to "true" to use local storage (no AWS required)
  - LOCAL_STORAGE_PATH: Path for local storage (default: "./data/images")
  - AWS_REGION: (optional) AWS region
  - AWS_ACCESS_KEY_ID: (optional) AWS access key
  - AWS_SECRET_ACCESS_KEY: (optional) AWS secret key

## Code Style Guidelines
- Formatting: Follow standard Go formatting (gofmt)
- Imports: Group stdlib first, then third-party, then local packages
- Naming: CamelCase (PascalCase for exported, camelCase for unexported)
- Error handling: Always check errors, don't use underscore unless intentional
- Comments: Use godoc style comments (start with function name)
- File organization: One package per directory, package name matches dir name
- Testing: Use table-driven tests when appropriate
- Type definitions: Prefer composition over inheritance
- Concurrency: Use channels for communication, mutexes for state

## Project Structure
- `/cmd` - Main applications
- `/internal` - Private code
  - `/handlers` - HTTP handlers
  - `/models` - Data models
  - `/services` - Business logic
  - `/templates` - HTML templates
    - `/components` - Templ components
- `/web` - Web assets

## Templ Templates
- Templates are defined in `.templ` files in `/internal/templates/components/`
- Generated Go code is stored in `*_templ.go` files
- To render a template, use the helper functions in `templates.go`:
  - `components.RenderListPage(w, images)`
  - `components.RenderViewPage(w, image)`
  - `components.RenderUploadPage(w)`
  - `components.RenderEditPage(w, image)`