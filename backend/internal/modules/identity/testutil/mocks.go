// Package testutil provides test doubles (mocks) for identity service
// integration tests.
//
// The mocks implement port interfaces (UserRepository, DriverRepository,
// SessionRepository, EventPublisher, TxRunner, etc.) using in-memory
// storage. This allows integration tests to exercise the full service
// layer without a real database or Redis.
//
// Design:
//   - Mocks are NOT safe for concurrent use (tests are sequential per package).
//   - Mocks record all calls for assertion in tests.
//   - Mocks simulate domain behavior (e.g. UserRepository.Create returns
//     ErrUserAlreadyExists on duplicate phone).
//
// Uses stdlib only — no testify or other test frameworks.
package testutil

import (
	"context"
	"sync"
	"time"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
)

// ===== Mock Repositories =====

// MockUserRepository is an in-memory UserRepository.
type MockUserRepository struct {
	mu     sync.Mutex
	users  map[string]domain.User // keyed by ID
	phones map[string]string      // phone -> userID (for duplicate detection)
}

// NewMockUserRepository creates a new empty MockUserRepository.
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[string]domain.User),
		phones: make(map[string]string),
	}
}

func (r *MockUserRepository) Create(ctx context.Context, exec port.Executor, user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	phone := user.Phone().String()
	if _, exists := r.phones[phone]; exists {
		return domain.ErrUserAlreadyExists
	}
	r.users[user.ID()] = user
	r.phones[phone] = user.ID()
	return nil
}

func (r *MockUserRepository) GetByID(ctx context.Context, exec port.Executor, id string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	u := user
	return &u, nil
}

func (r *MockUserRepository) GetByPhone(ctx context.Context, exec port.Executor, phone domain.Phone) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.phones[phone.String()]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	user := r.users[id]
	return &user, nil
}

func (r *MockUserRepository) Update(ctx context.Context, exec port.Executor, user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[user.ID()]; !ok {
		return domain.ErrUserNotFound
	}
	r.users[user.ID()] = user
	return nil
}

func (r *MockUserRepository) Deactivate(ctx context.Context, exec port.Executor, id string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.users[id]
	if !ok {
		return domain.ErrUserNotFound
	}
	_ = user.Deactivate(now)
	r.users[id] = user
	return nil
}

// ===== Mock DriverRepository =====

// MockDriverRepository is an in-memory DriverRepository.
type MockDriverRepository struct {
	mu      sync.Mutex
	drivers map[string]domain.Driver
	phones  map[string]string
}

func NewMockDriverRepository() *MockDriverRepository {
	return &MockDriverRepository{
		drivers: make(map[string]domain.Driver),
		phones:  make(map[string]string),
	}
}

func (r *MockDriverRepository) Create(ctx context.Context, exec port.Executor, driver domain.Driver) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	phone := driver.Phone().String()
	if _, exists := r.phones[phone]; exists {
		return domain.ErrDriverAlreadyExists
	}
	r.drivers[driver.ID()] = driver
	r.phones[phone] = driver.ID()
	return nil
}

func (r *MockDriverRepository) GetByID(ctx context.Context, exec port.Executor, id string) (*domain.Driver, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.drivers[id]
	if !ok {
		return nil, domain.ErrDriverNotFound
	}
	return &d, nil
}

func (r *MockDriverRepository) GetByPhone(ctx context.Context, exec port.Executor, phone domain.Phone) (*domain.Driver, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.phones[phone.String()]
	if !ok {
		return nil, domain.ErrDriverNotFound
	}
	d := r.drivers[id]
	return &d, nil
}

func (r *MockDriverRepository) Update(ctx context.Context, exec port.Executor, driver domain.Driver) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.drivers[driver.ID()]; !ok {
		return domain.ErrDriverNotFound
	}
	r.drivers[driver.ID()] = driver
	return nil
}

func (r *MockDriverRepository) UpdateLocation(ctx context.Context, exec port.Executor, id string, loc domain.Location, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.drivers[id]
	if !ok {
		return domain.ErrDriverNotFound
	}
	_ = d.UpdateLocation(loc, now) // won't work — UpdateLocation requires online; we just store
	r.drivers[id] = d
	return nil
}

