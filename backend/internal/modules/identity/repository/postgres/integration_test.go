// Package postgres integration tests: real PostgreSQL integration tests.
//
// These tests require a running PostgreSQL instance. They are gated behind
// the "integration" build tag — run with:
//
//   DATABASE_URL=postgres://avex@localhost:5432/avex_test?sslmode=disable \
//   go test -tags=integration -race -count=1 ./internal/modules/identity/repository/postgres/...
//
//go:build integration

package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
	"avex-backend/internal/platform/database"
	migrations "avex-backend/migrations"
)

var testDB *pgxpool.Pool

func TestMain(m *testing.M) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL not set — skipping integration tests")
		os.Exit(0)
	}

	ctx := context.Background()
	if err := database.RunUp(ctx, dsn, migrations.IdentityMigrations, "identity", "identity"); err != nil {
		fmt.Fprintf(os.Stderr, "migrations failed: %v\n", err)
		os.Exit(1)
	}

	cfg, _ := pgxpool.ParseConfig(dsn)
	cfg.MaxConns = 5
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create pool: %v\n", err)
		os.Exit(1)
	}
	testDB = pool

	code := m.Run()
	pool.Close()
	os.Exit(code)
}

func cleanupTables(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	_, err := testDB.Exec(ctx, `
		TRUNCATE identity.users, identity.drivers, identity.merchants,
		         identity.support_agents, identity.sessions,
		         identity.password_resets, identity.outbox, identity.inbox
		CASCADE
	`)
	if err != nil {
		t.Fatalf("cleanup: %v", err)
	}
}

var fixedTime = time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

// uuidStr generates a new UUID string for test entities.
func uuidStr() string { return uuid.NewString() }

// ===== User Repository Tests =====

func TestUserRepo_CreateAndGetByID(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &UsersRepository{}
	id := uuidStr()

	user, err := domain.NewUser(domain.UserParams{
		ID: id, Name: "Ahmed Ali", Phone: "01012345678",
		Email: "ahmed@example.com", PasswordHash: "$2a$12$hash", Now: fixedTime,
	})
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}

	if err := repo.Create(ctx, testDB, user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, testDB, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name() != "Ahmed Ali" {
		t.Errorf("Name = %q", got.Name())
	}
	if got.Phone().String() != "01012345678" {
		t.Errorf("Phone = %q", got.Phone())
	}
}

func TestUserRepo_GetByPhone(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &UsersRepository{}
	id := uuidStr()

	user, _ := domain.NewUser(domain.UserParams{
		ID: id, Name: "Test", Phone: "01012345678", PasswordHash: "hash", Now: fixedTime,
	})
	_ = repo.Create(ctx, testDB, user)

	phone, _ := domain.NewPhone("01012345678")
	got, err := repo.GetByPhone(ctx, testDB, phone)
	if err != nil {
		t.Fatalf("GetByPhone: %v", err)
	}
	if got.ID() != id {
		t.Errorf("ID = %q, want %q", got.ID(), id)
	}
}

