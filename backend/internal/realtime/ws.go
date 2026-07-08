package realtime

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"avex-backend/internal/shared"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// InboundMessage is what the driver/admin app sends over the socket.
type InboundMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// OutboundMessage is what the server pushes over the socket.
type OutboundMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// HeartbeatPayload is sent by drivers every few seconds.
type HeartbeatPayload struct {
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	Status string  `json:"status,omitempty"`
}

func send(c *Client, msgType string, data interface{}) {
	b, err := json.Marshal(OutboundMessage{Type: msgType, Data: data})
	if err != nil {
		return
	}
	select {
	case c.Send <- b:
	default:
	}
}

// RegisterRoutes wires the WebSocket endpoints. If REALTIME_ENABLED is not
// "true", both endpoints respond 404 and have zero effect on existing
// REST/polling behavior.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /ws/driver", handleDriverWS)
	mux.HandleFunc("GET /ws/admin", handleAdminWS)
}

func authFromQuery(r *http.Request) (*shared.Claims, error) {
	token := r.URL.Query().Get("token")
	return shared.VerifyJWT(token)
}

func handleDriverWS(w http.ResponseWriter, r *http.Request) {
	if !Enabled() {
		http.NotFound(w, r)
		return
	}
	claims, err := authFromQuery(r)
	if err != nil || !claims.IsDriver {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &Client{Send: make(chan []byte, 32)}
	driverID := claims.DriverID
	GlobalHub.RegisterDriver(driverID, client)
	shared.DB.Exec("UPDATE drivers SET last_seen_at = NOW() WHERE id = $1", driverID)

	done := make(chan struct{})
	go writePump(conn, client, done)
	readPumpDriver(conn, driverID, client, done)
}

func handleAdminWS(w http.ResponseWriter, r *http.Request) {
	if !Enabled() {
		http.NotFound(w, r)
		return
	}
	claims, err := authFromQuery(r)
	if err != nil || !claims.Admin {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &Client{Send: make(chan []byte, 32)}
	GlobalHub.RegisterAdmin(client)

	done := make(chan struct{})
	go writePump(conn, client, done)
	readPumpAdmin(conn, client, done)
}

func writePump(conn *websocket.Conn, c *Client, done chan struct{}) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				conn.Close()
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}

func readPumpDriver(conn *websocket.Conn, driverID string, client *Client, done chan struct{}) {
	defer func() {
		close(done)
		GlobalHub.UnregisterDriver(driverID, client)
		conn.Close()
	}()
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var msg InboundMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		switch msg.Type {
		case "heartbeat":
			var hb HeartbeatPayload
			if err := json.Unmarshal(msg.Data, &hb); err != nil {
				continue
			}
			_, err := shared.DB.Exec(
				`UPDATE drivers SET lat = $1, lng = $2, location_updated_at = NOW(), last_seen_at = NOW() WHERE id = $3`,
				hb.Lat, hb.Lng, driverID,
			)
			if err != nil {
				log.Printf("ws heartbeat update failed for driver %s: %v", driverID, err)
				continue
			}
			GlobalHub.BroadcastAdmin(mustJSON("driver_status_changed", map[string]interface{}{
				"driver_id": driverID, "lat": hb.Lat, "lng": hb.Lng,
			}))
		}
	}
}

func readPumpAdmin(conn *websocket.Conn, client *Client, done chan struct{}) {
	defer func() {
		close(done)
		GlobalHub.UnregisterAdmin(client)
		conn.Close()
	}()
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
		// Admin dispatch channel is currently receive-only (server -> admin).
	}
}

func mustJSON(msgType string, data interface{}) []byte {
	b, _ := json.Marshal(OutboundMessage{Type: msgType, Data: data})
	return b
}

// NotifyDriverOrderOffer pushes a new order offer to a driver's socket(s).
// Safe to call even if the driver has no open socket (falls back to polling).
func NotifyDriverOrderOffer(driverID string, offer interface{}) {
	if !Enabled() {
		return
	}
	GlobalHub.SendToDriver(driverID, mustJSON("order_offer", offer))
}

// NotifyDriverOrderUpdate pushes an order status change to a driver's socket(s).
func NotifyDriverOrderUpdate(driverID string, order interface{}) {
	if !Enabled() {
		return
	}
	GlobalHub.SendToDriver(driverID, mustJSON("order_update", order))
}

// NotifyAdminZoneUpdate broadcasts a zone change to all connected admin sockets.
func NotifyAdminZoneUpdate(zone interface{}) {
	if !Enabled() {
		return
	}
	GlobalHub.BroadcastAdmin(mustJSON("zone_update", zone))
}