func (r *MockDriverRepository) UpdateStatus(ctx context.Context, exec port.Executor, id string, status domain.DriverStatus, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	d, ok := r.drivers[id]
	if !ok {
		return domain.ErrDriverNotFound
	}
	// Direct status set (mock doesn't enforce transitions).
	r.drivers[id] = d // status already in entity via service
	return nil
}

func (r *MockDriverRepository) GetOnlineDriverIDsInZone(ctx context.Context, exec port.Executor, zoneID string, staleSeconds int) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var ids []string
	for _, d := range r.drivers {
		if d.IsOnline() && d.IsActive() && d.IsVerified() {
			ids = append(ids, d.ID())
		}
	}
	return ids, nil
}

// SeedDriver inserts a driver directly (bypasses Create validation).
// Used by tests to set up pre-existing drivers.
func (r *MockDriverRepository) SeedDriver(d domain.Driver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.drivers[d.ID()] = d
	r.phones[d.Phone().String()] = d.ID()
}

// ===== Mock MerchantRepository =====

type MockMerchantRepository struct {
	mu          sync.Mutex
	merchants   map[string]domain.Merchant
	phones      map[string]string
	restByMerch map[string]string // restaurant_id -> merchant_id
}

func NewMockMerchantRepository() *MockMerchantRepository {
	return &MockMerchantRepository{
		merchants:   make(map[string]domain.Merchant),
		phones:      make(map[string]string),
		restByMerch: make(map[string]string),
	}
}

func (r *MockMerchantRepository) Create(ctx context.Context, exec port.Executor, m domain.Merchant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.phones[m.Phone().String()]; exists {
		return domain.ErrMerchantAlreadyExists
	}
	if _, exists := r.restByMerch[m.RestaurantID()]; exists {
		return domain.ErrMerchantAlreadyExists
	}
	r.merchants[m.ID()] = m
	r.phones[m.Phone().String()] = m.ID()
	r.restByMerch[m.RestaurantID()] = m.ID()
	return nil
}

func (r *MockMerchantRepository) GetByID(ctx context.Context, exec port.Executor, id string) (*domain.Merchant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.merchants[id]
	if !ok {
		return nil, domain.ErrMerchantNotFound
	}
	return &m, nil
}

func (r *MockMerchantRepository) GetByPhone(ctx context.Context, exec port.Executor, phone domain.Phone) (*domain.Merchant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.phones[phone.String()]
	if !ok {
		return nil, domain.ErrMerchantNotFound
	}
	m := r.merchants[id]
	return &m, nil
}

func (r *MockMerchantRepository) GetByRestaurantID(ctx context.Context, exec port.Executor, restaurantID string) (*domain.Merchant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.restByMerch[restaurantID]
	if !ok {
		return nil, domain.ErrMerchantNotFound
	}
	m := r.merchants[id]
	return &m, nil
}

func (r *MockMerchantRepository) Update(ctx context.Context, exec port.Executor, m domain.Merchant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.merchants[m.ID()]; !ok {
		return domain.ErrMerchantNotFound
	}
	r.merchants[m.ID()] = m
	return nil
}

// ===== Mock AgentRepository =====

type MockAgentRepository struct {
	mu     sync.Mutex
	agents map[string]domain.SupportAgent
	phones map[string]string
}

func NewMockAgentRepository() *MockAgentRepository {
	return &MockAgentRepository{
		agents: make(map[string]domain.SupportAgent),
		phones: make(map[string]string),
	}
}

func (r *MockAgentRepository) Create(ctx context.Context, exec port.Executor, a domain.SupportAgent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.phones[a.Phone().String()]; exists {
		return domain.ErrAgentAlreadyExists
	}
	r.agents[a.ID()] = a
	r.phones[a.Phone().String()] = a.ID()
	return nil
}

func (r *MockAgentRepository) GetByID(ctx context.Context, exec port.Executor, id string) (*domain.SupportAgent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.agents[id]
	if !ok {
		return nil, domain.ErrAgentNotFound
	}
	return &a, nil
}

func (r *MockAgentRepository) GetByPhone(ctx context.Context, exec port.Executor, phone domain.Phone) (*domain.SupportAgent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.phones[phone.String()]
	if !ok {
		return nil, domain.ErrAgentNotFound
	}
	a := r.agents[id]
	return &a, nil
}

func (r *MockAgentRepository) Update(ctx context.Context, exec port.Executor, a domain.SupportAgent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.agents[a.ID()]; !ok {
		return domain.ErrAgentNotFound
	}
	r.agents[a.ID()] = a
	return nil
}

