// Package postgres mapper: converts between domain entities and DB rows.
//
// All SQL scanning and entity reconstruction lives here. The repository
// files call these helpers to keep SQL and mapping concerns separated.
package postgres

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"

	"avex-backend/internal/modules/orders/domain"
	"avex-backend/internal/modules/orders/port"
)

// ===== Order =====

// scanOrder scans a full order row from the given pgx.Row.
// Column order MUST match the SELECT statements in orders_repository.go.
func scanOrder(row pgx.Row) (domain.Order, error) {
	var (
		id               string
		orderNumber      string
		userID           string
		restaurantID     string
		driverID         *string
		customerName     string
		customerPhone    string
		deliveryLat      float64
		deliveryLng      float64
		deliveryAddress  string
		deliveryNotes    *string
		subtotalCents    int64
		deliveryFeeCents int64
		discountCents    int64
		taxCents         int64
		totalCents       int64
		currency         string
		paymentMethod    string
		status           string
		couponCode       *string
		zoneID           *string
		dispatchDistance *int
		deliveryDistance *int
		createdAt        time.Time
		updatedAt        time.Time
		confirmedAt      *time.Time
		preparingAt      *time.Time
		readyAt          *time.Time
		dispatchingAt    *time.Time
		assignedAt       *time.Time
		pickedUpAt       *time.Time
		deliveredAt      *time.Time
		cancelledAt      *time.Time
		cancelReason     *string
		cancelledBy      *string
		pickupPhotoURL   *string
		deliveryPhotoURL *string
		idempotencyKey   *string
	)

	err := row.Scan(
		&id, &orderNumber, &userID, &restaurantID, &driverID,
		&customerName, &customerPhone, &deliveryLat, &deliveryLng, &deliveryAddress, &deliveryNotes,
		&subtotalCents, &deliveryFeeCents, &discountCents, &taxCents, &totalCents, &currency, &paymentMethod,
		&status, &couponCode,
		&zoneID, &dispatchDistance, &deliveryDistance,
		&createdAt, &updatedAt, &confirmedAt, &preparingAt, &readyAt,
		&dispatchingAt, &assignedAt, &pickedUpAt, &deliveredAt,
		&cancelledAt, &cancelReason, &cancelledBy,
		&pickupPhotoURL, &deliveryPhotoURL,
		&idempotencyKey,
	)
	if err != nil {
		return domain.Order{}, err
	}

	// Build value objects.
	subtotal, _ := domain.NewMoney(subtotalCents, currency)
	deliveryFee, _ := domain.NewMoney(deliveryFeeCents, currency)
	discount, _ := domain.NewMoney(discountCents, currency)
	tax, _ := domain.NewMoney(taxCents, currency)
	total, _ := domain.NewMoney(totalCents, currency)
	pm, _ := domain.ParsePaymentMethod(paymentMethod)
	orderStatus, _ := domain.ParseOrderStatus(status)

	deliveryInfo, _ := domain.NewDeliveryInfo(deliveryLat, deliveryLng, deliveryAddress, derefStr(deliveryNotes))

	// Build dispatch info.
	dispatch := domain.DispatchInfo{}
	if zoneID != nil {
		dispatch = domain.NewDispatchInfo(*zoneID, deliveryDistance)
	}
	// For reconstruction, we need to set pointer fields directly.
	// DispatchInfo is a value object with unexported fields, so we use
	// ReconstructOrder which takes a DispatchInfo via OrderRecord.
	// We build the DispatchInfo via a helper that accepts all pointer fields.
	dispatch = domain.ReconstructDispatchInfo(domain.DispatchInfoRecord{
		DriverID:         driverID,
		ZoneID:           derefStr(zoneID),
		DispatchDistance: dispatchDistance,
		DeliveryDistance: deliveryDistance,
		PickupPhotoURL:   pickupPhotoURL,
		DeliveryPhotoURL: deliveryPhotoURL,
	})

	rec := domain.OrderRecord{
		ID:             id,
		OrderNumber:    orderNumber,
		UserID:         userID,
		RestaurantID:   restaurantID,
		CustomerName:   customerName,
		CustomerPhone:  customerPhone,
		DeliveryInfo:   deliveryInfo,
		Subtotal:       subtotal,
		DeliveryFee:    deliveryFee,
		Discount:       discount,
		Tax:            tax,
		Total:          total,
		PaymentMethod:  pm,
		Status:         orderStatus,
		CouponCode:     couponCode,
		Dispatch:       dispatch,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		ConfirmedAt:    confirmedAt,
		PreparingAt:    preparingAt,
		ReadyAt:        readyAt,
		DispatchingAt:  dispatchingAt,
		AssignedAt:     assignedAt,
		PickedUpAt:     pickedUpAt,
		DeliveredAt:    deliveredAt,
		CancelledAt:    cancelledAt,
		CancelReason:   cancelReason,
		CancelledBy:    cancelledBy,
		IdempotencyKey: derefStr(idempotencyKey),
	}
	return domain.ReconstructOrder(rec), nil
}

