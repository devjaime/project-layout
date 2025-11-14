# Getting Started with Microservices Project

This guide will help you set up and run the microservices project locally.

## Prerequisites

### Required
- **Go 1.23+** - [Install Go](https://golang.org/doc/install)
- **Docker** - [Install Docker](https://docs.docker.com/get-docker/)
- **Docker Compose** - [Install Docker Compose](https://docs.docker.com/compose/install/)
- **Make** - Usually pre-installed on Linux/macOS

### Optional (for development)
- **kubectl** - [Install kubectl](https://kubernetes.io/docs/tasks/tools/)
- **buf** - [Install buf](https://buf.build/docs/installation)
- **golangci-lint** - [Install golangci-lint](https://golangci-lint.run/usage/install/)
- **grpcurl** - [Install grpcurl](https://github.com/fullstorydev/grpcurl)

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/golang-standards/project-layout.git
cd project-layout
```

### 2. Install Development Tools

```bash
make install-tools
```

This will install:
- golangci-lint (linter)
- air (hot reload)
- buf (protobuf tooling)
- protoc plugins
- security scanners

### 3. Set Up Environment

```bash
cp .env.example .env
# Edit .env with your configuration if needed
```

### 4. Start Dependencies

```bash
# Start PostgreSQL
docker-compose up -d postgres

# Wait for database to be ready
sleep 5
```

### 5. Download Dependencies

```bash
make deps
```

### 6. Generate Protobuf Code

```bash
make proto
```

### 7. Run the Service

#### Option A: Direct Run
```bash
make run
```

#### Option B: Hot Reload (Recommended for Development)
```bash
make dev
```

### 8. Verify Service is Running

```bash
# Health check
curl http://localhost:8080/health

# Version info
curl http://localhost:8080/version
```

## Development Workflow

### Running Tests

```bash
# Unit tests
make test

# Tests with coverage
make test-coverage

# Integration tests
make test-integration

# All tests
make check
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint

# Run all checks
make check
```

### Building

```bash
# Build all services
make build

# Build specific service
make build-service SERVICE=user-service

# Build for Linux
make build-linux
```

### Working with Proto Files

```bash
# Generate code from proto files
make proto

# Lint proto files
make proto-lint
```

## Docker Usage

### Build Docker Image

```bash
# Build all services
make docker-build

# Build specific service
make docker-build-service SERVICE=user-service
```

### Run with Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f user-service

# Stop all services
docker-compose down
```

### Optional Services

```bash
# Start with pgAdmin
docker-compose --profile tools up -d

# Start with monitoring (Prometheus + Grafana)
docker-compose --profile monitoring up -d
```

Access:
- **pgAdmin**: http://localhost:5050 (admin@admin.com / admin)
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin / admin)

## Kubernetes Deployment

### Prerequisites
- Kubernetes cluster (minikube, kind, or cloud provider)
- kubectl configured

### Deploy to Kubernetes

```bash
# Deploy PostgreSQL
kubectl apply -f deployments/kubernetes/postgres-deployment.yaml

# Deploy User Service
kubectl apply -f deployments/kubernetes/user-service-deployment.yaml

# Check status
kubectl get pods
kubectl get services
```

### Access Service

```bash
# Port forward to access locally
kubectl port-forward svc/user-service 50051:50051 8080:8080

# Or use LoadBalancer/Ingress in production
```

## Testing the Service

### Using grpcurl

```bash
# Install grpcurl if not already installed
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List services
grpcurl -plaintext localhost:50051 list

# Create a user
grpcurl -plaintext -d '{
  "email": "user@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890"
}' localhost:50051 user.v1.UserService/CreateUser

# Get user by ID
grpcurl -plaintext -d '{
  "id": "user-id-here"
}' localhost:50051 user.v1.UserService/GetUser

# List users
grpcurl -plaintext -d '{
  "page": 1,
  "page_size": 10
}' localhost:50051 user.v1.UserService/ListUsers
```

### Using HTTP Health Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Readiness check
curl http://localhost:8080/ready

# Version info
curl http://localhost:8080/version

# Metrics
curl http://localhost:8080/metrics
```

## Common Tasks

### Add a New Service

1. Create directory structure:
   ```bash
   mkdir -p cmd/new-service
   mkdir -p internal/app/new-service/{handler,service,repository,model}
   ```

2. Create proto file:
   ```bash
   mkdir -p api/proto/new-service/v1
   # Add your .proto file
   ```

3. Generate code:
   ```bash
   make proto
   ```

4. Implement service logic

5. Add to Makefile `SERVICES` variable

6. Create Dockerfile in `build/package/new-service/`

### Update Dependencies

```bash
# Update all dependencies
make deps-update

# Vendor dependencies (optional)
make deps-vendor
```

### Database Migrations

```bash
# Run migrations
make db-migrate

# Rollback migrations
make db-migrate-down

# Seed database
make db-seed
```

## Troubleshooting

### Port Already in Use

```bash
# Find process using port
lsof -i :50051
lsof -i :8080

# Kill process
kill -9 <PID>
```

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# View PostgreSQL logs
docker-compose logs postgres

# Restart PostgreSQL
docker-compose restart postgres
```

### Proto Generation Fails

```bash
# Ensure buf is installed
which buf

# Check buf.yaml configuration
buf lint

# Reinstall protoc plugins
make install-tools
```

### Build Fails

```bash
# Clean build artifacts
make clean

# Clean everything
make clean-all

# Redownload dependencies
make deps
```

## Configuration

### Environment Variables

The service reads configuration from:
1. `configs/config.yaml` (default values)
2. Environment variables (override config file)
3. `.env` file (loaded in development)

Priority: **Environment Variables** > **Config File** > **Defaults**

### Key Configuration Options

```bash
# Server
APP_SERVER_GRPC_PORT=50051
APP_SERVER_HTTP_PORT=8080

# Database
APP_DATABASE_HOST=localhost
APP_DATABASE_PORT=5432
APP_DATABASE_USER=postgres
APP_DATABASE_PASSWORD=postgres
APP_DATABASE_DATABASE=users

# Logging
APP_LOGGER_LEVEL=info  # debug, info, warn, error
APP_LOGGER_FORMAT=json # json, console
```

## Next Steps

1. **Read Documentation**
   - [Microservices Architecture](MICROSERVICES_ARCHITECTURE.md)
   - [API Documentation](../api/README.md)
   - Project READMEs in each directory

2. **Explore the Code**
   - Start with `cmd/user-service/main.go`
   - Review service layer in `internal/app/user-service/`
   - Check shared packages in `internal/pkg/`

3. **Add Features**
   - Implement authentication
   - Add new endpoints
   - Create additional services

4. **Deploy**
   - Set up CI/CD pipeline
   - Deploy to staging environment
   - Configure monitoring and alerts

## Useful Makefile Commands

```bash
# See all available commands
make help

# Development
make dev          # Run with hot reload
make run          # Run service
make build        # Build all services

# Testing
make test         # Unit tests
make test-coverage # Tests with coverage
make bench        # Benchmarks

# Quality
make lint         # Run linter
make fmt          # Format code
make check        # Run all checks

# Docker
make docker-build       # Build images
make docker-compose-up  # Start with compose

# Kubernetes
make k8s-deploy   # Deploy to K8s
make k8s-delete   # Delete from K8s

# Security
make sec-scan     # Security scan
make vuln-check   # Vulnerability check

# Cleanup
make clean        # Clean build artifacts
make clean-all    # Deep clean
```

## Getting Help

- **Documentation**: Check `docs/` directory
- **Issues**: Create an issue on GitHub
- **Discussions**: Use GitHub Discussions
- **Community**: Join our community chat

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

---

Happy coding! 🚀
