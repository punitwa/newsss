# News Aggregator Makefile

.PHONY: help build run test clean docker-build docker-run docker-stop docker-clean deps lint

# Variables
DOCKER_COMPOSE = docker-compose
GO = go
GOLINT = golangci-lint

# Default target
help: ## Show this help message
	@echo 'Usage: make <target>'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
deps: ## Install dependencies
	$(GO) mod download
	$(GO) mod tidy

build: ## Build all services
	$(GO) build -o bin/api-gateway ./cmd/api-gateway
	$(GO) build -o bin/data-collector ./cmd/data-collector
	$(GO) build -o bin/processor ./cmd/processor

run-api: ## Run API Gateway locally
	$(GO) run ./cmd/api-gateway

run-collector: ## Run Data Collector locally
	$(GO) run ./cmd/data-collector

run-processor: ## Run Processor locally
	$(GO) run ./cmd/processor

test: ## Run tests
	$(GO) test -v ./...

test-coverage: ## Run tests with coverage
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	$(GOLINT) run

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker operations
docker-build: ## Build Docker images
	$(DOCKER_COMPOSE) build

docker-run: ## Start all services with Docker Compose
	$(DOCKER_COMPOSE) up -d

docker-run-logs: ## Start all services with Docker Compose and show logs
	$(DOCKER_COMPOSE) up

docker-stop: ## Stop all services
	$(DOCKER_COMPOSE) down

docker-restart: ## Restart all services
	$(DOCKER_COMPOSE) restart

docker-clean: ## Stop and remove all containers, networks, and volumes
	$(DOCKER_COMPOSE) down -v --remove-orphans
	docker system prune -f

docker-logs: ## Show logs for all services
	$(DOCKER_COMPOSE) logs -f

docker-logs-api: ## Show logs for API Gateway
	$(DOCKER_COMPOSE) logs -f api-gateway

docker-logs-collector: ## Show logs for Data Collector
	$(DOCKER_COMPOSE) logs -f data-collector

docker-logs-processor: ## Show logs for Processor
	$(DOCKER_COMPOSE) logs -f processor

# Database operations
db-migrate: ## Run database migrations (if implemented)
	@echo "Database migrations not implemented yet"

db-seed: ## Seed database with sample data
	@echo "Database seeding not implemented yet"

db-reset: ## Reset database
	$(DOCKER_COMPOSE) down postgres
	docker volume rm newsaggregator_postgres_data
	$(DOCKER_COMPOSE) up -d postgres

# Monitoring
monitoring-up: ## Start monitoring stack (Prometheus + Grafana)
	$(DOCKER_COMPOSE) up -d prometheus grafana

monitoring-down: ## Stop monitoring stack
	$(DOCKER_COMPOSE) stop prometheus grafana

# Health checks
health-check: ## Check health of all services
	@echo "Checking API Gateway health..."
	@curl -f http://localhost:8080/health || echo "API Gateway is not healthy"
	@echo "\nChecking Prometheus..."
	@curl -f http://localhost:9090/-/healthy || echo "Prometheus is not healthy"
	@echo "\nChecking Grafana..."
	@curl -f http://localhost:3000/api/health || echo "Grafana is not healthy"

# Development helpers
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	make deps
	make docker-run
	@echo "Waiting for services to start..."
	sleep 30
	make health-check

dev-reset: ## Reset development environment
	make docker-clean
	make dev-setup

# Production helpers
prod-deploy: ## Deploy to production (customize as needed)
	@echo "Production deployment not implemented"

prod-backup: ## Backup production data
	@echo "Production backup not implemented"

# API testing
api-test: ## Test API endpoints
	@echo "Testing API endpoints..."
	@curl -X GET http://localhost:8080/health
	@echo "\n"
	@curl -X GET http://localhost:8080/api/v1/news
	@echo "\n"
	@curl -X GET http://localhost:8080/api/v1/categories

# Load testing (if you have tools installed)
load-test: ## Run load tests
	@echo "Load testing not implemented - consider using tools like hey, wrk, or k6"

# Documentation
docs: ## Generate documentation
	@echo "Documentation generation not implemented"

# Security
security-scan: ## Run security scans
	@echo "Security scanning not implemented - consider using tools like gosec"

# Format code
fmt: ## Format Go code
	$(GO) fmt ./...

# Generate code (if using code generation)
generate: ## Run go generate
	$(GO) generate ./...

# Vendor dependencies
vendor: ## Vendor dependencies
	$(GO) mod vendor

# Check for outdated dependencies
deps-check: ## Check for outdated dependencies
	$(GO) list -u -m all