// orderColumns is the canonical column list for SELECT queries.
const orderColumns = `
        id, order_number, user_id, restaurant_id, driver_id,
        customer_name, customer_phone, delivery_lat, delivery_lng, delivery_address, delivery_notes,
        subtotal_cents, delivery_fee_cents, discount_cents, tax_cents, total_cents, currency, payment_method,
        status, coupon_code,
        zone_id, dispatch_distance_m, delivery_distance_m,
        created_at, updated_at, confirmed_at, preparing_at, ready_at,
        dispatching_at, assigned_at, picked_up_at, delivered_at,
        cancelled_at, cancel_reason, cancelled_by,
        pickup_photo_url, delivery_photo_url,
        idempotency_key
`

// orderInsertArgs returns args for INSERT, in column order.
func orderInsertArgs(o domain.Order) []any {
	return []any{
		o.ID(), o.OrderNumber(), o.UserID(), o.RestaurantID(), nilIfEmptyStr(o.DriverID()),
		o.CustomerName(), o.CustomerPhone(), o.DeliveryInfo().Lat(), o.DeliveryInfo().Lng(), o.DeliveryInfo().Address(), nilIfEmptyStr(o.DeliveryInfo().Notes()),
		o.Subtotal().Amount(), o.DeliveryFee().Amount(), o.Discount().Amount(), o.Tax().Amount(), o.Total().Amount(), o.Subtotal().Currency(), o.PaymentMethod().String(),
		o.Status().String(), nilIfEmptyStr(o.CouponCode()),
		nilIfEmptyStr(o.Dispatch().ZoneID()), o.Dispatch().DispatchDistancePtr(), o.Dispatch().DeliveryDistancePtr(),
		o.CreatedAt(), o.UpdatedAt(),
		o.ConfirmedAt(), o.PreparingAt(), o.ReadyAt(),
		o.DispatchingAt(), o.AssignedAt(), o.PickedUpAt(), o.DeliveredAt(),
		o.CancelledAt(), nilIfEmptyStr(o.CancelReason()), nilIfEmptyStr(o.CancelledBy()),
		o.Dispatch().PickupPhotoURLPtr(), o.Dispatch().DeliveryPhotoURLPtr(),
		nilIfEmptyStr(o.IdempotencyKey()),
	}
}

// ===== OrderItem =====

// scanOrderItem scans a single order_item row.
func scanOrderItem(row pgx.Row) (domain.OrderItem, error) {
	var (
		id         string
		orderID    string
		menuItemID string
		name       string
		nameAr     *string
		priceCents int64
		currency   string
		quantity   int
	)
	err := row.Scan(&id, &orderID, &menuItemID, &name, &nameAr, &priceCents, &currency, &quantity)
	if err != nil {
		return domain.OrderItem{}, err
	}
	price, _ := domain.NewMoney(priceCents, currency)
	item, _ := domain.NewOrderItem(menuItemID, name, derefStr(nameAr), price, quantity)
	return item, nil
}

const orderItemColumns = `id, order_id, menu_item_id, name, name_ar, price_cents, currency, quantity`

// ===== OrderAssignment =====

