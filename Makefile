.PHONY: build run test lint vet fmt clean update docker-build docker-run help

BINARY    := light-simulator
BUILD_DIR := bin
GO_FILES  := $(shell find . -name '*.go' -not -path './vendor/*')
VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS   := -ldflags "-s -w -X main.version=$(VERSION)"

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/server

run: build ## Build and run the server
	$(BUILD_DIR)/$(BINARY)

dev: ## Run with live reload (requires air)
	air

test: ## Run all tests with race detector and coverage
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out

lint: ## Run golangci-lint
	golangci-lint run ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format all Go files and check for diff
	gofmt -s -w $(GO_FILES)

fmt-check: ## Verify formatting without modifying files
	@test -z "$$(gofmt -s -l $(GO_FILES))" || { echo "Files not formatted:"; gofmt -s -l $(GO_FILES); exit 1; }

update: fmt-check vet lint test build ## Pre-deploy verification: format check, vet, lint, test, build
	@echo ""
	@echo "============================================"
	@echo "  All checks passed. Ready to deploy."
	@echo "============================================"

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR) coverage.out uploads/

docker-build: ## Build Docker image
	docker build -t $(BINARY):$(VERSION) .

docker-run: docker-build ## Run in Docker
	docker run -p 8080:8080 $(BINARY):$(VERSION)
