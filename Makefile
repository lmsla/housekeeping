.PHONY: build test test-race test-coverage clean run fmt vet

# Binary name
BINARY_NAME=housekeeping
BUILD_DIR=bin

# Build the application
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/housekeeping

# Run tests
test:
	go test ./... -v

# Run tests with race detector
test-race:
	GOCACHE=$${GOCACHE:-/tmp/go-build} go test -race -count=1 ./...

# Run tests with coverage
test-coverage:
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Run the application (requires configs in current directory)
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Lint and format
lint: fmt vet

# Tidy dependencies
tidy:
	go mod tidy

# Download dependencies
deps:
	go mod download

# Build for multiple platforms
build-all:
	@mkdir -p $(BUILD_DIR)
### Linux 64-bit (Ubuntu/RedHat x86_64)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/housekeeping
### macOS Intel
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/housekeeping
### macOS Apple Silicon
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/housekeeping