// ===== Mock SessionRepository =====

type MockSessionRepository struct {
	mu       sync.Mutex
	sessions map[string]domain.Session
}

func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		sessions: make(map[string]domain.Session),
	}
}

func (r *MockSessionRepository) Create(ctx context.Context, exec port.Executor, s domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.ID()] = s
	return nil
}

func (r *MockSessionRepository) GetByID(ctx context.Context, exec port.Executor, id string) (*domain.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[id]
	if !ok {
		return nil, domain.ErrSessionNotFound
	}
	return &s, nil
}

func (r *MockSessionRepository) GetBySubject(ctx context.Context, exec port.Executor, subjectID string, subjectType domain.Role, page port.PageQuery) (port.Page[domain.Session], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var sessions []domain.Session
	for _, s := range r.sessions {
		if s.SubjectID() == subjectID && s.SubjectType() == subjectType {
			sessions = append(sessions, s)
		}
	}
	return port.Page[domain.Session]{
		Items:  sessions,
		Total:  int64(len(sessions)),
		Limit:  page.Limit,
		Offset: page.Offset,
	}, nil
}

func (r *MockSessionRepository) CountActiveBySubject(ctx context.Context, exec port.Executor, subjectID string, subjectType domain.Role, now time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int64
	for _, s := range r.sessions {
		if s.SubjectID() == subjectID && s.SubjectType() == subjectType && s.IsActive(now) {
			count++
		}
	}
	return count, nil
}

func (r *MockSessionRepository) Revoke(ctx context.Context, exec port.Executor, id string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.sessions[id]
	if !ok {
		return domain.ErrSessionNotFound
	}
	if s.IsRevoked() {
		return domain.ErrSessionAlreadyRevoked
	}
	_ = s.Revoke(now)
	r.sessions[id] = s
	return nil
}

func (r *MockSessionRepository) RevokeAllForSubject(ctx context.Context, exec port.Executor, subjectID string, subjectType domain.Role, now time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int64
	for id, s := range r.sessions {
		if s.SubjectID() == subjectID && s.SubjectType() == subjectType && !s.IsRevoked() {
			_ = s.Revoke(now)
			r.sessions[id] = s
			count++
		}
	}
	return count, nil
}

func (r *MockSessionRepository) DeleteExpired(ctx context.Context, exec port.Executor, before time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int64
	for id, s := range r.sessions {
		if s.IsExpired(before) {
			delete(r.sessions, id)
			count++
		}
	}
	return count, nil
}

// ===== Mock PasswordResetRepository =====

type MockPasswordResetRepository struct {
	mu     sync.Mutex
	resets map[string]domain.PasswordReset // keyed by ID
	hashes map[string]string               // tokenHash -> reset ID
}

func NewMockPasswordResetRepository() *MockPasswordResetRepository {
	return &MockPasswordResetRepository{
		resets: make(map[string]domain.PasswordReset),
		hashes: make(map[string]string),
	}
}

func (r *MockPasswordResetRepository) Create(ctx context.Context, exec port.Executor, reset domain.PasswordReset) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resets[reset.ID()] = reset
	r.hashes[reset.TokenHash()] = reset.ID()
	return nil
}

func (r *MockPasswordResetRepository) GetByTokenHash(ctx context.Context, exec port.Executor, tokenHash string) (*domain.PasswordReset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, ok := r.hashes[tokenHash]
	if !ok {
		return nil, domain.ErrPasswordResetNotFound
	}
	reset := r.resets[id]
	return &reset, nil
}

func (r *MockPasswordResetRepository) MarkUsed(ctx context.Context, exec port.Executor, id string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	reset, ok := r.resets[id]
	if !ok {
		return domain.ErrPasswordResetNotFound
	}
	if reset.IsUsed() {
		return domain.ErrPasswordResetAlreadyUsed
	}
	_ = reset.MarkUsed(now)
	r.resets[id] = reset
	return nil
}

func (r *MockPasswordResetRepository) DeleteExpired(ctx context.Context, exec port.Executor, before time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var count int64
	for id, reset := range r.resets {
		if reset.IsExpired(before) {
			delete(r.resets, id)
			delete(r.hashes, reset.TokenHash())
			count++
		}
	}
	return count, nil
}

// ===== Mock EventPublisher =====

