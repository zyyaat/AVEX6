# ============================================================================
# AVEX Backend — Makefile
# Modular Monolith (Identity reference implementation)
# ============================================================================

# Default
.DEFAULT_GOAL := help

# Binary names
SERVER_BIN := bin/server
WORKER_BIN := bin/worker

# Go
GO        := go
GOFLAGS   := -trimpath
LDFLAGS   := -s -w

# Env (loaded from .env if present)
-include .env
export

# Colors
COLOR_RESET := \033[0m
COLOR_CYAN  := \033[36m
COLOR_GREEN := \033[32m

.PHONY: help
help: ## Show this help
	@printf "$(COLOR_CYAN)AVEX Backend — commands$(COLOR_RESET)\n\n"
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n\n"} \
	/^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: dev
dev: ## Run server + worker concurrently (dev only)
	@printf "$(COLOR_GREEN)Starting dev (server + worker)...$(COLOR_RESET)\n"
	@trap 'kill 0' EXIT; \
	$(GO) run ./cmd/server & \
	$(GO) run ./cmd/worker & \
	wait

.PHONY: server
server: ## Run HTTP API server only
	@$(GO) run ./cmd/server

.PHONY: worker
worker: ## Run outbox publisher worker only
	@$(GO) run ./cmd/worker

.PHONY: build
build: ## Build server + worker binaries into bin/
	@mkdir -p bin
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(SERVER_BIN) ./cmd/server
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(WORKER_BIN) ./cmd/worker

.PHONY: test
test: ## Run all tests
	$(GO) test -race -timeout 60s ./...

.PHONY: test-domain
test-domain: ## Run identity domain tests only
	$(GO) test -race ./internal/modules/identity/domain/...

.PHONY: migrate-up
migrate-up: ## Apply all pending migrations
	$(GO) run ./cmd/migrator up

.PHONY: migrate-down
migrate-down: ## Roll back the last migration
	$(GO) run ./cmd/migrator down

.PHONY: migrate-version
migrate-version: ## Show current migration version
	$(GO) run ./cmd/migrator version

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: arch-check
arch-check: ## Run architecture dependency rules check
	@bash scripts/arch-check.sh

.PHONY: tidy
tidy: ## Run go mod tidy
	$(GO) mod tidy

.PHONY: fmt
fmt: ## Format all Go code
	$(GO) fmt ./...

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: clean
clean: ## Remove bin/ and build artifacts
	rm -rf bin/

.PHONY: deps
deps: ## Download dependencies
	$(GO) mod download
