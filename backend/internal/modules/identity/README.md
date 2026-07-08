# Identity Module

Bounded context for users, drivers, merchants, and support agents — authentication, sessions, and identity-related events.

## Responsibilities

- User registration and login
- Driver login and status management
- Session management (JWT + DB-backed)
- Password hashing (bcrypt) and password reset (hashed tokens)
- Publishing identity events to the outbox

## Phase 1 Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Register a new user |
| POST | `/api/v1/auth/login` | User login |
| POST | `/api/v1/auth/logout` | Logout (revokes session) |
| GET | `/api/v1/auth/me` | Current user profile |
| POST | `/api/v1/auth/change-password` | Change user password |
| POST | `/api/v1/driver/auth/login` | Driver login |
| POST | `/api/v1/admin/drivers/{id}/suspend` | Admin: suspend driver |

## Layer rules (enforced by linter)

- `domain/` — stdlib only, no external imports
- `port/` — domain + stdlib only
- `service/` — domain, port (own + other modules'), platform utilities
- `repository/postgres/` — domain, port, platform/database only
- `transport/http/` — domain, service, port, platform/httperr only (no repository)
- `events/` — domain, port, platform/bus, platform/outbox

## Events published

See `events/types.go` and `events/snapshots.go` (to be implemented).

## DB schema

PostgreSQL schema `identity`. Migrations under `migrations/identity/`.
