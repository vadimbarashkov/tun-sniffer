APP_NAME := tun-sniffer
SRC_DIR := ./cmd
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BUILD_DIR := ./bin/${GOOS}_${GOARCH}

TUN_IP ?= 10.0.0.1/24
TUN_ROUTE ?= 10.0.0.0/24

.PHONY: all ci build run fmt vet lint test tidy clean docker-build docker-run help

all: build ## Default target: Build the project.

ci: tidy fmt vet lint test build clean ## Run all CI checks.

build: ## Build the project binary.
	@echo "Building the binary for ${GOOS}/${GOARCH}..."
	@mkdir -p "${BUILD_DIR}"
	GOOS="${GOOS}" GOARCH="${GOARCH}" go build -o "${BUILD_DIR}/${APP_NAME}" "${SRC_DIR}"

run: build ## Build and run the application.
	@echo "Running the application..."
	"${BUILD_DIR}/${APP_NAME}" \
		-tunIP=${TUN_IP} \
		-tunRoute=${TUN_ROUTE}

fmt: ## Format code using gofmt.
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

lint: ## Lint the code.
	@echo "Running lint checks..."
	@command -v golangci-lint > /dev/null 2>&1 || { echo "golangci-lint not found"; exit 1; }
	golangci-lint run

test: ## Run unit tests.
	@echo "Running unit tests..."
	go test -cover ./...

tidy: ## Ensure module dependencies are tidy.
	@echo "Tydying up go.mod and go.sum..."
	go mod tidy

clean: ## Remove build files and artifacts.
	@echo "Cleaning up..."
	@if test -d "${BUILD_DIR}"; then rm -rf "${BUILD_DIR}"; fi
	go clean -testcache

docker-build: ## Build an image from a Dockerfile.
	@echo "Building an image..."
	docker build --build-arg TUN_IP=${TUN_IP} --build-arg TUN_ROUTE=${TUN_ROUTE} -t tun-sniffer .

docker-run: docker-build  ## Create and run a new container from an image.
	@echo "Creating a new container..."
	docker run -d --rm --name tun-sniffer \
		--cap-add=NET_ADMIN \
		--device /dev/net/tun \
		-e TUN_IP="${TUN_IP}" \
		-e TUN_ROUTE="${TUN_ROUTE}" \
		tun-sniffer

help: ## Display help for each target.
	@echo "Usage: make [target]"
	@echo
	@echo "Available targets:"
	@grep -E '^[a-zA-Z0-9_/.-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'
	@echo
