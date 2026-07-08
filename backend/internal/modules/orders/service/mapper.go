// Package service mapper: converts between domain entities and port DTOs.
package service

import (
	"avex-backend/internal/modules/orders/domain"
	"avex-backend/internal/modules/orders/port"
)

func toMoneyDTO(m domain.Money) port.MoneyDTO {
	return port.MoneyDTO{AmountCents: m.Amount(), Currency: m.Currency()}
}

func toOrderItemDTO(item domain.OrderItem) port.OrderItemDTO {
	lineTotal, _ := item.LineTotal()
	return port.OrderItemDTO{
		MenuItemID: item.MenuItemID(),
		Name:       item.Name(),
		NameAr:     item.NameAr(),
		Price:      toMoneyDTO(item.Price()),
		Quantity:   item.Quantity(),
		LineTotal:  toMoneyDTO(lineTotal),
	}
}

func toDeliveryInfoDTO(d domain.DeliveryInfo) port.DeliveryInfoDTO {
	return port.DeliveryInfoDTO{
		Lat:     d.Lat(),
		Lng:     d.Lng(),
		Address: d.Address(),
		Notes:   d.Notes(),
	}
}

func toDispatchInfoDTO(d domain.DispatchInfo) port.DispatchInfoDTO {
	return port.DispatchInfoDTO{
		DriverID:          d.DriverID(),
		ZoneID:            d.ZoneID(),
		DispatchDistanceM: d.DispatchDistance(),
		DeliveryDistanceM: d.DeliveryDistance(),
		PickupPhotoURL:    d.PickupPhotoURL(),
		DeliveryPhotoURL:  d.DeliveryPhotoURL(),
	}
}

func toOrderDTO(o domain.Order, items []domain.OrderItem) port.OrderDTO {
	itemDTOs := make([]port.OrderItemDTO, 0, len(items))
	for _, item := range items {
		itemDTOs = append(itemDTOs, toOrderItemDTO(item))
	}

	return port.OrderDTO{
		ID:            o.ID(),
		OrderNumber:   o.OrderNumber(),
		UserID:        o.UserID(),
		RestaurantID:  o.RestaurantID(),
		DriverID:      o.DriverID(),
		CustomerName:  o.CustomerName(),
		CustomerPhone: o.CustomerPhone(),
		DeliveryInfo:  toDeliveryInfoDTO(o.DeliveryInfo()),
		Items:         itemDTOs,
		Subtotal:      toMoneyDTO(o.Subtotal()),
		DeliveryFee:   toMoneyDTO(o.DeliveryFee()),
		Discount:      toMoneyDTO(o.Discount()),
		Tax:           toMoneyDTO(o.Tax()),
		Total:         toMoneyDTO(o.Total()),
		PaymentMethod: o.PaymentMethod().String(),
		Status:        o.Status().String(),
		CouponCode:    o.CouponCode(),
		Dispatch:      toDispatchInfoDTO(o.Dispatch()),
		CreatedAt:     o.CreatedAt(),
		UpdatedAt:     o.UpdatedAt(),
		ConfirmedAt:   o.ConfirmedAt(),
		PreparingAt:   o.PreparingAt(),
		ReadyAt:       o.ReadyAt(),
		DispatchingAt: o.DispatchingAt(),
		AssignedAt:    o.AssignedAt(),
		PickedUpAt:    o.PickedUpAt(),
		DeliveredAt:   o.DeliveredAt(),
		CancelledAt:   o.CancelledAt(),
		CancelReason:  o.CancelReason(),
		CancelledBy:   o.CancelledBy(),
	}
}
