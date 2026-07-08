// Package service tests: integration test setup helper.
//
// setupTestService creates a fully wired Service with mock dependencies.
// Each test gets a fresh set of mocks — no state leakage between tests.
package service_test

import (
	"testing"
	"time"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
	"avex-backend/internal/modules/identity/service"
	"avex-backend/internal/modules/identity/testutil"
)

// testService bundles a Service with all its mock dependencies for
// assertion in tests.
type testService struct {
	svc          *service.Service
	userRepo     *testutil.MockUserRepository
	driverRepo   *testutil.MockDriverRepository
	merchantRepo *testutil.MockMerchantRepository
	agentRepo    *testutil.MockAgentRepository
	sessionRepo  *testutil.MockSessionRepository
	resetRepo    *testutil.MockPasswordResetRepository
	eventPub     *testutil.MockEventPublisher
	clock        *testutil.MockClock
	idGen        *testutil.MockIDGenerator
}

// setupTestService creates a fresh Service with mock dependencies.
// The clock starts at 2026-01-01 00:00:00 UTC.
func setupTestService(t *testing.T) *testService {
	t.Helper()

	clock := testutil.NewMockClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	idGen := testutil.NewMockIDGenerator()
	eventPub := testutil.NewMockEventPublisher()

	repoSet, userRepo, driverRepo, merchantRepo, agentRepo, sessionRepo, resetRepo := testutil.NewMockRepositorySet()

	deps := port.Deps{
		Clock:          clock,
		IDGenerator:    idGen,
		PasswordHasher: testutil.NewMockPasswordHasher(),
		JWTIssuer:      testutil.NewMockJWTIssuer(),
		EventPublisher: eventPub,
		Logger:         testutil.NewMockLogger(),
		TxRunner:       testutil.NewMockTxRunner(),
		Repos:          repoSet,
	}

	svc := service.New(deps, "mock-pool-exec", service.Config{
		AccessTokenTTL: 24 * time.Hour,
	})

	return &testService{
		svc:          svc,
		userRepo:     userRepo,
		driverRepo:   driverRepo,
		merchantRepo: merchantRepo,
		agentRepo:    agentRepo,
		sessionRepo:  sessionRepo,
		resetRepo:    resetRepo,
		eventPub:     eventPub,
		clock:        clock,
		idGen:        idGen,
	}
}

// seedVerifiedDriver creates a verified driver in the mock repo for tests
// that need a driver ready to go online. Returns a pointer for convenience.
func (ts *testService) seedVerifiedDriver(t *testing.T, phone string) *domain.Driver {
	t.Helper()
	driver, err := domain.NewDriver(domain.DriverParams{
		ID:            "driver-seeded",
		Name:          "Test Driver",
		Phone:         phone,
		PasswordHash:  "hash:password123",
		VehicleType:   domain.VehicleTypeMotorcycle,
		LicenseNumber: "LIC-001",
		NationalID:    "NID-001",
		Now:           ts.clock.Now(),
	})
	if err != nil {
		t.Fatalf("seed driver: %v", err)
	}
	_ = driver.Verify(ts.clock.Now())
	ts.driverRepo.SeedDriver(driver)
	return &driver
}
