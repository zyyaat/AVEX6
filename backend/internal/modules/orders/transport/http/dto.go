// Package http is the orders module's HTTP transport layer.
package http

// ===== Request DTOs =====

type CreateOrderRequest struct {
	UserID        string              `json:"user_id"`
	RestaurantID  string              `json:"restaurant_id"`
	CustomerName  string              `json:"customer_name"`
	CustomerPhone string              `json:"customer_phone"`
	DeliveryInfo  DeliveryInfoRequest `json:"delivery_info"`
	Items         []OrderItemRequest  `json:"items"`
	Subtotal      int64               `json:"subtotal_cents"`
	DeliveryFee   int64               `json:"delivery_fee_cents"`
	Discount      int64               `json:"discount_cents,omitempty"`
	Tax           int64               `json:"tax_cents,omitempty"`
	Total         int64               `json:"total_cents"`
	Currency      string              `json:"currency,omitempty"`
	PaymentMethod string              `json:"payment_method"`
	CouponCode    string              `json:"coupon_code,omitempty"`
	ZoneID        string              `json:"zone_id,omitempty"`
}

type DeliveryInfoRequest struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Address string  `json:"address"`
	Notes   string  `json:"notes,omitempty"`
}

type OrderItemRequest struct {
	MenuItemID string `json:"menu_item_id"`
	Name       string `json:"name"`
	NameAr     string `json:"name_ar,omitempty"`
	PriceCents int64  `json:"price_cents"`
	Quantity   int    `json:"quantity"`
}

type CancelOrderRequest struct {
	Reason string `json:"reason"`
}

type AssignDriverRequest struct {
	DriverID      string `json:"driver_id"`
	AssignmentID  string `json:"assignment_id,omitempty"`
	DispatchDistM *int   `json:"dispatch_dist_m,omitempty"`
}

type MarkPickedUpRequest struct {
	PickupPhotoURL string `json:"pickup_photo_url,omitempty"`
}

type MarkDeliveredRequest struct {
	DeliveryPhotoURL  string `json:"delivery_photo_url,omitempty"`
	DeliveryDistanceM *int   `json:"delivery_distance_m,omitempty"`
}

type ListQueryParams struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}
