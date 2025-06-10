# Makefile for actions-toolkit

# Variables
BINARY_NAME=actions-toolkit
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "0.0.0-SNAPSHOT")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "ffffff")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/behnh/actions-toolkit/cmd.version=$(VERSION) -X github.com/behnh/actions-toolkit/cmd.gitCommit=$(GIT_COMMIT)"

GO_FILES=$(shell find . -name "*.go" -type f)
GO_BUILD=go build
GO_TEST=go test
GO_INSTALL=go install
GO_CLEAN=go clean

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	$(GO_BUILD) $(LDFLAGS) -o $(BINARY_NAME) .

# Install the application
.PHONY: install
install:
	$(GO_INSTALL) $(LDFLAGS) .

# Run tests
.PHONY: test
test:
	$(GO_TEST) ./...

# Clean build artifacts
.PHONY: clean
clean:
	$(GO_CLEAN)
	rm -f $(BINARY_NAME)

# Show version information
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all      - Build the application (default)"
	@echo "  build    - Build the application"
	@echo "  install  - Install the application"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  version  - Show version information"
	@echo "  help     - Show this help message"