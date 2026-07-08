---
name: AVEX Go backend setup
description: How the Go backend is wired into the api-server artifact for the AVEX food delivery platform
---

# AVEX Go backend setup

## The rule
The api-server artifact runs the Go backend from `backend/cmd/server` directly via pnpm dev script. The Go binary is NOT a Node.js server.

**Why:** The imported Vercel project used a Go backend (not Node.js). The api-server artifact was repurposed to shell out to Go rather than rewriting the entire backend in Express.

## How to apply
- Dev command in `artifacts/api-server/package.json`: `"dev": "cd ../../backend && go run ./cmd/server"`
- Production build: `go build -o ../artifacts/api-server/avex-server ./cmd/server` (CGO_ENABLED=0)
- Health check path: `/api/health` (NOT `/api/healthz`)
- Go reads `PORT` and `DATABASE_URL` from environment (both provided by Replit)
- CORS: AllowCredentials=false with wildcard origins (apps use Authorization header, not cookies)

## Key files
- `backend/cmd/server/main.go` — Go entry point, registers all route groups
- `backend/internal/shared/db.go` — InitDB, createSchema, runMigrations, Seed functions
- `artifacts/api-server/package.json` — dev/build scripts pointing to Go backend
- `artifacts/api-server/.replit-artifact/artifact.toml` — health path = `/api/health`
