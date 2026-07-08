// Package service sessions: session creation and JWT issuance helpers
// shared across auth use cases (RegisterUser, LoginUser, LoginDriver).
//
// These helpers are NOT part of the ServicePort interface — they are
// internal utilities used by the auth use cases to avoid code duplication.
package service

import (
	"context"
	"fmt"
	"time"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
)

// createSessionAndToken creates a new Session entity, persists it,
// and issues a JWT. Returns the session ID and the signed token.
//
// This must be called within a transaction (the caller passes the exec
// obtained from TxRunner.RunInTx).
func (s *Service) createSessionAndToken(
	ctx context.Context,
	exec port.Executor,
	subjectID string,
	subjectType domain.Role,
	ip string,
	userAgent string,
) (sessionID string, token string, err error) {
	now := s.deps.Clock.Now()
	expiresAt := now.Add(s.cfg.AccessTokenTTL)

	// Generate session ID.
	sessionID = s.deps.IDGenerator.New()

	// Create session entity.
	session, err := domain.NewSession(domain.SessionParams{
		ID:          sessionID,
		SubjectID:   subjectID,
		SubjectType: subjectType,
		IP:          ip,
		UserAgent:   userAgent,
		IssuedAt:    now,
		TTL:         s.cfg.AccessTokenTTL,
	})
	if err != nil {
		return "", "", fmt.Errorf("create session entity: %w", err)
	}

	// Persist session.
	if err := s.deps.Repos.Sessions.Create(ctx, exec, session); err != nil {
		return "", "", fmt.Errorf("persist session: %w", err)
	}

	// Issue JWT.
	token, err = s.deps.JWTIssuer.Issue(ctx, port.IssueJWTParams{
		Subject:   subjectID,
		Role:      subjectType.String(),
		SessionID: sessionID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", "", fmt.Errorf("issue jwt: %w", err)
	}

	return sessionID, token, nil
}

// defaultExpiry returns a default JWT expiry time.
// Currently unused but kept for future use cases that need a default.
func (s *Service) defaultExpiry() time.Time {
	return s.deps.Clock.Now().Add(s.cfg.AccessTokenTTL)
}
