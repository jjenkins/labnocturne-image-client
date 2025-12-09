M = $(shell printf "\033[34;1m▶\033[0m")
# Go parameters
GOCMD=go
GOBUILD=go build -ldflags="-s -w"
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=ecfr-analyzer
BINARY_UNIX=$(BINARY_NAME)_unix
VERSION=$(shell git log -n 1 --pretty="%h")
COVERAGE_DIR ?= .coverage

# Docker commands with hot reload - recommended for development
dev: ; $(info $(M) Starting Docker development environment with hot reload...)
	docker compose up --build

# Legacy commands (for backwards compatibility)
up: dev
logs: ; $(info $(M) Showing Docker logs...)
	docker compose logs -f
restart: ; $(info $(M) Restarting Docker services...)
	docker compose restart
rebuild: ; $(info $(M) Rebuilding and restarting Docker services...)
	docker compose up -d --build --force-recreate

# Shutdown
down: ; $(info $(M) Shutting down Docker services...)
	docker compose down
down-volumes: ; $(info $(M) Shutting down Docker services and removing volumes...)
	docker compose down -v

# Database access
psql: ; $(info $(M) Connecting to postgres console...)
	docker compose exec postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME)

# Maintenance
clean: ; $(info $(M) Cleaning up Docker resources...)
	docker compose down -v --remove-orphans
	docker system prune -f
