.PHONY: build test lint clean build-all

# Binary name
BINARY := coding-agent
ifeq ($(OS),Windows_NT)
	BINARY := $(BINARY).exe
endif

# Build for current platform
build:
	go build -trimpath -o $(BINARY) ./cmd/coding-agent

# Run all tests
test:
	go test ./... -v

# Run tests with coverage
test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Run linter (requires golangci-lint)
lint:
	golangci-lint run ./...

# Clean build artifacts
clean:
	rm -f $(BINARY) coding-agent
	rm -f coverage.out coverage.html

# Cross-compile for all platforms
build-all:
	GOOS=windows GOARCH=amd64 go build -trimpath -o coding-agent-windows-amd64.exe ./cmd/coding-agent
	GOOS=linux   GOARCH=amd64 go build -trimpath -o coding-agent-linux-amd64 ./cmd/coding-agent
	GOOS=darwin  GOARCH=amd64 go build -trimpath -o coding-agent-darwin-amd64 ./cmd/coding-agent
	GOOS=darwin  GOARCH=arm64 go build -trimpath -o coding-agent-darwin-arm64 ./cmd/coding-agent

# Build with stripped debug info (smaller binary)
build-release:
	go build -trimpath -ldflags="-s -w" -o $(BINARY) ./cmd/coding-agent

# Show help
help:
	@echo "CodingAgent — Go Edition"
	@echo ""
	@echo "Targets:"
	@echo "  build         Build for current platform"
	@echo "  build-release Build with stripped debug info"
	@echo "  build-all     Cross-compile for all platforms"
	@echo "  test          Run all tests"
	@echo "  test-cover    Run tests with coverage report"
	@echo "  lint          Run linter"
	@echo "  clean         Remove build artifacts"
