// Package port deps: dependency interfaces and the Deps struct.
//
// The Deps struct holds all dependencies the identity service layer needs.
// Each dependency is an interface defined HERE (in port/) — not imported
// from platform/. This is true dependency inversion:
//
//   - port/ defines what it needs (interfaces).
//   - platform/ provides implementations that satisfy those interfaces.
//   - module.go wires the implementations into the Deps struct.
//
// This means:
//   - port/ has zero imports on platform/ packages.
//   - Swapping a platform implementation (e.g. bcrypt -> argon2) requires
//     only changing module.go, not port/ or service/.
//   - Tests can provide mock implementations of each interface.
//
// EventPublisher lives here (not in events.go) because it is a dependency
// contract — the service layer depends on it. The event PAYLOADS (what
// gets published) stay in events.go.
//
// The EventPublisher is STATELESS: it holds no mutable actor/metadata
// state. Each Publish call receives an EventContext that carries the
// actor + metadata explicitly. This makes the publisher safe for concurrent
// use as a singleton (no race conditions from shared mutable state).
//
// Imports: stdlib + domain only. No platform/ imports.
package port

import (
	"context"
	"time"
)

// ===== Infrastructure Dependencies =====

// Clock provides the current time. All service code depends on this
// interface, not on time.Now() directly, for testability.
// Satisfied by: platform/timeutil.RealClock, FixedClock.
type Clock interface {
	Now() time.Time
}

// IDGenerator generates unique IDs (UUIDs).
// Satisfied by: platform/id (wraps google/uuid).
type IDGenerator interface {
	New() string
}

// PasswordHasher abstracts password hashing and verification.
// Satisfied by: platform/crypto.BcryptHasher.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

// JWTClaims is the identity-port-specific JWT claims struct.
// It contains only the fields identity needs:
//   - Subject: the actor's ID (user/driver/merchant/agent)
//   - Role: the actor's role
//   - SessionID: the DB-backed session ID (for revocation)
//   - ExpiresAt: token expiry
//
// The implementation converts between this struct and platform/crypto.Claims.
type JWTClaims struct {
	Subject   string
	Role      string
	SessionID string
	ExpiresAt time.Time
}

// IssueJWTParams holds the parameters for issuing a JWT.
type IssueJWTParams struct {
	Subject   string
	Role      string
	SessionID string
	ExpiresAt time.Time
}

// JWTIssuer abstracts JWT token issuance and verification.
// Satisfied by: an adapter wrapping platform/crypto.HS256Issuer.
// The adapter converts IssueJWTParams/JWTClaims to/from platform/crypto.Claims.
type JWTIssuer interface {
	Issue(ctx context.Context, params IssueJWTParams) (string, error)
	Verify(ctx context.Context, token string) (*JWTClaims, error)
}

// Logger is a minimal logging interface.
// Satisfied by: *slog.Logger (Go 1.21+ stdlib).
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// ===== EventPublisher Dependency =====

// EventContext carries the actor and metadata for a single publish call.
// It is passed explicitly to every EventPublisher method — the publisher
// holds NO mutable state, making it safe for concurrent use as a singleton.
//
// The service layer constructs an EventContext once at the start of a use
// case (from the request context) and passes it to every Publish call
// within that operation.
type EventContext struct {
	Actor    ActorContext
	Metadata EventMetadata
}

// EventPublisher publishes identity events to the outbox within the current
// transaction. It is STATELESS — every Publish call receives:
//   - ctx: request context (cancellation, trace propagation)
//   - exec: the transaction executor (pool or pgx.Tx)
//   - payload: the event-specific snapshot DTO
//   - ec: the EventContext (actor + metadata) for this operation
//
// The implementation:
//  1. Constructs a full bus.EventEnvelope from the payload + EventContext.
//  2. Saves it to identity.outbox via the outbox.Outbox interface, using
//     the provided Executor (which may be a pool or a transaction).
//  3. The outbox publisher worker (cmd/worker) later publishes to Redis.
//
// All methods are transactional — if the surrounding transaction rolls back,
// the event is NOT published (outbox row is discarded with the rollback).
//
// Event payload types (UserRegisteredPayload, etc.) are defined in events.go.
type EventPublisher interface {
	// User events
	PublishUserRegistered(ctx context.Context, exec Executor, payload UserRegisteredPayload, ec EventContext) error
	PublishUserLoggedIn(ctx context.Context, exec Executor, payload UserLoggedInPayload, ec EventContext) error
	PublishUserLoggedOut(ctx context.Context, exec Executor, payload UserLoggedOutPayload, ec EventContext) error
	PublishUserProfileUpdated(ctx context.Context, exec Executor, payload UserProfileUpdatedPayload, ec EventContext) error
	PublishUserPasswordChanged(ctx context.Context, exec Executor, payload UserPasswordChangedPayload, ec EventContext) error

	// Driver events
	PublishDriverRegistered(ctx context.Context, exec Executor, payload DriverRegisteredPayload, ec EventContext) error
	PublishDriverVerified(ctx context.Context, exec Executor, payload DriverVerifiedPayload, ec EventContext) error
	PublishDriverStatusChanged(ctx context.Context, exec Executor, payload DriverStatusChangedPayload, ec EventContext) error
	PublishDriverSuspended(ctx context.Context, exec Executor, payload DriverSuspendedPayload, ec EventContext) error

	// Merchant events
	PublishMerchantRegistered(ctx context.Context, exec Executor, payload MerchantRegisteredPayload, ec EventContext) error
	PublishMerchantVerified(ctx context.Context, exec Executor, payload MerchantVerifiedPayload, ec EventContext) error

	// Agent events
	PublishAgentCreated(ctx context.Context, exec Executor, payload AgentCreatedPayload, ec EventContext) error
}

// ===== Deps Struct =====

// Deps holds all dependencies the identity service layer needs.
// Constructed in module.go (the composition root) and passed to the
// service constructor.
//
// This is a struct (not an interface) because:
//   - All fields are required — no optional deps.
//   - Structs are more idiomatic in Go for dependency injection.
//   - Tests construct a Deps with mock implementations.
type Deps struct {
	Clock          Clock
	IDGenerator    IDGenerator
	PasswordHasher PasswordHasher
	JWTIssuer      JWTIssuer
	EventPublisher EventPublisher
	Logger         Logger
	TxRunner       TxRunner
	Repos          RepositorySet
}

// Moment is a type alias for a function that returns the current time.
// Rarely needed — Clock is preferred. Included for cases where a callable
// is more convenient than an interface (e.g. passing to goroutines).
type Moment func() time.Time
