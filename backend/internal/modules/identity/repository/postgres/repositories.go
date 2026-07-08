// Package postgres implements the identity module's repository interfaces
// using pgx/v5 against a PostgreSQL database.
//
// Design rules (enforced by the port layer):
//   - No business logic. The repository only CRUDs + maps.
//   - No direct pool access inside methods. Every method receives a
//     port.Executor which is converted to database.DBTX via the
//     toDBTX adapter.
//   - No SQL mapping inside domain entities. All row <-> entity
//     conversion lives in mapper.go.
//   - Methods return domain sentinel errors (e.g. ErrUserNotFound)
//     on expected failure paths, wrapped errors on infrastructure
//     failures.
//
// Schema: all tables live in the PostgreSQL schema "identity".
package postgres

import (
	"avex-backend/internal/modules/identity/port"
	"avex-backend/internal/platform/database"
)

// Repositories is the concrete implementation of port.RepositorySet.
// It is constructed once at application startup and shared across
// all goroutines (each method is stateless and safe for concurrent use).
type Repositories struct {
	users          *UsersRepository
	drivers        *DriversRepository
	merchants      *MerchantsRepository
	agents         *AgentsRepository
	sessions       *SessionsRepository
	passwordResets *PasswordResetsRepository
}

// NewRepositories constructs a Repositories backed by the given pgxpool.
// The pool is used only for non-transactional operations; transactional
// operations receive their Executor from the service layer via TxRunner.
//
// The returned *Repositories satisfies port.RepositorySet when assigned
// to it (see RepositorySet() below).
func NewRepositories() *Repositories {
	return &Repositories{
		users:          &UsersRepository{},
		drivers:        &DriversRepository{},
		merchants:      &MerchantsRepository{},
		agents:         &AgentsRepository{},
		sessions:       &SessionsRepository{},
		passwordResets: &PasswordResetsRepository{},
	}
}

// RepositorySet returns a port.RepositorySet backed by this Repositories.
// This is the form the service layer expects (it accesses repos via
// interface fields).
func (r *Repositories) RepositorySet() port.RepositorySet {
	return port.RepositorySet{
		Users:          r.users,
		Drivers:        r.drivers,
		Merchants:      r.merchants,
		Agents:         r.agents,
		Sessions:       r.sessions,
		PasswordResets: r.passwordResets,
	}
}

// toDBTX converts a port.Executor (opaque interface{}) into a
// database.DBTX. The Executor is expected to be either *pgxpool.Pool
// (non-transactional) or pgx.Tx (transactional) — both satisfy
// database.DBTX.
//
// This adapter keeps the port layer pure (no pgx imports) while
// allowing the postgres layer to use driver-native APIs.
//
// Panics if the Executor does not satisfy database.DBTX. This is a
// programming error (wiring mistake in the composition root), not a
// runtime error — fail fast at the first call rather than silently
// degrading.
func toDBTX(exec port.Executor) database.DBTX {
	dbtx, ok := exec.(database.DBTX)
	if !ok {
		panic("postgres: port.Executor does not satisfy database.DBTX — check composition root wiring")
	}
	return dbtx
}