func TestUserRepo_DuplicatePhone(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &UsersRepository{}

	user1, _ := domain.NewUser(domain.UserParams{
		ID: uuidStr(), Name: "First", Phone: "01012345678", PasswordHash: "hash", Now: fixedTime,
	})
	_ = repo.Create(ctx, testDB, user1)

	user2, _ := domain.NewUser(domain.UserParams{
		ID: uuidStr(), Name: "Second", Phone: "01012345678", PasswordHash: "hash", Now: fixedTime,
	})
	err := repo.Create(ctx, testDB, user2)
	if !errors.Is(err, domain.ErrUserAlreadyExists) {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestUserRepo_GetByID_NotFound(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &UsersRepository{}

	_, err := repo.GetByID(ctx, testDB, uuidStr())
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepo_Update(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &UsersRepository{}
	id := uuidStr()

	user, _ := domain.NewUser(domain.UserParams{
		ID: id, Name: "Original", Phone: "01012345678", PasswordHash: "hash", Now: fixedTime,
	})
	_ = repo.Create(ctx, testDB, user)

	_ = user.ChangePassword("$2a$12$newhash", fixedTime.Add(time.Hour))
	if err := repo.Update(ctx, testDB, user); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := repo.GetByID(ctx, testDB, id)
	if got.PasswordHash() != "$2a$12$newhash" {
		t.Errorf("PasswordHash = %q", got.PasswordHash())
	}
}

func TestUserRepo_Deactivate(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &UsersRepository{}
	id := uuidStr()

	user, _ := domain.NewUser(domain.UserParams{
		ID: id, Name: "Test", Phone: "01012345678", PasswordHash: "hash", Now: fixedTime,
	})
	_ = repo.Create(ctx, testDB, user)

	if err := repo.Deactivate(ctx, testDB, id, fixedTime.Add(time.Hour)); err != nil {
		t.Fatalf("Deactivate: %v", err)
	}

	got, _ := repo.GetByID(ctx, testDB, id)
	if got.IsActive() {
		t.Error("user should be deactivated")
	}
}

// ===== Driver Repository Tests =====

func TestDriverRepo_CreateAndGet(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &DriversRepository{}
	id := uuidStr()

	driver, err := domain.NewDriver(domain.DriverParams{
		ID: id, Name: "Test Driver", Phone: "01112345678", PasswordHash: "hash",
		VehicleType: domain.VehicleTypeMotorcycle, LicenseNumber: "LIC-001", NationalID: "NID-001",
		Now: fixedTime,
	})
	if err != nil {
		t.Fatalf("NewDriver: %v", err)
	}

	if err := repo.Create(ctx, testDB, driver); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, testDB, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name() != "Test Driver" {
		t.Errorf("Name = %q", got.Name())
	}
	if got.Status() != domain.DriverStatusOffline {
		t.Errorf("Status = %q", got.Status())
	}
}

func TestDriverRepo_DuplicatePhone(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &DriversRepository{}

	d1, _ := domain.NewDriver(domain.DriverParams{
		ID: uuidStr(), Name: "First", Phone: "01112345678", PasswordHash: "h",
		VehicleType: domain.VehicleTypeMotorcycle, LicenseNumber: "L1", NationalID: "N1", Now: fixedTime,
	})
	_ = repo.Create(ctx, testDB, d1)

	d2, _ := domain.NewDriver(domain.DriverParams{
		ID: uuidStr(), Name: "Second", Phone: "01112345678", PasswordHash: "h",
		VehicleType: domain.VehicleTypeMotorcycle, LicenseNumber: "L2", NationalID: "N2", Now: fixedTime,
	})
	err := repo.Create(ctx, testDB, d2)
	if !errors.Is(err, domain.ErrDriverAlreadyExists) {
		t.Errorf("expected ErrDriverAlreadyExists, got %v", err)
	}
}

func TestDriverRepo_DuplicateLicense(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &DriversRepository{}

	d1, _ := domain.NewDriver(domain.DriverParams{
		ID: uuidStr(), Name: "First", Phone: "01112345678", PasswordHash: "h",
		VehicleType: domain.VehicleTypeMotorcycle, LicenseNumber: "SAME-LIC", NationalID: "N1", Now: fixedTime,
	})
	_ = repo.Create(ctx, testDB, d1)

	d2, _ := domain.NewDriver(domain.DriverParams{
		ID: uuidStr(), Name: "Second", Phone: "01112345679", PasswordHash: "h",
		VehicleType: domain.VehicleTypeMotorcycle, LicenseNumber: "SAME-LIC", NationalID: "N2", Now: fixedTime,
	})
	err := repo.Create(ctx, testDB, d2)
	if !errors.Is(err, domain.ErrDriverAlreadyExists) {
		t.Errorf("expected ErrDriverAlreadyExists for duplicate license, got %v", err)
	}
}

func TestDriverRepo_UpdateStatus(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &DriversRepository{}
	id := uuidStr()

	driver, _ := domain.NewDriver(domain.DriverParams{
		ID: id, Name: "Test", Phone: "01112345678", PasswordHash: "h",
		VehicleType: domain.VehicleTypeMotorcycle, LicenseNumber: "L", NationalID: "N", Now: fixedTime,
	})
	_ = repo.Create(ctx, testDB, driver)

	err := repo.UpdateStatus(ctx, testDB, id, domain.DriverStatusOnline, fixedTime)
	if err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	got, _ := repo.GetByID(ctx, testDB, id)
	if got.Status() != domain.DriverStatusOnline {
		t.Errorf("Status = %q, want 'online'", got.Status())
	}
}

func TestDriverRepo_UpdateLocation(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &DriversRepository{}
	id := uuidStr()

	driver, _ := domain.NewDriver(domain.DriverParams{
		ID: id, Name: "Test", Phone: "01112345678", PasswordHash: "h",
		VehicleType: domain.VehicleTypeMotorcycle, LicenseNumber: "L", NationalID: "N", Now: fixedTime,
	})
	_ = repo.Create(ctx, testDB, driver)

	loc := domain.Location{Lat: 30.05, Lng: 31.36}
	err := repo.UpdateLocation(ctx, testDB, id, loc, fixedTime)
	if err != nil {
		t.Fatalf("UpdateLocation: %v", err)
	}

	got, _ := repo.GetByID(ctx, testDB, id)
	if got.Location().Lat != 30.05 || got.Location().Lng != 31.36 {
		t.Errorf("Location = %v, want {30.05, 31.36}", got.Location())
	}
}

// ===== Session Repository Tests =====

func TestSessionRepo_CreateAndGet(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &SessionsRepository{}
	sid := uuidStr()
	uid := uuidStr()

	// Create a user first (for FK via subject_id — though it's soft ref, the column is UUID).
	session, _ := domain.NewSession(domain.SessionParams{
		ID: sid, SubjectID: uid, SubjectType: domain.RoleUser,
		IP: "1.2.3.4", UserAgent: "test", IssuedAt: fixedTime, TTL: 24 * time.Hour,
	})

	if err := repo.Create(ctx, testDB, session); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, testDB, sid)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.SubjectID() != uid {
		t.Errorf("SubjectID = %q", got.SubjectID())
	}
}

func TestSessionRepo_Revoke(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &SessionsRepository{}
	sid := uuidStr()

	session, _ := domain.NewSession(domain.SessionParams{
		ID: sid, SubjectID: uuidStr(), SubjectType: domain.RoleUser,
		IssuedAt: fixedTime, TTL: 24 * time.Hour,
	})
	_ = repo.Create(ctx, testDB, session)

	if err := repo.Revoke(ctx, testDB, sid, fixedTime.Add(time.Hour)); err != nil {
		t.Fatalf("Revoke: %v", err)
	}

	got, _ := repo.GetByID(ctx, testDB, sid)
	if !got.IsRevoked() {
		t.Error("session should be revoked")
	}
}

func TestSessionRepo_RevokeAlreadyRevoked(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &SessionsRepository{}
	sid := uuidStr()

	session, _ := domain.NewSession(domain.SessionParams{
		ID: sid, SubjectID: uuidStr(), SubjectType: domain.RoleUser,
		IssuedAt: fixedTime, TTL: 24 * time.Hour,
	})
	_ = repo.Create(ctx, testDB, session)
	_ = repo.Revoke(ctx, testDB, sid, fixedTime)

	err := repo.Revoke(ctx, testDB, sid, fixedTime.Add(time.Hour))
	if !errors.Is(err, domain.ErrSessionAlreadyRevoked) {
		t.Errorf("expected ErrSessionAlreadyRevoked, got %v", err)
	}
}

func TestSessionRepo_RevokeAllForSubject(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &SessionsRepository{}
	uid := uuidStr()

	for i := 0; i < 3; i++ {
		s, _ := domain.NewSession(domain.SessionParams{
			ID: uuidStr(), SubjectID: uid, SubjectType: domain.RoleUser,
			IssuedAt: fixedTime, TTL: 24 * time.Hour,
		})
		_ = repo.Create(ctx, testDB, s)
	}

	count, err := repo.RevokeAllForSubject(ctx, testDB, uid, domain.RoleUser, fixedTime.Add(time.Hour))
	if err != nil {
		t.Fatalf("RevokeAllForSubject: %v", err)
	}
	if count != 3 {
		t.Errorf("revoked count = %d, want 3", count)
	}

	page, _ := repo.GetBySubject(ctx, testDB, uid, domain.RoleUser, port.PageQuery{Limit: 100})
	for _, s := range page.Items {
		if !s.IsRevoked() {
			t.Error("session should be revoked")
		}
	}
}

func TestSessionRepo_CountActiveBySubject(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &SessionsRepository{}
	uid := uuidStr()

	// 2 active + 1 revoked.
	for i := 0; i < 2; i++ {
		s, _ := domain.NewSession(domain.SessionParams{
			ID: uuidStr(), SubjectID: uid, SubjectType: domain.RoleUser,
			IssuedAt: fixedTime, TTL: 24 * time.Hour,
		})
		_ = repo.Create(ctx, testDB, s)
	}
	revoked, _ := domain.NewSession(domain.SessionParams{
		ID: uuidStr(), SubjectID: uid, SubjectType: domain.RoleUser,
		IssuedAt: fixedTime, TTL: 24 * time.Hour,
	})
	_ = repo.Create(ctx, testDB, revoked)
	_ = repo.Revoke(ctx, testDB, revoked.ID(), fixedTime)

	count, err := repo.CountActiveBySubject(ctx, testDB, uid, domain.RoleUser, fixedTime)
	if err != nil {
		t.Fatalf("CountActiveBySubject: %v", err)
	}
	if count != 2 {
		t.Errorf("active count = %d, want 2", count)
	}
}

// ===== Transaction Rollback Test =====

func TestTransaction_Rollback(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &UsersRepository{}
	id := uuidStr()

	tx, err := testDB.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}

	user, _ := domain.NewUser(domain.UserParams{
		ID: id, Name: "TxTest", Phone: "01012345678", PasswordHash: "hash", Now: fixedTime,
	})
	if err := repo.Create(ctx, tx, user); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("Create in tx: %v", err)
	}

	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	_, err = repo.GetByID(ctx, testDB, id)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound after rollback, got %v", err)
	}
}

