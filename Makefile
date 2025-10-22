.PHONY: build run clean test install-deps fmt lint

# Build the application
build:
	@echo "Building VoicePilot-Eino..."
	go build -o bin/voicepilot-eino cmd/server/main.go

# Run the application
run:
	@echo "Running VoicePilot-Eino..."
	go run cmd/server/main.go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf temp/*
	rm -rf static/audio/*

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Install dependencies
install-deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Initialize project (first time setup)
init:
	@echo "Initializing project..."
	mkdir -p bin temp static/audio
	cp .env.example .env
	@echo "Project initialized. Please edit .env file with your configuration."

# Build for production
build-prod:
	@echo "Building for production..."
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/voicepilot-eino cmd/server/main.go

# Run in development mode with hot reload (requires air)
dev:
	@echo "Starting development server..."
	air
