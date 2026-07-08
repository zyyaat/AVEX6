// Package service mapper: converts between domain entities and port DTOs.
//
// The service layer returns DTOs (not domain entities) to prevent other
// modules from importing identity/domain. This file centralizes all
// entity → DTO conversion.
//
// PII handling:
//   - UserDTO contains the FULL phone (returned only to the user themselves
//     or to admins via authenticated endpoints).
//   - DriverProfileDTO contains a MASKED phone (returned to other modules
//     like dispatch and support that should not see the full number).
package service

import (
	"avex-backend/internal/modules/identity/domain"
	"avex-backend/internal/modules/identity/port"
)

// toUserDTO converts a domain User to a UserDTO.
// Phone is FULL (not masked) — this DTO is only returned to the user
// themselves or to admins.
func toUserDTO(u domain.User) port.UserDTO {
	return port.UserDTO{
		ID:            u.ID(),
		Name:          u.Name(),
		Phone:         u.Phone().String(),
		Email:         u.Email(),
		LoyaltyPoints: u.LoyaltyPoints(),
		IsAdmin:       u.IsAdmin(),
		Locale:        u.Locale(),
		Timezone:      u.Timezone(),
		CreatedAt:     u.CreatedAt(),
	}
}

// toDriverProfileDTO converts a domain Driver to a DriverProfileDTO.
// Phone is MASKED — this DTO is safe to return to other modules
// (dispatch, support) that should not see the full phone number.
func toDriverProfileDTO(d domain.Driver) port.DriverProfileDTO {
	return port.DriverProfileDTO{
		ID:                 d.ID(),
		Name:               d.Name(),
		PhoneMasked:        d.Phone().Masked(),
		VehicleType:        d.VehicleType().String(),
		TierID:             d.TierID(),
		Status:             d.Status().String(),
		IsOnline:           d.IsOnline(),
		IsVerified:         d.IsVerified(),
		IsActive:           d.IsActive(),
		Lat:                d.Location().Lat,
		Lng:                d.Location().Lng,
		LastSeenAt:         d.LastSeenAt(),
		MustChangePassword: d.MustChangePassword(),
	}
}

// toMerchantProfileDTO converts a domain Merchant to a MerchantProfileDTO.
func toMerchantProfileDTO(m domain.Merchant) port.MerchantProfileDTO {
	return port.MerchantProfileDTO{
		ID:           m.ID(),
		RestaurantID: m.RestaurantID(),
		Name:         m.Name(),
		IsActive:     m.IsActive(),
	}
}