func TestTransaction_Commit(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &UsersRepository{}
	id := uuidStr()

	tx, err := testDB.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin: %v", err)
	}

	user, _ := domain.NewUser(domain.UserParams{
		ID: id, Name: "Commit", Phone: "01012345678", PasswordHash: "hash", Now: fixedTime,
	})
	_ = repo.Create(ctx, tx, user)

	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	got, err := repo.GetByID(ctx, testDB, id)
	if err != nil {
		t.Fatalf("GetByID after commit: %v", err)
	}
	if got.Name() != "Commit" {
		t.Errorf("Name = %q", got.Name())
	}
}

// ===== Password Reset Repository Tests =====

func TestPasswordResetRepo_CreateAndGetByTokenHash(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &PasswordResetsRepository{}
	userRepo := &UsersRepository{}
	uid := uuidStr()

	user, _ := domain.NewUser(domain.UserParams{
		ID: uid, Name: "Test", Phone: "01012345678", PasswordHash: "hash", Now: fixedTime,
	})
	_ = userRepo.Create(ctx, testDB, user)

	reset, _ := domain.NewPasswordReset(domain.PasswordResetParams{
		ID: uuidStr(), UserID: uid, TokenHash: "hash-of-token", Now: fixedTime,
	})
	if err := repo.Create(ctx, testDB, reset); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByTokenHash(ctx, testDB, "hash-of-token")
	if err != nil {
		t.Fatalf("GetByTokenHash: %v", err)
	}
	if got.UserID() != uid {
		t.Errorf("UserID = %q", got.UserID())
	}
}

