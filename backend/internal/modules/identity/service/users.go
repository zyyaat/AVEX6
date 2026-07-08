// Package service users: user read and verification use cases.
package service

import (
	"context"
	"errors"

	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
)

// GetUser retrieves a user by ID.
// Returns ErrUserNotFound if not found.
func (s *Service) GetUser(ctx context.Context, userID string) (*port.UserDTO, error) {
	user, err := s.deps.Repos.Users.GetByID(ctx, s.pool, userID)
	if err != nil {
		return nil, err
	}
	dto := toUserDTO(*user)
	return &dto, nil
}

// VerifyUserExists checks if a user with the given ID exists and is active.
// Returns true if the user exists and is active, false otherwise.
// Does NOT return an error for "not found" — only for infrastructure issues.
//
// This method is called by other modules (orders, payments) to validate
// user references without leaking entity details.
func (s *Service) VerifyUserExists(ctx context.Context, userID string) (bool, error) {
	user, err := s.deps.Repos.Users.GetByID(ctx, s.pool, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return false, nil
		}
		return false, err
	}
	return user.IsActive(), nil
}
