# AVEX Backend

Modular Monolith architecture for the AVEX delivery platform.

> **Status**: Reference implementation — Identity module only.

## Architecture

This codebase follows a **Modular Monolith** architecture with strict module boundaries.

### Layered structure

```
backend/
├── cmd/                          # entry points
│   ├── server/                   #   HTTP API server
│   └── worker/                   #   outbox publisher worker
├── internal/
│   ├── platform/                 # shared kernel (config, db, bus, outbox, ...)
│   ├── modules/                  # bounded contexts (identity, ...)
│   │   └── identity/             #   reference implementation
│   └── api/                      # HTTP layer (router, middleware)
└── migrations/                   # versioned SQL migrations (per-module)
    └── identity/
```

### Module structure (per bounded context)

```
modules/<name>/
├── domain/            # pure entities (stdlib only)
├── port/              # interfaces (contracts)
├── service/           # use cases (business logic)
├── repository/postgres/   # persistence (raw pgx/v5)
├── transport/http/    # HTTP handlers
├── events/            # event publishers + consumers
├── jobs/              # scheduled jobs
└── module.go          # composition root
```

### Key decisions

| Decision | Choice | ADR |
|----------|--------|-----|
| HTTP Router | `net/http` stdlib (Go 1.22+) | ADR-001 |
| DB Driver | `pgx/v5` + pgxpool | ADR-002 |
| Migrations | `pressly/goose/v3` (embedded, per-module) | ADR-003 |
| Event Bus | Redis Pub/Sub | ADR-004 |
| Outbox | per-module `<module>.outbox` table + worker | ADR-005 |
| JWT | HS256 behind `JWTIssuer` interface (RS256 later) | ADR-006 |
| Sessions | JWT + DB-backed `sessions` table (revocable) | ADR-007 |
| Password Reset Tokens | stored as `token_hash` only | ADR-008 |
| PII | plaintext now, schema ready for future encryption | ADR-009 |

### Module dependency rules (enforced)

- `domain/` — stdlib only
- `transport/` — never touches `repository/` directly
- `repository/` — never imports another module
- Modules communicate via `port/` interfaces + events only

## Getting started

```bash
# 1. Copy env file
cp .env.example .env
# Edit .env with real values

# 2. Run migrations
make migrate-up

# 3. Run dev (server + worker)
make dev
```

## Phase 1 endpoints (Identity)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Register a new user |
| POST | `/api/v1/auth/login` | User login |
| POST | `/api/v1/auth/logout` | Logout (revokes session) |
| GET | `/api/v1/auth/me` | Current user profile |
| POST | `/api/v1/auth/change-password` | Change user password |
| POST | `/api/v1/driver/auth/login` | Driver login |
| POST | `/api/v1/admin/drivers/{id}/suspend` | Admin: suspend driver |

## Status

- [x] Folder structure scaffolded
- [x] Migrations placeholders (identity)
- [ ] Platform implementations (config, db, bus, outbox, ...)
- [ ] Identity domain + ports + service + repository + transport
- [ ] Domain tests
- [ ] CMD entry points
- [ ] Documentation (ADRs)
