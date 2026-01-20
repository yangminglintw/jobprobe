# Makefile for jprobe

.PHONY: all build test clean install lint fmt help docker-build docker-up docker-down docker-test test-local

# Build variables
BINARY_NAME=jprobe
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/user/jobprobe/cmd.Version=$(VERSION) -X github.com/user/jobprobe/cmd.Commit=$(COMMIT) -X github.com/user/jobprobe/cmd.BuildDate=$(BUILD_DATE)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

all: test build

## build: Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) .

## build-all: Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe .

## test: Run tests
test:
	$(GOTEST) -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## clean: Clean build files
clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -f coverage.out coverage.html

## install: Install the binary
install: build
	mv $(BINARY_NAME) $(GOPATH)/bin/

## lint: Run linters
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed" && exit 1)
	golangci-lint run ./...

## fmt: Format code
fmt:
	$(GOFMT) ./...

## tidy: Tidy dependencies
tidy:
	$(GOMOD) tidy

## run: Run with example config
run: build
	./$(BINARY_NAME) run --config configs/ --dry-run

## docker-build: Build Docker images
docker-build:
	docker compose build

## docker-up: Start Docker services
docker-up:
	docker compose up -d mock-api

## docker-down: Stop Docker services
docker-down:
	docker compose down

## docker-test: Run tests in Docker
docker-test: docker-build
	docker compose up -d mock-api
	sleep 3
	docker compose run --rm jprobe run --config /etc/jprobe --verbose
	docker compose down

## test-local: Run local tests with mock API
test-local:
	./scripts/test-local.sh

## mock-api: Build and run mock API locally
mock-api:
	$(GOBUILD) -o test/mock-api/mock-api test/mock-api/main.go
	./test/mock-api/mock-api

## help: Show this help
help:
	@echo "Usage:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