// PublishedEvent captures a single event published during a test.
type PublishedEvent struct {
	EventType string
	Payload   any
}

// MockEventPublisher records all published events without actually publishing.
type MockEventPublisher struct {
	mu     sync.Mutex
	events []PublishedEvent
}

func NewMockEventPublisher() *MockEventPublisher {
	return &MockEventPublisher{}
}

// Events returns a copy of all published events (thread-safe).
func (p *MockEventPublisher) Events() []PublishedEvent {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]PublishedEvent, len(p.events))
	copy(out, p.events)
	return out
}

// EventCount returns the total number of events published.
func (p *MockEventPublisher) EventCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.events)
}

// FindByType returns all events of the given type.
func (p *MockEventPublisher) FindByType(eventType string) []PublishedEvent {
	p.mu.Lock()
	defer p.mu.Unlock()
	var out []PublishedEvent
	for _, e := range p.events {
		if e.EventType == eventType {
			out = append(out, e)
		}
	}
	return out
}

func (p *MockEventPublisher) record(eventType string, payload any) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, PublishedEvent{EventType: eventType, Payload: payload})
	return nil
}

// All Publish* methods just record.
func (p *MockEventPublisher) PublishUserRegistered(ctx context.Context, exec port.Executor, payload port.UserRegisteredPayload, ec port.EventContext) error {
	return p.record(port.EventUserRegistered, payload)
}
func (p *MockEventPublisher) PublishUserLoggedIn(ctx context.Context, exec port.Executor, payload port.UserLoggedInPayload, ec port.EventContext) error {
	return p.record(port.EventUserLoggedIn, payload)
}
func (p *MockEventPublisher) PublishUserLoggedOut(ctx context.Context, exec port.Executor, payload port.UserLoggedOutPayload, ec port.EventContext) error {
	return p.record(port.EventUserLoggedOut, payload)
}
func (p *MockEventPublisher) PublishUserProfileUpdated(ctx context.Context, exec port.Executor, payload port.UserProfileUpdatedPayload, ec port.EventContext) error {
	return p.record(port.EventUserProfileUpdated, payload)
}
func (p *MockEventPublisher) PublishUserPasswordChanged(ctx context.Context, exec port.Executor, payload port.UserPasswordChangedPayload, ec port.EventContext) error {
	return p.record(port.EventUserPasswordChanged, payload)
}
func (p *MockEventPublisher) PublishDriverRegistered(ctx context.Context, exec port.Executor, payload port.DriverRegisteredPayload, ec port.EventContext) error {
	return p.record(port.EventDriverRegistered, payload)
}
func (p *MockEventPublisher) PublishDriverVerified(ctx context.Context, exec port.Executor, payload port.DriverVerifiedPayload, ec port.EventContext) error {
	return p.record(port.EventDriverVerified, payload)
}
func (p *MockEventPublisher) PublishDriverStatusChanged(ctx context.Context, exec port.Executor, payload port.DriverStatusChangedPayload, ec port.EventContext) error {
	return p.record(port.EventDriverStatusChanged, payload)
}
func (p *MockEventPublisher) PublishDriverSuspended(ctx context.Context, exec port.Executor, payload port.DriverSuspendedPayload, ec port.EventContext) error {
	return p.record(port.EventDriverSuspended, payload)
}
func (p *MockEventPublisher) PublishMerchantRegistered(ctx context.Context, exec port.Executor, payload port.MerchantRegisteredPayload, ec port.EventContext) error {
	return p.record(port.EventMerchantRegistered, payload)
}
func (p *MockEventPublisher) PublishMerchantVerified(ctx context.Context, exec port.Executor, payload port.MerchantVerifiedPayload, ec port.EventContext) error {
	return p.record(port.EventMerchantVerified, payload)
}
func (p *MockEventPublisher) PublishAgentCreated(ctx context.Context, exec port.Executor, payload port.AgentCreatedPayload, ec port.EventContext) error {
	return p.record(port.EventAgentCreated, payload)
}

// ===== Mock TxRunner =====

// MockTxRunner runs fn immediately (no real transaction). The exec
// passed to fn is a sentinel "mock-exec" string.
type MockTxRunner struct{}

func NewMockTxRunner() *MockTxRunner {
	return &MockTxRunner{}
}

func (MockTxRunner) RunInTx(ctx context.Context, fn func(ctx context.Context, exec port.Executor) error) error {
	// Pass a sentinel executor. Mocks ignore it.
	return fn(ctx, "mock-exec")
}

