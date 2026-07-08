// Package postgres agents: AgentRepository implementation for SupportAgent entities.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
)

// AgentsRepository implements port.AgentRepository using pgx/v5.
type AgentsRepository struct{}

// Compile-time assertion.
var _ port.AgentRepository = (*AgentsRepository)(nil)

// agentColumns is the canonical column list for SELECT queries.
// Order MUST match scanAgent() in mapper.go.
const agentColumns = `
	id, name, phone, email, password_hash,
	is_active, must_change_password, last_login,
	locale, timezone, created_at, updated_at
`

// Create inserts a new support agent.
// Returns domain.ErrAgentAlreadyExists if phone or email is already registered.
func (r *AgentsRepository) Create(ctx context.Context, exec port.Executor, agent domain.SupportAgent) error {
	dbtx := toDBTX(exec)
	_, err := dbtx.Exec(ctx, `
		INSERT INTO identity.support_agents (
			id, name, phone, email, password_hash,
			is_active, must_change_password, last_login,
			locale, timezone, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, agentInsertArgs(agent)...)
	if err != nil {
		return mapAgentWriteError(err)
	}
	return nil
}

// GetByID retrieves an agent by ID.
func (r *AgentsRepository) GetByID(ctx context.Context, exec port.Executor, id string) (*domain.SupportAgent, error) {
	dbtx := toDBTX(exec)
	row := dbtx.QueryRow(ctx, `
		SELECT `+agentColumns+` FROM identity.support_agents WHERE id = $1
	`, id)
	agent, err := scanAgent(row)
	if err != nil {
		return nil, mapAgentReadError(err)
	}
	return &agent, nil
}

// GetByPhone retrieves an agent by phone.
func (r *AgentsRepository) GetByPhone(ctx context.Context, exec port.Executor, phone domain.Phone) (*domain.SupportAgent, error) {
	dbtx := toDBTX(exec)
	row := dbtx.QueryRow(ctx, `
		SELECT `+agentColumns+` FROM identity.support_agents WHERE phone = $1
	`, phone.String())
	agent, err := scanAgent(row)
	if err != nil {
		return nil, mapAgentReadError(err)
	}
	return &agent, nil
}

// Update saves all fields of an existing agent.
func (r *AgentsRepository) Update(ctx context.Context, exec port.Executor, agent domain.SupportAgent) error {
	dbtx := toDBTX(exec)
	ct, err := dbtx.Exec(ctx, `
		UPDATE identity.support_agents SET
			name = $1, phone = $2, email = $3, password_hash = $4,
			is_active = $5, must_change_password = $6, last_login = $7,
			locale = $8, timezone = $9, updated_at = $10
		WHERE id = $11
	`, agentUpdateArgs(agent)...)
	if err != nil {
		return mapAgentWriteError(err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrAgentNotFound
	}
	return nil
}

// mapAgentReadError converts pgx read errors to domain sentinel errors.
func mapAgentReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrAgentNotFound
	}
	return fmt.Errorf("agent read: %w", err)
}

// mapAgentWriteError converts pgx write errors to domain sentinel errors.
// Unique violations on phone or email are both mapped to ErrAgentAlreadyExists.
func mapAgentWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" { // unique_violation
			return domain.ErrAgentAlreadyExists
		}
	}
	return fmt.Errorf("agent write: %w", err)
}