// scanAssignment scans a full order_assignment row.
func scanAssignment(row pgx.Row) (domain.OrderAssignment, error) {
	var (
		id             string
		orderID        string
		driverID       string
		status         string
		assignedAt     time.Time
		offerExpiresAt time.Time
		respondedAt    *time.Time
		acceptedAt     *time.Time
		rejectedAt     *time.Time
		expiredAt      *time.Time
		rejectedReason *string
		distanceM      *int
		attemptNumber  int
	)
	err := row.Scan(
		&id, &orderID, &driverID, &status,
		&assignedAt, &offerExpiresAt, &respondedAt, &acceptedAt, &rejectedAt, &expiredAt,
		&rejectedReason, &distanceM, &attemptNumber,
	)
	if err != nil {
		return domain.OrderAssignment{}, err
	}

	asStatus, _ := domain.ParseAssignmentStatus(status)
	rec := domain.AssignmentRecord{
		ID:             id,
		OrderID:        orderID,
		DriverID:       driverID,
		Status:         asStatus,
		AssignedAt:     assignedAt,
		OfferExpiresAt: offerExpiresAt,
		RespondedAt:    respondedAt,
		AcceptedAt:     acceptedAt,
		RejectedAt:     rejectedAt,
		ExpiredAt:      expiredAt,
		RejectedReason: derefStr(rejectedReason),
		DistanceM:      distanceM,
		AttemptNumber:  attemptNumber,
	}
	return domain.ReconstructAssignment(rec), nil
}

const assignmentColumns = `
        id, order_id, driver_id, assignment_status,
        assigned_at, offer_expires_at, responded_at, accepted_at, rejected_at, expired_at,
        rejected_reason, distance_m, attempt_number
`

// assignmentInsertArgs returns args for INSERT.
func assignmentInsertArgs(a domain.OrderAssignment) []any {
	return []any{
		a.ID(), a.OrderID(), a.DriverID(), a.Status().String(),
		a.AssignedAt(), a.OfferExpiresAt(), a.RespondedAt(), a.AcceptedAt(), a.RejectedAt(), a.ExpiredAt(),
		nilIfEmptyStr(a.RejectedReason()), a.DistanceMPtr(), a.AttemptNumber(),
	}
}

// ===== Outbox =====

// scanOutboxEnvelope scans an outbox row into an EventEnvelope.
func scanOutboxEnvelope(row pgx.Row) (port.EventEnvelope, error) {
	var (
		eventID       string
		eventType     string
		eventVersion  int
		schemaVersion int
		payload       []byte
		occurredAt    time.Time
		producer      string
		correlationID *string
		traceID       *string
		actorType     *string
		actorID       *string
		actorIP       *string
		actorUA       *string
	)
	err := row.Scan(
		&eventID, &eventType, &eventVersion, &schemaVersion,
		&payload, &occurredAt, &producer,
		&correlationID, &traceID,
		&actorType, &actorID, &actorIP, &actorUA,
	)
	if err != nil {
		return port.EventEnvelope{}, err
	}
	return port.EventEnvelope{
		EventID:       eventID,
		EventType:     eventType,
		EventVersion:  eventVersion,
		SchemaVersion: schemaVersion,
		Payload:       payload,
		OccurredAt:    occurredAt,
		Producer:      producer,
		CorrelationID: derefStr(correlationID),
		TraceID:       derefStr(traceID),
		ActorType:     derefStr(actorType),
		ActorID:       derefStr(actorID),
		ActorIP:       derefStr(actorIP),
		ActorUA:       derefStr(actorUA),
	}, nil
}

// ===== Helpers =====

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nilIfEmptyStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// metadataToJSON marshals a port.Metadata map to JSON bytes.
// Returns nil if the map is empty or nil.
func metadataToJSON(m port.Metadata) []byte {
	if len(m) == 0 {
		return nil
	}
	b, _ := json.Marshal(m)
	return b
}

// jsonToMetadata unmarshals JSON bytes into a port.Metadata map.
// Returns nil if the bytes are nil or empty.
func jsonToMetadata(b []byte) port.Metadata {
	if len(b) == 0 {
		return nil
	}
	var m port.Metadata
	_ = json.Unmarshal(b, &m)
	return m
}
