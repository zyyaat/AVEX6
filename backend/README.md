# AVEX Backend

Modular Monolith architecture for the AVEX delivery platform backend.

## Overview

Built with Go 1.25 + PostgreSQL 16 + Redis 7, using Clean Architecture (Domain → Port → Service → Repository → Transport) and event-driven communication via the Outbox pattern.

## Modules (12)

| Module | Description | Key Features |
|--------|-------------|--------------|
| **Identity** | User/driver/merchant auth | JWT sessions, bcrypt, 6 tables |
| **Orders** | Order lifecycle | 9-state machine, idempotency, parallel dispatch |
| **Catalog** | Restaurants + menu | Categories, menu items, store hours |
| **Financial** | Wallets + pricing + promotions | Money (int cents), wallets (hold/release/settle), pricing engine, promotions |
| **Dispatch** | Driver matching | Nearest-driver search, Mapbox distance matrix, auto-retry |
| **Realtime** | WebSocket hub | coder/websocket, channel-based pub/sub, 18 event types |
| **Notifications** | Push + SMS + email | Per-user preferences, multi-channel, retry with backoff |
| **Support** | Tickets + messages | 5-state machine, attachments, agent assignment |
| **Permissions** | RBAC | Roles, permissions, wildcard matching, seed data |
| **Settings** | Versioned config | Type-safe settings, revision history, rollback, feature flags |
| **Audit** | Immutable audit log | Append-only, auto-auditing via 26 event subscriptions |
| **System** | Health checks | DB + Redis probes, K8s liveness/readiness, maintenance mode |

## Quick Start

### Prerequisites
- Go 1.25+
- PostgreSQL 16+
- Redis 7+
- Docker + Docker Compose (optional, for containerized setup)

### Option 1: Docker Compose (recommended)

```bash
# Start the full stack (PostgreSQL + Redis + API server + worker)
docker compose up -d

# Check health
curl http://localhost:8080/health

# View logs
docker compose logs -f server

# Stop
docker compose down
```

### Option 2: Local Development

```bash
# 1. Start infrastructure only
docker compose up -d postgres redis

# 2. Copy env file
cp .env.example .env
# Edit .env with your DATABASE_URL, REDIS_URL, JWT_SECRET, MAPBOX_ACCESS_TOKEN

# 3. Run server + worker
make dev

# Or run separately:
make server    # API server on :8080
make worker    # Outbox publisher worker
```

### Common Commands

```bash
make help              # Show all commands
make test              # Run all unit tests (707+ tests)
make test-integration  # Run integration tests (requires PostgreSQL)
make vet               # Run go vet
make fmt               # Format all Go code
make build             # Build server + worker binaries
make docker-up         # Start full stack via Docker
make docker-down       # Stop all containers
make docker-rebuild    # Rebuild + restart app containers
```

## API Endpoints

### Public (no auth)
- `GET /health` — Full system health
- `GET /health/live` — Liveness probe
- `GET /health/ready` — Readiness probe
- `GET /system/info` — Build + runtime info
- `GET /api/v1/restaurants` — List restaurants
- `GET /api/v1/restaurants/{id}/menu` — Restaurant menu
- `GET /api/v1/promotions` — Active promotions
- `GET /api/v1/i18n/languages` — Supported languages
- `GET /api/v1/i18n/translate?lang=ar&key=orders.status.pending` — Translate

### Authenticated (Bearer token)
- `POST /api/v1/auth/register` — Register user
- `POST /api/v1/auth/login` — Login
- `POST /api/v1/orders` — Create order (triggers parallel dispatch)
- `GET /api/v1/ws?token=<JWT>` — WebSocket connection
- `POST /api/v1/pricing/quote` — Calculate delivery quote
- `POST /api/v1/promotions/validate` — Validate promo code
- `POST /api/v1/support/tickets` — Create support ticket