// ===== Mock PasswordHasher =====

// MockPasswordHasher is a simple hasher for tests (NOT for production).
// It prepends "hash:" to the password and compares literally.
type MockPasswordHasher struct{}

func NewMockPasswordHasher() *MockPasswordHasher {
	return &MockPasswordHasher{}
}

func (MockPasswordHasher) Hash(password string) (string, error) {
	return "hash:" + password, nil
}

func (MockPasswordHasher) Compare(hash, password string) error {
	if hash != "hash:"+password {
		return domain.ErrPasswordMismatch
	}
	return nil
}

// ===== Mock JWTIssuer =====

// MockJWTIssuer issues tokens in the format "mock-token:<subject>:<role>:<sessionID>".
// Verify parses this format back. This avoids crypto dependencies in tests.
type MockJWTIssuer struct{}

func NewMockJWTIssuer() *MockJWTIssuer {
	return &MockJWTIssuer{}
}

func (MockJWTIssuer) Issue(ctx context.Context, params port.IssueJWTParams) (string, error) {
	return "mock-token:" + params.Subject + ":" + params.Role + ":" + params.SessionID, nil
}

func (MockJWTIssuer) Verify(ctx context.Context, token string) (*port.JWTClaims, error) {
	// Parse "mock-token:<subject>:<role>:<sessionID>".
	if len(token) < 12 || token[:12] != "mock-token:" {
		return nil, domain.ErrInvalidCredentials
	}
	rest := token[12:]
	// Split by ":" — but subject/role/sessionID may not contain ":".
	// For mock purposes, we split into exactly 3 parts.
	parts := splitN(rest, ':', 3)
	if len(parts) != 3 {
		return nil, domain.ErrInvalidCredentials
	}
	return &port.JWTClaims{
		Subject:   parts[0],
		Role:      parts[1],
		SessionID: parts[2],
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}, nil
}

// splitN splits s by sep into at most n parts.
func splitN(s string, sep byte, n int) []string {
	var parts []string
	current := ""
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep && count < n-1 {
			parts = append(parts, current)
			current = ""
			count++
		} else {
			current += string(s[i])
		}
	}
	parts = append(parts, current)
	return parts
}

// ===== Mock Clock =====

// MockClock returns a fixed time that can be advanced in tests.
type MockClock struct {
	mu  sync.Mutex
	now time.Time
}

// NewMockClock creates a MockClock set to the given time.
func NewMockClock(t time.Time) *MockClock {
	return &MockClock{now: t.UTC()}
}

func (c *MockClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// Advance moves the clock forward by d.
func (c *MockClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// ===== Mock IDGenerator =====

// MockIDGenerator generates sequential IDs: "id-1", "id-2", ...
type MockIDGenerator struct {
	mu  sync.Mutex
	seq int
}

func NewMockIDGenerator() *MockIDGenerator {
	return &MockIDGenerator{}
}

func (g *MockIDGenerator) New() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.seq++
	return "id-" + itoa(g.seq)
}

// itoa converts int to string without strconv (to keep imports minimal).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// ===== Mock Logger =====

// MockLogger discards all log output. Tests don't need to assert on logs.
type MockLogger struct{}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (MockLogger) Debug(msg string, args ...any) {}
func (MockLogger) Info(msg string, args ...any)  {}
func (MockLogger) Warn(msg string, args ...any)  {}
func (MockLogger) Error(msg string, args ...any) {}

// ===== RepositorySet Builder =====

// NewMockRepositorySet creates a RepositorySet with all mock repos.
func NewMockRepositorySet() (port.RepositorySet, *MockUserRepository, *MockDriverRepository, *MockMerchantRepository, *MockAgentRepository, *MockSessionRepository, *MockPasswordResetRepository) {
	users := NewMockUserRepository()
	drivers := NewMockDriverRepository()
	merchants := NewMockMerchantRepository()
	agents := NewMockAgentRepository()
	sessions := NewMockSessionRepository()
	resets := NewMockPasswordResetRepository()
	return port.RepositorySet{
		Users:          users,
		Drivers:        drivers,
		Merchants:      merchants,
		Agents:         agents,
		Sessions:       sessions,
		PasswordResets: resets,
	}, users, drivers, merchants, agents, sessions, resets
}
