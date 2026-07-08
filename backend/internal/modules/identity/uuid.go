// Package identity uuid: minimal UUID v4 generator.
//
// This avoids importing google/uuid directly in module.go (which would
// add an import that's already available via platform/id, but keeping
// module.go dependency-free of third-party packages makes the wiring
// clearer).
//
// For production, prefer platform/id.New() which wraps google/uuid.
package identity

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// uuidV4 generates a random UUID v4 string in canonical form.
func uuidV4() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// rand.Read should never fail; if it does, panic — it's a system issue.
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	// Set version (4) and variant (10xx).
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	// Format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	h := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32])
}
