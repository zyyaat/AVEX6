// Package postgres implements the orders module's repository interfaces
// using pgx/v5 against a PostgreSQL database.
//
// Design rules (enforced by the port layer):
//   - No business logic. The repository only CRUDs + maps.
//   - No direct pool access inside methods. Every method receives a
//     port.Executor which is converted to database.DBTX via the
//     toDBTX adapter.
//   - No SQL mapping inside domain entities. All row <-> entity
//     conversion lives in mapper.go.
//   - Methods return domain sentinel errors (e.g. ErrOrderNotFound)
//     on expected failure paths, wrapped errors on infrastructure
//     failures.
//
// Schema: all tables live in the PostgreSQL schema "orders".
package postgres

import (
	"avex-backend/internal/modules/orders/port"
	"avex-backend/internal/platform/database"
)

// Repositories is the concrete implementation of port.RepositorySet.
// It is constructed once at application startup and shared across
// all goroutines (each method is stateless and safe for concurrent use).
type Repositories struct {
	orders      *OrdersRepository
	items       *OrderItemsRepository
	history     *StatusHistoryRepository
	assignments *AssignmentsRepository
	outbox      *OutboxRepository
}

// NewRepositories constructs a Repositories backed by the given pgxpool.
func NewRepositories() *Repositories {
	return &Repositories{
		orders:      &OrdersRepository{},
		items:       &OrderItemsRepository{},
		history:     &StatusHistoryRepository{},
		assignments: &AssignmentsRepository{},
		outbox:      &OutboxRepository{},
	}
}

// RepositorySet returns a port.RepositorySet backed by this Repositories.
func (r *Repositories) RepositorySet() port.RepositorySet {
	return port.RepositorySet{
		Orders:      r.orders,
		Items:       r.items,
		History:     r.history,
		Assignments: r.assignments,
		Outbox:      r.outbox,
	}
}

// toDBTX converts a port.Executor (opaque interface{}) into a
// database.DBTX. Panics on wiring error (fail fast).
func toDBTX(exec port.Executor) database.DBTX {
	dbtx, ok := exec.(database.DBTX)
	if !ok {
		panic("postgres: port.Executor does not satisfy database.DBTX — check composition root wiring")
	}
	return dbtx
}
