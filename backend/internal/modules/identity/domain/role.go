// Package domain contains pure domain entities for the identity module.
// This file: Role enum — represents the type of identity in the system.
// Imports stdlib only.

package domain

// Role represents a user role in the system.
// Each role corresponds to a distinct identity table (users, drivers, merchants, support_agents).
type Role string

const (
	// RoleUser is a customer who places orders.
	RoleUser Role = "user"
	// RoleDriver is a delivery driver.
	RoleDriver Role = "driver"
	// RoleMerchant is a restaurant manager.
	RoleMerchant Role = "merchant"
	// RoleAgent is a support agent.
	RoleAgent Role = "agent"
	// RoleAdmin is a super-admin (stored in users table with is_admin = true).
	RoleAdmin Role = "admin"
)

// IsValid reports whether the role is a recognized value.
func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleDriver, RoleMerchant, RoleAgent, RoleAdmin:
		return true
	}
	return false
}

// String returns the string representation of the role.
func (r Role) String() string {
	return string(r)
}

// IsDriverLike reports whether the role belongs to a driver identity
// (used for session type checks).
func (r Role) IsDriverLike() bool {
	return r == RoleDriver
}

// IsMerchantLike reports whether the role belongs to a merchant identity.
func (r Role) IsMerchantLike() bool {
	return r == RoleMerchant
}

// IsAgentLike reports whether the role belongs to a support agent identity.
func (r Role) IsAgentLike() bool {
	return r == RoleAgent
}

// IsAdmin reports whether the role is an admin.
func (r Role) IsAdmin() bool {
	return r == RoleAdmin
}
