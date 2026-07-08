// Package port repository: persistence interfaces for the identity module.
//
// Each repository interface covers one domain entity. Methods accept an
// Executor explicitly — transactions are never hidden in context.
//
// Design rules:
//   - Every method takes (ctx, exec, ...) where exec is either a pool
//     (for non-transactional ops) or a transaction (for atomic ops).
//   - Methods return domain entity pointers on success, nil + sentinel
//     domain error on failure (e.g. ErrUserNotFound).
//   - Create/Update methods accept domain entities (not DTOs).
//   - Repositories do NOT publish events — the service layer calls
//     EventPublisher within the same transaction (passing the same exec).
//   - List methods accept PageQuery and return Page[T] for pagination.
//
// The RepositorySet struct aggregates all repos for dependency injection.
package port

import (
	"context"
	"time"

	"avex-backend/internal/modules/identity/domain"
)

// ===== Pagination =====

// PageQuery holds pagination parameters for list queries.
// Use Normalize() to apply defaults and clamp to maximum before passing
// to a repository method.
type PageQuery struct {
	Limit  int
	Offset int
}

// Pagination defaults.
const (
	DefaultPageLimit = 50
	MaxPageLimit     = 100
)

// Normalize returns a PageQuery with defaults applied and values clamped
// to valid ranges. Always call this before passing PageQuery to a repo.
func (p PageQuery) Normalize() PageQuery {
	if p.Limit <= 0 {
		p.Limit = DefaultPageLimit
	}
	if p.Limit > MaxPageLimit {
		p.Limit = MaxPageLimit
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	return p
}

// Page holds a single page of results plus the total count (for UI paging).
type Page[T any] struct {
	Items  []T
	Total  int64
	Limit  int
	Offset int
}

// HasMore reports whether there are more items beyond this page.
func (p Page[T]) HasMore() bool {
	return int64(p.Offset+p.Limit) < p.Total
}

// NextPage returns the PageQuery for the next page, or the zero value
// if there is no next page.
func (p Page[T]) NextPage() PageQuery {
	if !p.HasMore() {
		return PageQuery{}
	}
	return PageQuery{Limit: p.Limit, Offset: p.Offset + p.Limit}
}

// ===== Repository Interfaces =====

// ----- UserRepository -----

// UserRepository persists User entities (customers).
type UserRepository interface {
	// Create inserts a new user. Returns ErrUserAlreadyExists if the phone
	// is already registered.
	Create(ctx context.Context, exec Executor, user domain.User) error

	// GetByID retrieves a user by ID. Returns ErrUserNotFound if not found.
	GetByID(ctx context.Context, exec Executor, id string) (*domain.User, error)

	// GetByPhone retrieves a user by phone number. Returns ErrUserNotFound
	// if not found. Used for login.
	GetByPhone(ctx context.Context, exec Executor, phone domain.Phone) (*domain.User, error)

	// Update saves all fields of an existing user. The service layer must
	// have loaded the user first (via GetByID) before calling Update.
	Update(ctx context.Context, exec Executor, user domain.User) error

	// Deactivate marks a user as deactivated by setting deactivated_at.
	Deactivate(ctx context.Context, exec Executor, id string, now time.Time) error
}

// ----- DriverRepository -----

// DriverRepository persists Driver entities.
type DriverRepository interface {
	// Create inserts a new driver. Returns ErrDriverAlreadyExists if phone,
	// national_id, or license_number is already registered.
	Create(ctx context.Context, exec Executor, driver domain.Driver) error

	// GetByID retrieves a driver by ID. Returns ErrDriverNotFound if not found.
	GetByID(ctx context.Context, exec Executor, id string) (*domain.Driver, error)

	// GetByPhone retrieves a driver by phone. Returns ErrDriverNotFound if
	// not found. Used for driver login.
	GetByPhone(ctx context.Context, exec Executor, phone domain.Phone) (*domain.Driver, error)

	// Update saves all fields of an existing driver.
	Update(ctx context.Context, exec Executor, driver domain.Driver) error

	// UpdateLocation updates only the driver's location and timestamps.
	// Optimized for high-frequency heartbeat updates (does not touch
	// updated_at or other fields).
	UpdateLocation(ctx context.Context, exec Executor, id string, loc domain.Location, now time.Time) error

	// UpdateStatus updates only the driver's status and related fields.
	// Used by GoOnline/GoOffline/Suspend/Unsuspend flows.
	UpdateStatus(ctx context.Context, exec Executor, id string, status domain.DriverStatus, now time.Time) error

	// GetOnlineDriverIDsInZone returns IDs of online, active, verified
	// drivers whose location is within the given zone. Used by the dispatch
	// module (via ServicePort, not directly).
	// zoneID is a soft reference to financial.delivery_zones.id.
	// staleSeconds is the maximum age of the driver's location update.
	// Returns a capped list (implementation-defined max, e.g. 100 IDs).
	GetOnlineDriverIDsInZone(ctx context.Context, exec Executor, zoneID string, staleSeconds int) ([]string, error)
}

// ----- MerchantRepository -----

// MerchantRepository persists Merchant entities.
type MerchantRepository interface {
	// Create inserts a new merchant. Returns ErrMerchantAlreadyExists if
	// phone or restaurant_id is already linked.
	Create(ctx context.Context, exec Executor, merchant domain.Merchant) error

	// GetByID retrieves a merchant by ID.
	GetByID(ctx context.Context, exec Executor, id string) (*domain.Merchant, error)

	// GetByPhone retrieves a merchant by phone. Used for merchant login.
	GetByPhone(ctx context.Context, exec Executor, phone domain.Phone) (*domain.Merchant, error)

	// GetByRestaurantID retrieves the merchant managing a given restaurant.
	// Returns ErrMerchantNotFound if no merchant is linked.
	GetByRestaurantID(ctx context.Context, exec Executor, restaurantID string) (*domain.Merchant, error)

	// Update saves all fields of an existing merchant.
	Update(ctx context.Context, exec Executor, merchant domain.Merchant) error
}

// ----- AgentRepository -----

// AgentRepository persists SupportAgent entities.
type AgentRepository interface {
	// Create inserts a new support agent. Returns ErrAgentAlreadyExists if
	// phone or email is already registered.
	Create(ctx context.Context, exec Executor, agent domain.SupportAgent) error

	// GetByID retrieves an agent by ID.
	GetByID(ctx context.Context, exec Executor, id string) (*domain.SupportAgent, error)

	// GetByPhone retrieves an agent by phone. Used for agent login.
	GetByPhone(ctx context.Context, exec Executor, phone domain.Phone) (*domain.SupportAgent, error)

	// Update saves all fields of an existing agent.
	Update(ctx context.Context, exec Executor, agent domain.SupportAgent) error
}

// ----- SessionRepository -----

// SessionRepository persists Session entities for JWT revocation.
// List methods use PageQuery/Page for pagination.
type SessionRepository interface {
	// Create inserts a new session.
	Create(ctx context.Context, exec Executor, session domain.Session) error

	// GetByID retrieves a session by its ID (which equals the JWT jti).
	// Returns ErrSessionNotFound if not found.
	GetByID(ctx context.Context, exec Executor, id string) (*domain.Session, error)

	// GetBySubject retrieves a paginated list of sessions for a given
	// subject (user/driver/merchant/agent). Includes both active and
	// revoked sessions (filter by IsActive() on the caller side if needed).
	// Used for "revoke all sessions" on password change or suspend, and
	// for admin dashboards showing active sessions.
	GetBySubject(ctx context.Context, exec Executor, subjectID string, subjectType domain.Role, page PageQuery) (Page[domain.Session], error)

	// CountActiveBySubject returns the number of active (non-revoked,
	// non-expired) sessions for a subject. Cheaper than GetBySubject when
	// only a count is needed (e.g. rate limiting concurrent sessions).
	CountActiveBySubject(ctx context.Context, exec Executor, subjectID string, subjectType domain.Role, now time.Time) (int64, error)

	// Revoke marks a single session as revoked by setting revoked_at.
	// Returns ErrSessionAlreadyRevoked if already revoked.
	Revoke(ctx context.Context, exec Executor, id string, now time.Time) error

	// RevokeAllForSubject revokes all active sessions for a subject.
	// Used on password change, suspend, or deactivate.
	// Returns the number of sessions revoked.
	RevokeAllForSubject(ctx context.Context, exec Executor, subjectID string, subjectType domain.Role, now time.Time) (int64, error)

	// DeleteExpired removes sessions that have expired before the given time.
	// Used by a periodic cleanup job (not in Phase 1).
	// Returns the number of rows deleted.
	DeleteExpired(ctx context.Context, exec Executor, before time.Time) (int64, error)
}

// ----- PasswordResetRepository -----

// PasswordResetRepository persists PasswordReset entities.
// Only token hashes are stored — never the plain tokens.
type PasswordResetRepository interface {
	// Create inserts a new password reset entry.
	Create(ctx context.Context, exec Executor, reset domain.PasswordReset) error

	// GetByTokenHash retrieves a password reset by its token hash.
	// Returns ErrPasswordResetNotFound if not found.
	GetByTokenHash(ctx context.Context, exec Executor, tokenHash string) (*domain.PasswordReset, error)

	// MarkUsed marks a reset as used by setting used_at.
	// Returns ErrPasswordResetAlreadyUsed if already used.
	MarkUsed(ctx context.Context, exec Executor, id string, now time.Time) error

	// DeleteExpired removes password resets that have expired before the
	// given time. Used by a periodic cleanup job.
	// Returns the number of rows deleted.
	DeleteExpired(ctx context.Context, exec Executor, before time.Time) (int64, error)
}

// ===== Aggregate =====

// RepositorySet aggregates all identity repository interfaces.
// The service layer receives this struct and accesses repos via fields.
// Each repo shares the same Executor within a transaction.
type RepositorySet struct {
	Users          UserRepository
	Drivers        DriverRepository
	Merchants      MerchantRepository
	Agents         AgentRepository
	Sessions       SessionRepository
	PasswordResets PasswordResetRepository
}
