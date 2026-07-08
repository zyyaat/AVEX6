package realtime

import (
	"os"
	"sync"
)

// Enabled reports whether the realtime WebSocket layer is turned on.
// Defaults to OFF so existing polling-based behavior is never affected
// unless explicitly enabled via the REALTIME_ENABLED env var.
func Enabled() bool {
	return os.Getenv("REALTIME_ENABLED") == "true"
}

// Client represents one connected WebSocket socket.
type Client struct {
	Send chan []byte
}

// Hub keeps track of connected driver sockets and the admin dispatch channel.
type Hub struct {
	mu      sync.RWMutex
	drivers map[string]map[*Client]bool // driver_id -> set of clients (allow multiple tabs/devices)
	admins  map[*Client]bool
}

var GlobalHub = &Hub{
	drivers: make(map[string]map[*Client]bool),
	admins:  make(map[*Client]bool),
}

func (h *Hub) RegisterDriver(driverID string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.drivers[driverID] == nil {
		h.drivers[driverID] = make(map[*Client]bool)
	}
	h.drivers[driverID][c] = true
}

func (h *Hub) UnregisterDriver(driverID string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.drivers[driverID]; ok {
		delete(set, c)
		if len(set) == 0 {
			delete(h.drivers, driverID)
		}
	}
	close(c.Send)
}

func (h *Hub) RegisterAdmin(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.admins[c] = true
}

func (h *Hub) UnregisterAdmin(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.admins[c] {
		delete(h.admins, c)
		close(c.Send)
	}
}

// SendToDriver pushes a JSON payload to every socket a given driver has open.
// Non-blocking: if a client's buffer is full, the message is dropped for it
// (client will still have the REST fallback for state resync).
func (h *Hub) SendToDriver(driverID string, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.drivers[driverID] {
		select {
		case c.Send <- payload:
		default:
		}
	}
}

// BroadcastAdmin pushes a JSON payload to every connected admin dispatch socket.
func (h *Hub) BroadcastAdmin(payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.admins {
		select {
		case c.Send <- payload:
		default:
		}
	}
}

func (h *Hub) IsDriverOnline(driverID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	set, ok := h.drivers[driverID]
	return ok && len(set) > 0
}
