// Package domain tests: Role enum.
package domain

import "testing"

func TestRole_IsValid(t *testing.T) {
	valid := []Role{RoleUser, RoleDriver, RoleMerchant, RoleAgent, RoleAdmin}
	for _, r := range valid {
		if !r.IsValid() {
			t.Errorf("%q should be valid", r)
		}
	}
	invalid := []Role{"", "unknown", "customer", "USER"}
	for _, r := range invalid {
		if r.IsValid() {
			t.Errorf("%q should be invalid", r)
		}
	}
}

func TestRole_IsDriverLike(t *testing.T) {
	if !RoleDriver.IsDriverLike() {
		t.Error("Driver should be driver-like")
	}
	for _, r := range []Role{RoleUser, RoleMerchant, RoleAgent, RoleAdmin} {
		if r.IsDriverLike() {
			t.Errorf("%q should not be driver-like", r)
		}
	}
}

func TestRole_IsMerchantLike(t *testing.T) {
	if !RoleMerchant.IsMerchantLike() {
		t.Error("Merchant should be merchant-like")
	}
	for _, r := range []Role{RoleUser, RoleDriver, RoleAgent, RoleAdmin} {
		if r.IsMerchantLike() {
			t.Errorf("%q should not be merchant-like", r)
		}
	}
}

func TestRole_IsAgentLike(t *testing.T) {
	if !RoleAgent.IsAgentLike() {
		t.Error("Agent should be agent-like")
	}
	for _, r := range []Role{RoleUser, RoleDriver, RoleMerchant, RoleAdmin} {
		if r.IsAgentLike() {
			t.Errorf("%q should not be agent-like", r)
		}
	}
}

func TestRole_IsAdmin(t *testing.T) {
	if !RoleAdmin.IsAdmin() {
		t.Error("Admin should be admin")
	}
	for _, r := range []Role{RoleUser, RoleDriver, RoleMerchant, RoleAgent} {
		if r.IsAdmin() {
			t.Errorf("%q should not be admin", r)
		}
	}
}

func TestRole_String(t *testing.T) {
	tests := []struct {
		role Role
		want string
	}{
		{RoleUser, "user"},
		{RoleDriver, "driver"},
		{RoleMerchant, "merchant"},
		{RoleAgent, "agent"},
		{RoleAdmin, "admin"},
	}
	for _, tt := range tests {
		if tt.role.String() != tt.want {
			t.Errorf("%q.String() = %q, want %q", tt.role, tt.role.String(), tt.want)
		}
	}
}
