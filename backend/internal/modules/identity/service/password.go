// Package service password: password change and reset use cases.
package service

import (
	"context"
	"errors"
	"fmt"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
)

// ChangePassword changes a user's password.
// Returns ErrPasswordMismatch if oldPassword is wrong.
// Revokes all other sessions for the user after a successful change.
func (s *Service) ChangePassword(ctx context.Context, input port.ChangePasswordInput) error {
	if len(input.NewPassword) < minPasswordLength {
		return domain.ErrPasswordTooShort
	}

	user, err := s.deps.Repos.Users.GetByID(ctx, s.pool, input.SubjectID)
	if err != nil {
		return err
	}

	// Verify old password.
	if err := s.deps.PasswordHasher.Compare(user.PasswordHash(), input.OldPassword); err != nil {
		return domain.ErrPasswordMismatch
	}

	// Hash new password (outside transaction).
	newHash, err := s.deps.PasswordHasher.Hash(input.NewPassword)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	now := s.deps.Clock.Now()
	if err := user.ChangePassword(newHash, now); err != nil {
		return err
	}

	ec := s.eventContext(ctx, port.ActorContext{
		Type: "user",
		ID:   user.ID(),
	})

	return s.deps.TxRunner.RunInTx(ctx, func(ctx context.Context, exec port.Executor) error {
		// Persist updated user.
		if err := s.deps.Repos.Users.Update(ctx, exec, *user); err != nil {
			return err
		}

		// Revoke all sessions (forces re-login on all devices).
		_, err := s.deps.Repos.Sessions.RevokeAllForSubject(ctx, exec, user.ID(), domain.RoleUser, now)
		if err != nil {
			return fmt.Errorf("revoke user sessions: %w", err)
		}

		// Publish password changed event.
		return s.deps.EventPublisher.PublishUserPasswordChanged(ctx, exec, port.UserPasswordChangedPayload{
			UserID: user.ID(),
		}, ec)
	})
}

// ChangeDriverPassword changes a driver's password.
// Same flow as ChangePassword but for drivers.
func (s *Service) ChangeDriverPassword(ctx context.Context, input port.ChangePasswordInput) error {
	if len(input.NewPassword) < minPasswordLength {
		return domain.ErrPasswordTooShort
	}

	driver, err := s.deps.Repos.Drivers.GetByID(ctx, s.pool, input.SubjectID)
	if err != nil {
		return err
	}

	if err := s.deps.PasswordHasher.Compare(driver.PasswordHash(), input.OldPassword); err != nil {
		return domain.ErrPasswordMismatch
	}

	newHash, err := s.deps.PasswordHasher.Hash(input.NewPassword)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}

	now := s.deps.Clock.Now()
	if err := driver.ChangePassword(newHash, now); err != nil {
		return err
	}

	ec := s.eventContext(ctx, port.ActorContext{
		Type: "driver",
		ID:   driver.ID(),
	})

	return s.deps.TxRunner.RunInTx(ctx, func(ctx context.Context, exec port.Executor) error {
		if err := s.deps.Repos.Drivers.Update(ctx, exec, *driver); err != nil {
			return err
		}

		_, err := s.deps.Repos.Sessions.RevokeAllForSubject(ctx, exec, driver.ID(), domain.RoleDriver, now)
		if err != nil {
			return fmt.Errorf("revoke driver sessions: %w", err)
		}

		// Reuse the user password changed event for drivers (the event
		// type is generic enough — "user" here means "identity subject").
		return s.deps.EventPublisher.PublishUserPasswordChanged(ctx, exec, port.UserPasswordChangedPayload{
			UserID: driver.ID(),
		}, ec)
	})
}

// HashPassword hashes a plaintext password.
// Exposed for use cases where a hash is needed without going through
// the full register flow (e.g. seeding, admin-created accounts).
func (s *Service) HashPassword(ctx context.Context, password string) (string, error) {
	if len(password) < minPasswordLength {
		return "", domain.ErrPasswordTooShort
	}
	hash, err := s.deps.PasswordHasher.Hash(password)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return hash, nil
}

// Note: Password reset flow (RequestPasswordReset + ResetPassword) is
// defined here but NOT exposed via ServicePort in Phase 1. It will be
// added when the notifications module is ready to send reset emails.
// The implementation is provided for completeness.

// requestPasswordReset generates a reset token and persists its hash.
// The plain token is returned to the caller (service layer) which is
// responsible for sending it to the user via the notifications module.
//
// This is an internal helper — not part of ServicePort yet.
func (s *Service) requestPasswordReset(ctx context.Context, phone string) (plainToken string, err error) {
	p, err := domain.NewPhone(phone)
	if err != nil {
		return "", err
	}

	user, err := s.deps.Repos.Users.GetByPhone(ctx, s.pool, p)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// Silently succeed to prevent user enumeration.
			// Return a dummy token that won't match anything.
			return "no-such-user", nil
		}
		return "", err
	}

	plainToken = s.deps.IDGenerator.New()
	// Hash the token before storing (same algorithm as password hashing).
	tokenHash, err := s.deps.PasswordHasher.Hash(plainToken)
	if err != nil {
		return "", fmt.Errorf("hash reset token: %w", err)
	}

	resetID := s.deps.IDGenerator.New()
	reset, err := domain.NewPasswordReset(domain.PasswordResetParams{
		ID:        resetID,
		UserID:    user.ID(),
		TokenHash: tokenHash,
		Now:       s.deps.Clock.Now(),
	})
	if err != nil {
		return "", err
	}

	return plainToken, s.deps.TxRunner.RunInTx(ctx, func(ctx context.Context, exec port.Executor) error {
		return s.deps.Repos.PasswordResets.Create(ctx, exec, reset)
	})
}

// resetPassword validates a reset token and changes the user's password.
//
// This is an internal helper — not part of ServicePort yet.
func (s *Service) resetPassword(ctx context.Context, plainToken string, newPassword string) error {
	if len(newPassword) < minPasswordLength {
		return domain.ErrPasswordTooShort
	}

	// We can't query by plain token (only hash is stored). The caller
	// must provide the token hash. This method is a placeholder that
	// will be wired when the HTTP endpoint is added.
	// For now, this is intentionally unimplemented.
	return errors.New("resetPassword: not yet wired — requires token hash lookup")
}
