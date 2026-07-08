//go:build integration

package orders_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

var e2eServer *httptest.Server

func setupE2E() {
	mux := http.NewServeMux()
	testMod.RegisterRoutes(mux)
	e2eServer = httptest.NewServer(mux)
}

func doOrderRequest(t *testing.T, method, path, idemKey string, body any) (int, map[string]any) {
	t.Helper()
	var bodyReader *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(b)
	}

	var req *http.Request
	var err error
	if bodyReader != nil {
		req, err = http.NewRequest(method, e2eServer.URL+path, bodyReader)
	} else {
		req, err = http.NewRequest(method, e2eServer.URL+path, nil)
	}
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if idemKey != "" {
		req.Header.Set("Idempotency-Key", idemKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return resp.StatusCode, result
}

func TestE2E_Orders_CreateAndGet(t *testing.T) {
	cleanupOrders(t)
	setupE2E()

	// Create
	input := validCreateInput()
	status, body := doOrderRequest(t, "POST", "/api/v1/orders", input.IdempotencyKey, map[string]any{
		"user_id": input.UserID, "restaurant_id": input.RestaurantID,
		"customer_name": input.CustomerName, "customer_phone": input.CustomerPhone,
		"delivery_info": map[string]any{"lat": input.DeliveryLat, "lng": input.DeliveryLng, "address": input.DeliveryAddress, "notes": input.DeliveryNotes},
		"items": []map[string]any{
			{"menu_item_id": "item-001", "name": "Burger", "name_ar": "برجر", "price_cents": 1299, "quantity": 2},
		},
		"subtotal_cents": 2598, "delivery_fee_cents": 399, "total_cents": 2997,
		"currency": "EGP", "payment_method": "cash", "zone_id": "zone-nasr",
	})
	if status != http.StatusCreated {
		t.Fatalf("Create: status=%d, body=%v", status, body)
	}
	data := body["data"].(map[string]any)
	orderID := data["id"].(string)

	// Get
	status, body = doOrderRequest(t, "GET", "/api/v1/orders/"+orderID, "", nil)
	if status != http.StatusOK {
		t.Fatalf("Get: status=%d", status)
	}
	data = body["data"].(map[string]any)
	if data["status"] != "pending" {
		t.Errorf("status = %v", data["status"])
	}
}

func TestE2E_Orders_TrackByNumber(t *testing.T) {
	cleanupOrders(t)
	setupE2E()
	ctx := context.Background()

	dto, _ := testSvc.CreateOrder(ctx, validCreateInput())
	status, body := doOrderRequest(t, "GET", "/api/v1/orders/track/"+dto.OrderNumber, "", nil)
	if status != http.StatusOK {
		t.Fatalf("Track: status=%d", status)
	}
	data := body["data"].(map[string]any)
	if data["id"] != dto.ID {
		t.Errorf("id = %v", data["id"])
	}
}

func TestE2E_Orders_ValidationError(t *testing.T) {
	cleanupOrders(t)
	setupE2E()

	status, body := doOrderRequest(t, "POST", "/api/v1/orders", "idem-val-001", map[string]any{
		"user_id": "", "restaurant_id": "",
	})
	if status != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", status, http.StatusBadRequest)
	}
	errBody := body["error"].(map[string]any)
	if errBody["code"] != "validation_error" {
		t.Errorf("code = %v", errBody["code"])
	}
}

func TestE2E_Orders_NotFound(t *testing.T) {
	cleanupOrders(t)
	setupE2E()

	status, body := doOrderRequest(t, "GET", "/api/v1/orders/00000000-0000-0000-0000-000000000000", "", nil)
	if status != http.StatusNotFound {
		t.Errorf("status = %d, want %d", status, http.StatusNotFound)
	}
	errBody := body["error"].(map[string]any)
	if errBody["code"] != "order_not_found" {
		t.Errorf("code = %v", errBody["code"])
	}
}

func TestE2E_Orders_IdempotencyKey(t *testing.T) {
	cleanupOrders(t)
	setupE2E()
	ctx := context.Background()

	input := validCreateInput()
	dto, _ := testSvc.CreateOrder(ctx, input)

	status, body := doOrderRequest(t, "POST", "/api/v1/orders", input.IdempotencyKey, map[string]any{
		"user_id": input.UserID, "restaurant_id": input.RestaurantID,
		"customer_name": input.CustomerName, "customer_phone": input.CustomerPhone,
		"delivery_info":  map[string]any{"lat": input.DeliveryLat, "lng": input.DeliveryLng, "address": input.DeliveryAddress},
		"items":          []map[string]any{{"menu_item_id": "item-001", "name": "Burger", "price_cents": 1299, "quantity": 2}},
		"subtotal_cents": 2598, "delivery_fee_cents": 399, "total_cents": 2997,
		"currency": "EGP", "payment_method": "cash",
	})
	if status != http.StatusCreated {
		t.Fatalf("status = %d, body = %v", status, body)
	}
	data := body["data"].(map[string]any)
	if data["id"] != dto.ID {
		t.Errorf("idempotency: API returned different order ID: %s != %s", data["id"], dto.ID)
	}
}