func TestPasswordResetRepo_MarkUsed(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &PasswordResetsRepository{}
	userRepo := &UsersRepository{}
	uid := uuidStr()
	rid := uuidStr()

	user, _ := domain.NewUser(domain.UserParams{
		ID: uid, Name: "Test", Phone: "01012345678", PasswordHash: "hash", Now: fixedTime,
	})
	_ = userRepo.Create(ctx, testDB, user)

	reset, _ := domain.NewPasswordReset(domain.PasswordResetParams{
		ID: rid, UserID: uid, TokenHash: "hash-token-2", Now: fixedTime,
	})
	_ = repo.Create(ctx, testDB, reset)

	if err := repo.MarkUsed(ctx, testDB, rid, fixedTime.Add(time.Hour)); err != nil {
		t.Fatalf("MarkUsed: %v", err)
	}

	got, _ := repo.GetByTokenHash(ctx, testDB, "hash-token-2")
	if !got.IsUsed() {
		t.Error("reset should be marked as used")
	}
}

func TestPasswordResetRepo_MarkUsed_AlreadyUsed(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &PasswordResetsRepository{}
	userRepo := &UsersRepository{}
	uid := uuidStr()
	rid := uuidStr()

	user, _ := domain.NewUser(domain.UserParams{
		ID: uid, Name: "Test", Phone: "01012345678", PasswordHash: "hash", Now: fixedTime,
	})
	_ = userRepo.Create(ctx, testDB, user)

	reset, _ := domain.NewPasswordReset(domain.PasswordResetParams{
		ID: rid, UserID: uid, TokenHash: "hash-token-3", Now: fixedTime,
	})
	_ = repo.Create(ctx, testDB, reset)
	_ = repo.MarkUsed(ctx, testDB, rid, fixedTime.Add(time.Hour))

	err := repo.MarkUsed(ctx, testDB, rid, fixedTime.Add(2*time.Hour))
	if !errors.Is(err, domain.ErrPasswordResetAlreadyUsed) {
		t.Errorf("expected ErrPasswordResetAlreadyUsed, got %v", err)
	}
}

