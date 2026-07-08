// Package postgres sessions: SessionRepository implementation.
//
// Sessions are used for JWT revocation tracking. The session ID equals
// the JWT jti claim. On every authenticated request, the middleware
// verifies that the session exists, is not revoked, and is not expired.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
)

// SessionsRepository implements port.SessionRepository using pgx/v5.
type SessionsRepository struct{}

// Compile-time assertion.
var _ port.SessionRepository = (*SessionsRepository)(nil)

// sessionColumns is the canonical column list for SELECT queries.
// Order MUST match scanSession() in mapper.go.
const sessionColumns = `
	id, subject_id, subject_type, issued_at, expires_at,
	ip, user_agent, revoked_at, created_at
`

// Create inserts a new session.
func (r *SessionsRepository) Create(ctx context.Context, exec port.Executor, session domain.Session) error {
	dbtx := toDBTX(exec)
	_, err := dbtx.Exec(ctx, `
		INSERT INTO identity.sessions (
			id, subject_id, subject_type, issued_at, expires_at,
			ip, user_agent, revoked_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, sessionInsertArgs(session)...)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// GetByID retrieves a session by its ID (which equals the JWT jti).
// Returns domain.ErrSessionNotFound if not found.
func (r *SessionsRepository) GetByID(ctx context.Context, exec port.Executor, id string) (*domain.Session, error) {
	dbtx := toDBTX(exec)
	row := dbtx.QueryRow(ctx, `
		SELECT `+sessionColumns+` FROM identity.sessions WHERE id = $1
	`, id)
	session, err := scanSession(row)
	if err != nil {
		return nil, mapSessionReadError(err)
	}
	return &session, nil
}

// GetBySubject retrieves a paginated list of sessions for a given subject.
// Includes both active and revoked sessions. The caller filters by
// IsActive() if needed.
func (r *SessionsRepository) GetBySubject(ctx context.Context, exec port.Executor, subjectID string, subjectType domain.Role, page port.PageQuery) (port.Page[domain.Session], error) {
	page = page.Normalize()
	dbtx := toDBTX(exec)

	// Count total sessions for the subject.
	var total int64
	err := dbtx.QueryRow(ctx, `
		SELECT COUNT(*) FROM identity.sessions
		WHERE subject_id = $1 AND subject_type = $2
	`, subjectID, subjectType.String()).Scan(&total)
	if err != nil {
		return port.Page[domain.Session]{}, fmt.Errorf("count sessions: %w", err)
	}

	// Fetch the page.
	rows, err := dbtx.Query(ctx, `
		SELECT `+sessionColumns+` FROM identity.sessions
		WHERE subject_id = $1 AND subject_type = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, subjectID, subjectType.String(), page.Limit, page.Offset)
	if err != nil {
		return port.Page[domain.Session]{}, fmt.Errorf("query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return port.Page[domain.Session]{}, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return port.Page[domain.Session]{}, fmt.Errorf("iterate sessions: %w", err)
	}

	return port.Page[domain.Session]{
		Items:  sessions,
		Total:  total,
		Limit:  page.Limit,
		Offset: page.Offset,
	}, nil
}

// CountActiveBySubject returns the number of active (non-revoked,
// non-expired) sessions for a subject.
func (r *SessionsRepository) CountActiveBySubject(ctx context.Context, exec port.Executor, subjectID string, subjectType domain.Role, now time.Time) (int64, error) {
	dbtx := toDBTX(exec)
	var count int64
	err := dbtx.QueryRow(ctx, `
		SELECT COUNT(*) FROM identity.sessions
		WHERE subject_id = $1
		  AND subject_type = $2
		  AND revoked_at IS NULL
		  AND expires_at > $3
	`, subjectID, subjectType.String(), now).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active sessions: %w", err)
	}
	return count, nil
}

// Revoke marks a single session as revoked.
// Returns domain.ErrSessionAlreadyRevoked if already revoked.
// Returns domain.ErrSessionNotFound if the session does not exist.
func (r *SessionsRepository) Revoke(ctx context.Context, exec port.Executor, id string, now time.Time) error {
	dbtx := toDBTX(exec)
	ct, err := dbtx.Exec(ctx, `
		UPDATE identity.sessions
		SET revoked_at = $1
		WHERE id = $2 AND revoked_at IS NULL
	`, now, id)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if ct.RowsAffected() == 0 {
		// Either the session doesn't exist OR it's already revoked.
		// Distinguish by checking existence.
		var exists bool
		err := dbtx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM identity.sessions WHERE id = $1)
		`, id).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check session existence: %w", err)
		}
		if !exists {
			return domain.ErrSessionNotFound
		}
		return domain.ErrSessionAlreadyRevoked
	}
	return nil
}

// RevokeAllForSubject revokes all active sessions for a subject.
// Returns the number of sessions revoked.
func (r *SessionsRepository) RevokeAllForSubject(ctx context.Context, exec port.Executor, subjectID string, subjectType domain.Role, now time.Time) (int64, error) {
	dbtx := toDBTX(exec)
	ct, err := dbtx.Exec(ctx, `
		UPDATE identity.sessions
		SET revoked_at = $1
		WHERE subject_id = $2
		  AND subject_type = $3
		  AND revoked_at IS NULL
	`, now, subjectID, subjectType.String())
	if err != nil {
		return 0, fmt.Errorf("revoke all sessions: %w", err)
	}
	return ct.RowsAffected(), nil
}

// DeleteExpired removes sessions that have expired before the given time.
// Returns the number of rows deleted.
func (r *SessionsRepository) DeleteExpired(ctx context.Context, exec port.Executor, before time.Time) (int64, error) {
	dbtx := toDBTX(exec)
	ct, err := dbtx.Exec(ctx, `
		DELETE FROM identity.sessions WHERE expires_at < $1
	`, before)
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}
	return ct.RowsAffected(), nil
}

// mapSessionReadError converts pgx read errors to domain sentinel errors.
func mapSessionReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrSessionNotFound
	}
	return fmt.Errorf("session read: %w", err)
}
