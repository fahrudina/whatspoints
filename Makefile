# Makefile for WhatsPoints
# Supports building for Ubuntu Linux and macOS

# Binary name
BINARY_NAME=whatspoints

# Build directory
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Version information
VERSION?=1.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Linker flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

.PHONY: all build clean test deps linux macos run help

# Default target
all: test build

# Build for current platform
build:
	@echo "Building $(BINARY_NAME) for current platform..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

# Build for Ubuntu Linux (amd64)
linux:
	@echo "Building $(BINARY_NAME) for Linux (amd64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

# Build for Ubuntu Linux (arm64)
linux-arm64:
	@echo "Building $(BINARY_NAME) for Linux (arm64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64"

# Build for macOS (Intel)
macos:
	@echo "Building $(BINARY_NAME) for macOS (amd64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64"

# Build for macOS (Apple Silicon)
macos-arm64:
	@echo "Building $(BINARY_NAME) for macOS (arm64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64"

# Build for all platforms
build-all: linux linux-arm64 macos macos-arm64
	@echo "All builds complete in $(BUILD_DIR)/"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Run the application
run: build
	@echo "Starting $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Run with QR code pairing
add-sender:
	@echo "Adding new sender with QR code..."
	./$(BINARY_NAME) -add-sender

# Run with phone pairing code
add-sender-code:
	@read -p "Enter phone number (e.g., +1234567890): " phone; \
	./$(BINARY_NAME) -add-sender-code=$$phone

# Clear all sessions
clear-sessions:
	@echo "Clearing all sessions..."
	./$(BINARY_NAME) -clear-sessions

# Install the binary to $GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	cp $(BINARY_NAME) $(GOPATH)/bin/
	@echo "Install complete"

# Show help
help:
	@echo "WhatsPoints Makefile Commands:"
	@echo ""
	@echo "Building:"
	@echo "  make build          - Build for current platform"
	@echo "  make linux          - Build for Ubuntu Linux (amd64)"
	@echo "  make linux-arm64    - Build for Ubuntu Linux (arm64)"
	@echo "  make macos          - Build for macOS (Intel/amd64)"
	@echo "  make macos-arm64    - Build for macOS (Apple Silicon/arm64)"
	@echo "  make build-all      - Build for all platforms"
	@echo ""
	@echo "Testing:"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo ""
	@echo "Running:"
	@echo "  make run            - Build and run the application"
	@echo "  make add-sender     - Add new sender with QR code"
	@echo "  make add-sender-code- Add new sender with phone pairing"
	@echo "  make clear-sessions - Clear all WhatsApp sessions"
	@echo ""
	@echo "Maintenance:"
	@echo "  make deps           - Download and tidy dependencies"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make install        - Install binary to GOPATH/bin"
	@echo ""
	@echo "Environment Variables:"
	@echo "  VERSION             - Set version number (default: 1.0.0)"
	@echo ""
	@echo "Examples:"
	@echo "  make build VERSION=1.2.3"
	@echo "  make build-all"
	@echo "  make test"