// ===== Merchant Repository Tests =====

func TestMerchantRepo_CreateAndGetByRestaurantID(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &MerchantsRepository{}
	id := uuidStr()

	merchant, _ := domain.NewMerchant(domain.MerchantParams{
		ID: id, RestaurantID: "rest-001", Name: "Burger Manager",
		Phone: "01212345678", PasswordHash: "hash", Now: fixedTime,
	})
	if err := repo.Create(ctx, testDB, merchant); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByRestaurantID(ctx, testDB, "rest-001")
	if err != nil {
		t.Fatalf("GetByRestaurantID: %v", err)
	}
	if got.ID() != id {
		t.Errorf("ID = %q", got.ID())
	}
}

func TestMerchantRepo_DuplicatePhone(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &MerchantsRepository{}

	m1, _ := domain.NewMerchant(domain.MerchantParams{
		ID: uuidStr(), RestaurantID: "rest-a", Name: "First",
		Phone: "01212345678", PasswordHash: "hash", Now: fixedTime,
	})
	_ = repo.Create(ctx, testDB, m1)

	m2, _ := domain.NewMerchant(domain.MerchantParams{
		ID: uuidStr(), RestaurantID: "rest-b", Name: "Second",
		Phone: "01212345678", PasswordHash: "hash", Now: fixedTime,
	})
	err := repo.Create(ctx, testDB, m2)
	if !errors.Is(err, domain.ErrMerchantAlreadyExists) {
		t.Errorf("expected ErrMerchantAlreadyExists, got %v", err)
	}
}

// ===== Internal FK Test =====

func TestInternalFK_PasswordResetToUser(t *testing.T) {
	cleanupTables(t)
	ctx := context.Background()
	repo := &PasswordResetsRepository{}

	// Create password reset for a non-existent user — should fail (FK constraint).
	reset, _ := domain.NewPasswordReset(domain.PasswordResetParams{
		ID: uuidStr(), UserID: uuidStr(), TokenHash: "hash-fk", Now: fixedTime,
	})
	err := repo.Create(ctx, testDB, reset)
	if err == nil {
		t.Error("expected FK violation error, got nil")
	}
}
