// Package service is the identity module's service layer.
//
// The Service struct implements port.ServicePort. It holds all
// dependencies (Deps) and a pool Executor for non-transactional reads.
// Each use case method runs business workflows within transactions
// (via TxRunner) and publishes events atomically (via EventPublisher).
//
// Design rules:
//   - No SQL inside service. All persistence goes through port.RepositorySet.
//   - No pgx/Redis imports. Only domain + port + dependencies.
//   - Methods return DTOs (not domain entities) to prevent cross-module
//     domain imports.
//   - Transaction boundaries are explicit: each state-changing use case
//     wraps its writes in TxRunner.RunInTx.
//
// The Service is safe for concurrent use (no mutable state).
package service

import (
	"context"
	"errors"
	"time"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
)

// Config holds service-layer configuration (not infrastructure config).
// These values come from the application config but are scoped to what
// the identity service needs.
type Config struct {
	// AccessTokenTTL is the duration of JWT access tokens.
	AccessTokenTTL time.Duration
}

// Service implements port.ServicePort.
type Service struct {
	deps port.Deps
	pool port.Executor // used for non-transactional reads
	cfg  Config
}

// Compile-time assertion that Service satisfies port.ServicePort.
var _ port.ServicePort = (*Service)(nil)

// New creates a new identity Service.
//
// pool is the Executor used for non-transactional reads (GetUser, etc.).
// It is typically *pgxpool.Pool, but the service only knows it as
// port.Executor (opaque).
func New(deps port.Deps, pool port.Executor, cfg Config) *Service {
	return &Service{
		deps: deps,
		pool: pool,
		cfg:  cfg,
	}
}

// GetMerchantProfile retrieves a merchant's profile by ID.
// Returns ErrMerchantNotFound if not found.
func (s *Service) GetMerchantProfile(ctx context.Context, merchantID string) (*port.MerchantProfileDTO, error) {
	merchant, err := s.deps.Repos.Merchants.GetByID(ctx, s.pool, merchantID)
	if err != nil {
		return nil, err
	}
	dto := toMerchantProfileDTO(*merchant)
	return &dto, nil
}

// Errors that are unused but kept for future use (avoid compile warnings).
var (
	_ = errors.Is
	_ = domain.RoleUser
)
