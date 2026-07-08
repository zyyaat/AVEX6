// Package port tx: transaction abstraction for the identity module.
//
// Transactions are EXPLICIT, never hidden in context. The service layer
// obtains an Executor from TxRunner.RunInTx and passes it to every
// repository and EventPublisher method that participates in the transaction.
//
// Why explicit over context-based:
//   - Type safety: the compiler enforces that every repo call receives an
//     Executor — no silent fallback to the pool inside a "transaction".
//   - Readability: transaction boundaries are visible at the call site.
//   - Testability: tests pass a mock Executor without touching context.
//   - No hidden state: context is for request-scoped values (correlation
//     IDs, trace IDs), not for infrastructure handles.
//
// Executor is intentionally opaque (empty interface). The port layer must
// not import database driver types. Repository implementations type-assert
// the Executor to their driver-specific DBTX interface (e.g. pgxpool.Pool
// or pgx.Tx). This keeps the port layer pure while allowing the postgres
// layer to use driver-native APIs.
//
// For non-transactional operations (single read/write), the caller passes
// a pool Executor directly — obtained from the composition root (module.go).
package port

import "context"

// Executor is an opaque handle to either a database connection pool or an
// active transaction. Repository and EventPublisher methods accept it
// explicitly so transaction boundaries are visible at the call site.
//
// The port layer does not define the Executor's methods (Exec, Query,
// QueryRow) because those would require importing database driver types.
// Repository implementations type-assert it to their driver-specific
// interface (e.g. database.DBTX satisfied by *pgxpool.Pool and pgx.Tx).
type Executor interface{}

// TxRunner executes a function within a database transaction.
// The function receives an Executor that it passes to repository and
// EventPublisher methods.
//
// Semantics:
//   - If fn returns nil, the transaction is committed.
//   - If fn returns a non-nil error, the transaction is rolled back.
//   - The Executor passed to fn is valid only for the duration of fn.
//   - Nested RunInTx calls on the same TxRunner may reuse the outer
//     transaction (savepoint) or start a new one — implementation-defined.
type TxRunner interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context, exec Executor) error) error
}
