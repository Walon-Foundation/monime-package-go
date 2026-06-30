package monime

import (
	"crypto/rand"
	"encoding/hex"
)

// generateIdempotencyKey returns a random hex string suitable for the
// Idempotency-Key header on mutating requests.
func generateIdempotencyKey() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
