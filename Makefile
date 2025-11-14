# Makefile for Go Microservices Project
# Note: Complex scripts should be placed in /scripts directory

.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Binary names
BINARY_DIR=bin
SERVICES=user-service auth-service api-gateway

# Docker parameters
DOCKER_REGISTRY?=your-registry
DOCKER_TAG?=latest

# Build info
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD)
VERSION?=v1.0.0

LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

## Development
.PHONY: dev
dev: ## Run service in development mode with hot reload
	@echo "Starting development server with hot reload..."
	@which air > /dev/null || go install github.com/air-verse/air@latest
	air -c .air.toml

.PHONY: run
run: ## Run the main application
	@$(GOCMD) run ./cmd/user-service

## Building
.PHONY: build
build: ## Build all services
	@echo "Building all services..."
	@mkdir -p $(BINARY_DIR)
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$$service ./cmd/$$service; \
	done
	@echo "Build complete!"

.PHONY: build-service
build-service: ## Build specific service (usage: make build-service SERVICE=user-service)
	@echo "Building $(SERVICE)..."
	@mkdir -p $(BINARY_DIR)
	@$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(SERVICE) ./cmd/$(SERVICE)

.PHONY: build-linux
build-linux: ## Build for Linux
	@echo "Building for Linux..."
	@mkdir -p $(BINARY_DIR)
	@for service in $(SERVICES); do \
		GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$$service-linux ./cmd/$$service; \
	done

.PHONY: build-all-platforms
build-all-platforms: ## Build for all platforms
	@echo "Building for all platforms..."
	@./scripts/build-all-platforms.sh

## Testing
.PHONY: test
test: ## Run unit tests
	@echo "Running unit tests..."
	@$(GOTEST) -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@$(GOTEST) -v -tags=integration ./test/...

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	@echo "Running E2E tests..."
	@$(GOTEST) -v -tags=e2e ./test/e2e/...

.PHONY: bench
bench: ## Run benchmarks
	@echo "Running benchmarks..."
	@$(GOTEST) -bench=. -benchmem ./...

## Code Quality
.PHONY: lint
lint: ## Run linter
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run --config .golangci.yml ./...

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	@$(GOFMT) ./...
	@which goimports > /dev/null || go install golang.org/x/tools/cmd/goimports@latest
	@goimports -w .

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@$(GOVET) ./...

.PHONY: check
check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)

## Dependencies
.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GOMOD) download

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	@$(GOMOD) tidy
	@$(GOGET) -u ./...

.PHONY: deps-vendor
deps-vendor: ## Vendor dependencies
	@echo "Vendoring dependencies..."
	@$(GOMOD) vendor

## Proto/gRPC
.PHONY: proto
proto: ## Generate protobuf code
	@echo "Generating protobuf code..."
	@which buf > /dev/null || go install github.com/bufbuild/buf/cmd/buf@latest
	@buf generate

.PHONY: proto-lint
proto-lint: ## Lint proto files
	@echo "Linting proto files..."
	@which buf > /dev/null || go install github.com/bufbuild/buf/cmd/buf@latest
	@buf lint

## Docker
.PHONY: docker-build
docker-build: ## Build Docker images for all services
	@echo "Building Docker images..."
	@for service in $(SERVICES); do \
		echo "Building $$service image..."; \
		docker build -f build/package/$$service/Dockerfile -t $(DOCKER_REGISTRY)/$$service:$(DOCKER_TAG) .; \
	done

.PHONY: docker-build-service
docker-build-service: ## Build Docker image for specific service (usage: make docker-build-service SERVICE=user-service)
	@echo "Building $(SERVICE) image..."
	@docker build -f build/package/$(SERVICE)/Dockerfile -t $(DOCKER_REGISTRY)/$(SERVICE):$(DOCKER_TAG) .

.PHONY: docker-push
docker-push: ## Push Docker images
	@echo "Pushing Docker images..."
	@for service in $(SERVICES); do \
		echo "Pushing $$service image..."; \
		docker push $(DOCKER_REGISTRY)/$$service:$(DOCKER_TAG); \
	done

.PHONY: docker-compose-up
docker-compose-up: ## Start services with docker-compose
	@echo "Starting services with docker-compose..."
	@docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down: ## Stop services with docker-compose
	@echo "Stopping services with docker-compose..."
	@docker-compose down

.PHONY: docker-compose-logs
docker-compose-logs: ## Show docker-compose logs
	@docker-compose logs -f

## Kubernetes
.PHONY: k8s-deploy
k8s-deploy: ## Deploy to Kubernetes
	@echo "Deploying to Kubernetes..."
	@kubectl apply -f deployments/kubernetes/

.PHONY: k8s-delete
k8s-delete: ## Delete from Kubernetes
	@echo "Deleting from Kubernetes..."
	@kubectl delete -f deployments/kubernetes/

.PHONY: k8s-logs
k8s-logs: ## Show Kubernetes logs (usage: make k8s-logs SERVICE=user-service)
	@kubectl logs -f -l app=$(SERVICE)

## Database
.PHONY: db-migrate
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	@./scripts/migrate.sh up

.PHONY: db-migrate-down
db-migrate-down: ## Rollback database migrations
	@echo "Rolling back database migrations..."
	@./scripts/migrate.sh down

.PHONY: db-seed
db-seed: ## Seed database
	@echo "Seeding database..."
	@./scripts/seed.sh

## Security
.PHONY: sec-scan
sec-scan: ## Run security scan
	@echo "Running security scan..."
	@which gosec > /dev/null || go install github.com/securego/gosec/v2/cmd/gosec@latest
	@gosec -fmt=json -out=security-report.json ./...

.PHONY: vuln-check
vuln-check: ## Check for known vulnerabilities
	@echo "Checking for vulnerabilities..."
	@which govulncheck > /dev/null || go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

## Monitoring & Profiling
.PHONY: profile-cpu
profile-cpu: ## Profile CPU usage
	@echo "Profiling CPU..."
	@$(GOTEST) -cpuprofile=cpu.prof -bench=. ./...
	@$(GOCMD) tool pprof cpu.prof

.PHONY: profile-mem
profile-mem: ## Profile memory usage
	@echo "Profiling memory..."
	@$(GOTEST) -memprofile=mem.prof -bench=. ./...
	@$(GOCMD) tool pprof mem.prof

## Cleanup
.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html
	@rm -f *.prof
	@rm -f security-report.json

.PHONY: clean-all
clean-all: clean ## Clean everything including vendor and caches
	@echo "Deep cleaning..."
	@rm -rf vendor/
	@$(GOCMD) clean -modcache

## Documentation
.PHONY: docs
docs: ## Generate documentation
	@echo "Generating documentation..."
	@which godoc > /dev/null || go install golang.org/x/tools/cmd/godoc@latest
	@echo "Documentation server starting at http://localhost:6060"
	@godoc -http=:6060

.PHONY: docs-swagger
docs-swagger: ## Generate Swagger documentation
	@echo "Generating Swagger docs..."
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	@swag init -g cmd/user-service/main.go -o api/swagger

## Tools Installation
.PHONY: install-tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/air-verse/air@latest
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "All tools installed!"

## CI/CD
.PHONY: ci
ci: check test-coverage sec-scan vuln-check ## Run CI pipeline

.PHONY: release
release: ## Create a new release (usage: make release VERSION=v1.0.0)
	@echo "Creating release $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)

.DEFAULT_GOAL := help
