APP_NAME ?= tun_cli
SRC_DIR := ./cmd
BUILD_DIR := ./bin

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: all
all: build ## Default target: Build the project.

.PHONY: ci
ci: tidy fmt vet lint test build clean ## Run all CI checks.

.PHONY: build
build: ## Build the project binary.
	@echo "Building the binary for ${GOOS}/${GOARCH}..."
	@mkdir -p "${BUILD_DIR}"
	GOOS="${GOOS}" GOARCH="${GOARCH}" go build -o "${BUILD_DIR}/${APP_NAME}" "${SRC_DIR}"

.PHONY: run
run: build ## Build and run the application.
	@echo "Running the application..."
	sudo "${BUILD_DIR}/${APP_NAME}"

.PHONY: fmt
fmt: ## Format code using gofmt.
	@echo "Formatting code..."
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: lint
lint: ## Lint the code.
	@echo "Running lint checks..."
	@command -v golangci-lint > /dev/null 2>&1 || echo "golangci-lint not found"
	golangci-lint run

.PHONY: test
test: ## Run unit tests.
	@echo "Running unit tests..."
	go test -cover ./...

.PHONY: tidy
tidy: ## Ensure module dependencies are tidy.
	@echo "Tydying up go.mod and go.sum..."
	go mod tidy

.PHONY: clean
clean: ## Remove build files and artifacts.
	@echo "Cleaning up..."
	@if [ -d "${BUILD_DIR}" ]; then rm -rf "${BUILD_DIR}"; fi
	go clean -testcache

.PHONY: help
help: ## Display help for each target.
	@echo "Usage: make [target]"
	@echo
	@echo "Available targets:"
	@grep -E '^[a-zA-Z0-9_/.-]+:.*?## .*$$' "${MAKEFILE_LIST}" | awk 'BEGIN { FS = ":.*?##" }; { printf " %-20s %s\n", $$1, $$2 }'
	@echo
