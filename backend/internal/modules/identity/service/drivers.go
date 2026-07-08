// Package service drivers: driver management use cases.
package service

import (
	"context"
	"errors"
	"fmt"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
)

// GetDriverProfile retrieves a driver's profile by ID.
// Returns ErrDriverNotFound if not found.
func (s *Service) GetDriverProfile(ctx context.Context, driverID string) (*port.DriverProfileDTO, error) {
	driver, err := s.deps.Repos.Drivers.GetByID(ctx, s.pool, driverID)
	if err != nil {
		return nil, err
	}
	dto := toDriverProfileDTO(*driver)
	return &dto, nil
}

// UpdateDriverStatus transitions a driver's status (online/offline).
//
// Flow:
//  1. Load driver.
//  2. Apply domain state transition (validates rules: suspended can't
//     go online, unverified can't go online, etc.).
//  3. Within transaction:
//     a. Update driver (full row).
//     b. Publish DriverStatusChanged event.
//  4. Return updated profile DTO.
func (s *Service) UpdateDriverStatus(ctx context.Context, input port.UpdateDriverStatusInput) (*port.DriverProfileDTO, error) {
	// Parse target status.
	var targetStatus domain.DriverStatus
	switch input.Status {
	case "online":
		targetStatus = domain.DriverStatusOnline
	case "offline":
		targetStatus = domain.DriverStatusOffline
	default:
		return nil, fmt.Errorf("%w: %s", domain.ErrInvalidDriverStatus, input.Status)
	}

	// Load driver (non-transactional read; the entity is mutated in memory).
	driver, err := s.deps.Repos.Drivers.GetByID(ctx, s.pool, input.DriverID)
	if err != nil {
		return nil, err
	}

	now := s.deps.Clock.Now()

	// Apply domain transition (validates invariants).
	if targetStatus == domain.DriverStatusOnline {
		loc := domain.Location{Lat: input.Lat, Lng: input.Lng}
		if err := driver.GoOnline(loc, now); err != nil {
			return nil, err
		}
	} else {
		if err := driver.GoOffline(now); err != nil {
			return nil, err
		}
	}

	ec := s.eventContext(ctx, port.ActorContext{
		Type: "driver",
		ID:   driver.ID(),
	})

	// --- Transaction boundary ---
	err = s.deps.TxRunner.RunInTx(ctx, func(ctx context.Context, exec port.Executor) error {
		// Persist full driver state.
		if err := s.deps.Repos.Drivers.Update(ctx, exec, *driver); err != nil {
			return err
		}

		// Publish status changed event.
		return s.deps.EventPublisher.PublishDriverStatusChanged(ctx, exec, port.DriverStatusChangedPayload{
			DriverID: driver.ID(),
			Status:   driver.Status().String(),
			Lat:      driver.Location().Lat,
			Lng:      driver.Location().Lng,
		}, ec)
	})
	if err != nil {
		return nil, err
	}

	dto := toDriverProfileDTO(*driver)
	return &dto, nil
}

// SuspendDriver suspends a driver (admin operation).
// Records who suspended, when, and why. Revokes all active sessions.
// Idempotent — returns nil if already suspended.
func (s *Service) SuspendDriver(ctx context.Context, input port.SuspendDriverInput) error {
	driver, err := s.deps.Repos.Drivers.GetByID(ctx, s.pool, input.DriverID)
	if err != nil {
		return err
	}

	now := s.deps.Clock.Now()

	// Apply domain suspension (idempotent).
	if err := driver.Suspend(input.Reason, input.SuspendedBy, now); err != nil {
		return err
	}

	ec := s.eventContext(ctx, port.ActorContext{
		Type: "admin",
		ID:   input.SuspendedBy,
	})

	return s.deps.TxRunner.RunInTx(ctx, func(ctx context.Context, exec port.Executor) error {
		// Persist driver state.
		if err := s.deps.Repos.Drivers.Update(ctx, exec, *driver); err != nil {
			return err
		}

		// Revoke all active sessions for the driver.
		_, err := s.deps.Repos.Sessions.RevokeAllForSubject(ctx, exec, driver.ID(), domain.RoleDriver, now)
		if err != nil {
			return fmt.Errorf("revoke driver sessions: %w", err)
		}

		// Publish DriverSuspended event.
		return s.deps.EventPublisher.PublishDriverSuspended(ctx, exec, port.DriverSuspendedPayload{
			DriverID: driver.ID(),
			Reason:   input.Reason,
		}, ec)
	})
}

// VerifyDriverExists checks if a driver with the given ID exists and is active.
// Returns true if the driver exists and is active, false otherwise.
func (s *Service) VerifyDriverExists(ctx context.Context, driverID string) (bool, error) {
	driver, err := s.deps.Repos.Drivers.GetByID(ctx, s.pool, driverID)
	if err != nil {
		if errors.Is(err, domain.ErrDriverNotFound) {
			return false, nil
		}
		return false, err
	}
	return driver.IsActive(), nil
}