func TestE2E_Orders_FullLifecycleViaAPI(t *testing.T) {
	cleanupOrders(t)
	setupE2E()
	ctx := context.Background()

	dto, _ := testSvc.CreateOrder(ctx, validCreateInput())
	id := dto.ID

	// Confirm
	status, _ := doOrderRequest(t, "POST", "/api/v1/orders/"+id+"/confirm", "", nil)
	if status != http.StatusOK {
		t.Fatalf("Confirm: status=%d", status)
	}

	// Prepare
	status, _ = doOrderRequest(t, "POST", "/api/v1/orders/"+id+"/prepare", "", nil)
	if status != http.StatusOK {
		t.Fatalf("Prepare: status=%d", status)
	}

	// Ready
	status, _ = doOrderRequest(t, "POST", "/api/v1/orders/"+id+"/ready", "", nil)
	if status != http.StatusOK {
		t.Fatalf("Ready: status=%d", status)
	}

	// Dispatch
	status, _ = doOrderRequest(t, "POST", "/api/v1/orders/"+id+"/dispatch", "", nil)
	if status != http.StatusOK {
		t.Fatalf("Dispatch: status=%d", status)
	}

	// Assign
	driverID := "00000000-0000-0000-0000-000000000001"
	status, _ = doOrderRequest(t, "POST", "/api/v1/orders/"+id+"/assign", "", map[string]any{"driver_id": driverID})
	if status != http.StatusOK {
		t.Fatalf("Assign: status=%d", status)
	}

	// Pickup
	status, _ = doOrderRequest(t, "POST", "/api/v1/orders/"+id+"/pickup", "", map[string]any{"pickup_photo_url": "photo.jpg"})
	if status != http.StatusOK {
		t.Fatalf("Pickup: status=%d", status)
	}

	// Deliver
	status, _ = doOrderRequest(t, "POST", "/api/v1/orders/"+id+"/deliver", "", map[string]any{"delivery_photo_url": "delivery.jpg"})
	if status != http.StatusOK {
		t.Fatalf("Deliver: status=%d", status)
	}

	// Verify final status
	status, body := doOrderRequest(t, "GET", "/api/v1/orders/"+id, "", nil)
	if status != http.StatusOK {
		t.Fatalf("Get: status=%d", status)
	}
	data := body["data"].(map[string]any)
	if data["status"] != "delivered" {
		t.Errorf("final status = %v, want 'delivered'", data["status"])
	}
}

func TestE2E_Orders_CancelWithReason(t *testing.T) {
	cleanupOrders(t)
	setupE2E()
	ctx := context.Background()

	dto, _ := testSvc.CreateOrder(ctx, validCreateInput())

	status, body := doOrderRequest(t, "POST", "/api/v1/orders/"+dto.ID+"/cancel", "", map[string]any{"reason": "changed mind"})
	if status != http.StatusOK {
		t.Fatalf("Cancel: status=%d, body=%v", status, body)
	}
	data := body["data"].(map[string]any)
	if data["status"] != "cancelled" {
		t.Errorf("status = %v", data["status"])
	}
}

func TestE2E_Orders_CancelNoReason(t *testing.T) {
	cleanupOrders(t)
	setupE2E()
	ctx := context.Background()

	dto, _ := testSvc.CreateOrder(ctx, validCreateInput())

	status, body := doOrderRequest(t, "POST", "/api/v1/orders/"+dto.ID+"/cancel", "", map[string]any{"reason": ""})
	if status != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", status, http.StatusBadRequest)
	}
	errBody := body["error"].(map[string]any)
	if errBody["code"] != "validation_error" {
		t.Errorf("code = %v", errBody["code"])
	}
}

func TestE2E_Orders_ListMyOrders(t *testing.T) {
	cleanupOrders(t)
	setupE2E()
	ctx := context.Background()

	input := validCreateInput()
	_, _ = testSvc.CreateOrder(ctx, input)

	// With X-User-ID header
	req, _ := http.NewRequest("GET", e2eServer.URL+"/api/v1/orders/my?limit=10", nil)
	req.Header.Set("X-User-ID", input.UserID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ListMyOrders: status=%d", resp.StatusCode)
	}
}
