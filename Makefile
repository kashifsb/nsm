# NSM Enterprise Development Environment Manager
# Build and release configuration for v1.0.0

# Build configuration
BINARY_NAME=nsm
SETUP_BINARY_NAME=nsm-setup
BUILD_DIR=build
INSTALL_DIR=$(HOME)/bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go configuration
LDFLAGS=-ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME) -s -w"
GOARCH=$(shell go env GOARCH)
GOOS=$(shell go env GOOS)
CGO_ENABLED=0

# Colors for output
BLUE=\033[0;34m
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
PURPLE=\033[0;35m
CYAN=\033[0;36m
NC=\033[0m

# Default target
.DEFAULT_GOAL := help

.PHONY: all build setup-build clean install setup-install install-all uninstall test lint fmt vet deps check release release-local tag test-release help

## Build the main NSM binary
build:
	@echo -e "$(BLUE)Building NSM $(VERSION) for $(GOOS)/$(GOARCH)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/nsm
	@echo -e "$(GREEN)✓ NSM build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## Build the setup binary
setup-build:
	@echo -e "$(BLUE)Building NSM Setup $(VERSION) for $(GOOS)/$(GOARCH)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $(BUILD_DIR)/$(SETUP_BINARY_NAME) ./cmd/setup
	@echo -e "$(GREEN)✓ NSM Setup build complete: $(BUILD_DIR)/$(SETUP_BINARY_NAME)$(NC)"

## Build both binaries
all: clean build setup-build

## Install NSM binary to local bin directory
install: build
	@echo -e "$(BLUE)Installing NSM to $(INSTALL_DIR)...$(NC)"
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@echo -e "$(GREEN)✓ NSM installed to $(INSTALL_DIR)/$(BINARY_NAME)$(NC)"

## Install setup binary to local bin directory
setup-install: setup-build
	@echo -e "$(BLUE)Installing NSM Setup to $(INSTALL_DIR)...$(NC)"
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(SETUP_BINARY_NAME) $(INSTALL_DIR)/
	@chmod +x $(INSTALL_DIR)/$(SETUP_BINARY_NAME)
	@echo -e "$(GREEN)✓ NSM Setup installed to $(INSTALL_DIR)/$(SETUP_BINARY_NAME)$(NC)"

## Install both binaries
install-all: install setup-install
	@echo -e "$(CYAN)Adding $(INSTALL_DIR) to PATH if not already present:$(NC)"
	@echo -e "  export PATH=\"$(INSTALL_DIR):\$$PATH\""
	@echo -e "  echo 'export PATH=\"$(INSTALL_DIR):\$$PATH\"' >> ~/.zshrc"

## Uninstall from local bin directory
uninstall:
	@echo -e "$(BLUE)Uninstalling NSM...$(NC)"
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@rm -f $(INSTALL_DIR)/$(SETUP_BINARY_NAME)
	@echo -e "$(GREEN)✓ NSM uninstalled$(NC)"

## Clean build artifacts
clean:
	@echo -e "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@echo -e "$(GREEN)✓ Clean complete$(NC)"

## Run tests
test:
	@echo -e "$(BLUE)Running tests...$(NC)"
	@go test -v -race ./...
	@echo -e "$(GREEN)✓ Tests passed$(NC)"

## Run linter
lint:
	@echo -e "$(BLUE)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
		echo -e "$(GREEN)✓ Linting passed$(NC)"; \
	else \
		echo -e "$(YELLOW)golangci-lint not found, install with:$(NC)"; \
		echo -e "  $(CYAN)go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
	fi

## Format code
fmt:
	@echo -e "$(BLUE)Formatting code...$(NC)"
	@go fmt ./...
	@echo -e "$(GREEN)✓ Code formatted$(NC)"

## Run go vet
vet:
	@echo -e "$(BLUE)Running go vet...$(NC)"
	@go vet ./...
	@echo -e "$(GREEN)✓ Vet passed$(NC)"

## Download dependencies
deps:
	@echo -e "$(BLUE)Downloading dependencies...$(NC)"
	@go mod download
	@go mod tidy
	@echo -e "$(GREEN)✓ Dependencies updated$(NC)"

## Run all checks
check: fmt vet lint test

## Build release packages for all supported platforms
release: clean
	@echo -e "$(PURPLE)Building NSM v$(VERSION) release packages...$(NC)"
	@mkdir -p $(BUILD_DIR)/release
	
	# Build for macOS
	@echo -e "$(BLUE)Building for macOS (amd64)...$(NC)"
	@mkdir -p $(BUILD_DIR)/release/darwin-amd64
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/darwin-amd64/$(BINARY_NAME) ./cmd/nsm
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/darwin-amd64/$(SETUP_BINARY_NAME) ./cmd/setup
	
	@echo -e "$(BLUE)Building for macOS (arm64)...$(NC)"
	@mkdir -p $(BUILD_DIR)/release/darwin-arm64
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/darwin-arm64/$(BINARY_NAME) ./cmd/nsm
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/darwin-arm64/$(SETUP_BINARY_NAME) ./cmd/setup
	
	# Build for Linux
	@echo -e "$(BLUE)Building for Linux (amd64)...$(NC)"
	@mkdir -p $(BUILD_DIR)/release/linux-amd64
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/linux-amd64/$(BINARY_NAME) ./cmd/nsm
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/linux-amd64/$(SETUP_BINARY_NAME) ./cmd/setup
	
	@echo -e "$(BLUE)Building for Linux (arm64)...$(NC)"
	@mkdir -p $(BUILD_DIR)/release/linux-arm64
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/linux-arm64/$(BINARY_NAME) ./cmd/nsm
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/release/linux-arm64/$(SETUP_BINARY_NAME) ./cmd/setup
	
	# Create documentation and example files for each platform
	@echo -e "$(BLUE)Preparing release packages...$(NC)"
	@for platform in darwin-amd64 darwin-arm64 linux-amd64 linux-arm64; do \
		echo "Packaging $$platform..."; \
		mkdir -p $(BUILD_DIR)/release/$$platform/examples; \
		if [ -f "README.md" ]; then cp README.md $(BUILD_DIR)/release/$$platform/; fi; \
		if [ -f "VERSION.md" ]; then cp VERSION.md $(BUILD_DIR)/release/$$platform/; fi; \
		if [ -f "PLATFORM_SUPPORT.md" ]; then cp PLATFORM_SUPPORT.md $(BUILD_DIR)/release/$$platform/; fi; \
		if [ -d "examples" ]; then cp -r examples/* $(BUILD_DIR)/release/$$platform/examples/ 2>/dev/null || true; fi; \
		chmod +x $(BUILD_DIR)/release/$$platform/$(BINARY_NAME); \
		chmod +x $(BUILD_DIR)/release/$$platform/$(SETUP_BINARY_NAME); \
	done
	
	# Create tarballs
	@echo -e "$(BLUE)Creating release archives...$(NC)"
	@cd $(BUILD_DIR)/release && \
	for platform in darwin-amd64 darwin-arm64 linux-amd64 linux-arm64; do \
		echo "Creating $(BINARY_NAME)-$(VERSION)-$$platform.tar.gz..."; \
		tar -czf $(BINARY_NAME)-$(VERSION)-$$platform.tar.gz $$platform/; \
	done
	
	@echo -e "$(GREEN)✓ Release packages created:$(NC)"
	@ls -la $(BUILD_DIR)/release/*.tar.gz

## Create checksums for release
checksums: release
	@echo -e "$(BLUE)Creating checksums...$(NC)"
	@cd $(BUILD_DIR)/release && \
	for file in *.tar.gz; do \
		shasum -a 256 "$$file" >> checksums.txt; \
	done
	@echo -e "$(GREEN)✓ Checksums created:$(NC)"
	@cat $(BUILD_DIR)/release/checksums.txt

## Install local release for testing
release-local: release
	@echo -e "$(BLUE)Installing local release for testing...$(NC)"
	@platform="$(GOOS)-$(GOARCH)"; \
	cd $(BUILD_DIR)/release && \
	tar -xzf $(BINARY_NAME)-$(VERSION)-$$platform.tar.gz; \
	cp $$platform/$(BINARY_NAME) $(INSTALL_DIR)/; \
	cp $$platform/$(SETUP_BINARY_NAME) $(INSTALL_DIR)/
	@echo -e "$(GREEN)✓ Local release installed to $(INSTALL_DIR)$(NC)"

## Test the release locally
test-release: release-local
	@echo -e "$(BLUE)Testing local release...$(NC)"
	@echo -e "$(CYAN)NSM Version:$(NC)"
	@$(INSTALL_DIR)/$(BINARY_NAME) --version || echo "NSM version check failed"
	@echo -e "$(CYAN)NSM Setup Version:$(NC)"
	@$(INSTALL_DIR)/$(SETUP_BINARY_NAME) --version || echo "NSM Setup version check failed"
	@echo -e "$(CYAN)NSM Help:$(NC)"
	@$(INSTALL_DIR)/$(BINARY_NAME) --help | head -5
	@echo -e "$(GREEN)✓ Release test completed$(NC)"

## Development build with debug symbols
dev:
	@echo -e "$(BLUE)Building development version with debug symbols...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -gcflags="all=-N -l" -ldflags="-X main.version=$(VERSION)-dev -X main.commit=$(COMMIT)" -o $(BUILD_DIR)/$(BINARY_NAME)-dev ./cmd/nsm
	@go build -gcflags="all=-N -l" -ldflags="-X main.version=$(VERSION)-dev -X main.commit=$(COMMIT)" -o $(BUILD_DIR)/$(SETUP_BINARY_NAME)-dev ./cmd/setup
	@echo -e "$(GREEN)✓ Development build complete$(NC)"

## Show version information
version:
	@echo -e "$(PURPLE)NSM Build Information$(NC)"
	@echo -e "$(CYAN)Version:$(NC)     $(VERSION)"
	@echo -e "$(CYAN)Commit:$(NC)      $(COMMIT)"
	@echo -e "$(CYAN)Build Time:$(NC)  $(BUILD_TIME)"
	@echo -e "$(CYAN)Go Version:$(NC)  $(shell go version)"
	@echo -e "$(CYAN)Platform:$(NC)    $(GOOS)/$(GOARCH)"

## Tag version for release
tag:
	@if [ "$(VERSION)" = "v1.0.0" ] || [ "$(VERSION)" = "1.0.0" ]; then \
		echo -e "$(GREEN)Version looks good: $(VERSION)$(NC)"; \
	else \
		echo -e "$(YELLOW)Current version: $(VERSION)$(NC)"; \
		echo -e "$(CYAN)To create v1.0.0 release:$(NC)"; \
		echo -e "  git tag v1.0.0"; \
		echo -e "  git push origin v1.0.0"; \
	fi

## Validate examples directory
validate-examples:
	@echo -e "$(BLUE)Validating examples...$(NC)"
	@if [ ! -d "examples" ]; then \
		echo -e "$(RED)✗ Examples directory not found$(NC)"; \
		exit 1; \
	fi
	@for example in react-vite-typescript go python java rust; do \
		if [ -d "examples/$$example" ]; then \
			echo -e "$(GREEN)✓ Found examples/$$example$(NC)"; \
			if [ -f "examples/$$example/cmd/nsm.sh" ]; then \
				echo -e "  $(GREEN)✓ NSM script present$(NC)"; \
			else \
				echo -e "  $(YELLOW)⚠ NSM script missing$(NC)"; \
			fi; \
		else \
			echo -e "$(YELLOW)⚠ Missing examples/$$example$(NC)"; \
		fi; \
	done

## CI/CD preparation check
ci-check: deps fmt vet lint test validate-examples
	@echo -e "$(GREEN)✓ All CI checks passed$(NC)"

## Full release workflow
release-workflow: clean ci-check release checksums
	@echo -e "$(PURPLE)🎉 Release workflow completed!$(NC)"
	@echo -e "$(CYAN)Next steps:$(NC)"
	@echo -e "  1. Review the release files in $(BUILD_DIR)/release/"
	@echo -e "  2. Test with: make test-release"
	@echo -e "  3. Tag the release: git tag v1.0.0"
	@echo -e "  4. Push tag: git push origin v1.0.0"
	@echo -e "  5. GitHub Actions will create the release"

## Show help
help:
	@echo -e "$(PURPLE)NSM v1.0.0 Build System$(NC)"
	@echo ""
	@echo -e "$(CYAN)Build Commands:$(NC)"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo -e "$(CYAN)Quick Start:$(NC)"
	@echo -e "  make all              # Build both binaries"
	@echo -e "  make install-all      # Install locally"
	@echo -e "  make release-workflow # Full release process"
	@echo ""
	@echo -e "$(CYAN)Environment Variables:$(NC)"
	@echo -e "  INSTALL_DIR    Installation directory (default: $(INSTALL_DIR))"
	@echo -e "  VERSION        Version string (default: auto from git)"
	@echo ""
	@echo -e "$(CYAN)Examples:$(NC)"
	@echo -e "  make INSTALL_DIR=/usr/local/bin install-all"
	@echo -e "  make VERSION=v1.0.0 release"
	@echo ""
	@echo -e "$(PURPLE)Platform Support:$(NC)"
	@echo -e "  ✅ macOS (Full Support)"
	@echo -e "  ⚠️  Linux (Partial Support)"
	@echo -e "  ❌ Windows (Not Supported)"