### Admin (Bearer token + role)
- `POST /api/v1/admin/restaurants` — Manage restaurants
- `POST /api/v1/admin/wallets/{id}/credit` — Credit wallet
- `POST /api/v1/admin/drivers` — Register driver
- `POST /api/v1/admin/roles/assign` — Assign role
- `PUT /api/v1/admin/settings/{id}` — Update setting
- `GET /api/v1/admin/audit` — Query audit log

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Transport (HTTP)                  │
│         REST + WebSocket + Health endpoints         │
├─────────────────────────────────────────────────────┤
│                      Service                        │
│    Business logic, use cases, event publishing      │
├─────────────────────────────────────────────────────┤
│                       Port                          │
│     Interfaces (Repository, Service, Providers)     │
├─────────────────────────────────────────────────────┤
│                      Domain                         │
│       Pure entities, value objects, errors          │
│            (zero external dependencies)             │
├─────────────────────────────────────────────────────┤
│                    Repository                        │
│              PostgreSQL (pgx/v5)                     │
├─────────────────────────────────────────────────────┤
│                    Platform                         │
│     Config, DB, Bus (Redis), Outbox, Inbox, etc.    │
└─────────────────────────────────────────────────────┘
```

### Event Flow (Outbox Pattern)

```
Service → Outbox Table (in DB tx) → Worker → Redis Pub/Sub → Subscribers
                                                         ├── Realtime (WebSocket broadcast)
                                                         ├── Notifications (push/SMS/email)
                                                         └── Audit (auto-log)
```

## Testing

```bash
# Unit tests (707+ tests, no external dependencies)
make test

# Integration tests (requires PostgreSQL)
DATABASE_URL=postgres://avex:avex@localhost:5432/avex_test?sslmode=disable \
  make test-integration

# Run with race detector
go test -race ./...
```

## Deployment

### Docker

```bash
# Build
docker build -t avex-backend:latest .

# Run server
docker run -d \
  -e APP_ROLE=server \
  -e DATABASE_URL=postgres://... \
  -e REDIS_URL=redis://... \
  -e JWT_SECRET=... \
  -e MAPBOX_ACCESS_TOKEN=... \
  -p 8080:8080 \
  avex-backend:latest

# Run worker (same image, different role)
docker run -d \
  -e APP_ROLE=worker \
  -e DATABASE_URL=postgres://... \
  -e REDIS_URL=redis://... \
  avex-backend:latest
```

### CI/CD (GitHub Actions)

The `.github/workflows/ci.yml` pipeline runs on every push/PR:

1. **Lint** — go vet + gofmt check
2. **Test** — 707+ unit tests with race detector
3. **Build** — Compile server + worker binaries
4. **Docker** — Build Docker image + smoke test
5. **Integration** — Cross-module integration tests with PostgreSQL service

## Project Structure

```
backend/
├── cmd/
│   ├── server/         # HTTP API server entry point
│   └── worker/         # Outbox publisher worker entry point
├── internal/
│   ├── modules/        # 12 business modules
│   │   ├── identity/
│   │   ├── orders/
│   │   ├── catalog/
│   │   ├── financial/
│   │   ├── dispatch/
│   │   ├── realtime/
│   │   ├── notifications/
│   │   ├── support/
│   │   ├── permissions/
│   │   ├── settings/
│   │   ├── audit/
│   │   ├── system/
│   │   └── localization/
│   ├── integration/    # Cross-module integration tests
│   └── platform/       # Shared infrastructure (config, db, bus, outbox, etc.)
├── migrations/         # Per-module SQL migrations (goose)
├── Dockerfile          # Multi-stage build (server + worker)
├── docker-compose.yml  # Full stack (infra + app)
└── Makefile            # Development commands
```

## Status

- [x] Phase 1: Infrastructure (config, db, bus, outbox, inbox, tracing, crypto)
- [x] Phase 2: Catalog Module (restaurants + menu + categories)
- [x] Phase 3: Financial Module (wallets + pricing + promotions)
- [x] Phase 4: Dispatch Module (driver matching + offers + Mapbox)
- [x] Phase 5: Realtime Module (WebSocket hub + event broadcasting)
- [x] Phase 6: Notifications Module (push + SMS + email + preferences)
- [x] Phase 7: Support Module (tickets + messages + attachments)
- [x] Phase 8: Permissions Module (RBAC + wildcard matching)
- [x] Phase 9: Settings Module (versioned config + feature flags)
- [x] Phase 10: Audit Module (immutable audit log + auto-auditing)
- [x] Phase 11: System Module (health checks + system info)
- [x] Phase 12: Localization Module (multi-language + fallback)
- [x] Phase 13: Integration Tests (11 cross-module tests)
- [x] Phase 14: Deployment (Docker + CI/CD)

**Total: 707 unit tests + 11 integration tests, all passing.**
