package monime

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

// maxIdempotencyKeyLength is the limit the API puts on the Idempotency-Key
// header. The docs recommend keys longer than 25 characters, which is advice,
// not a rule, so shorter keys are passed through.
const maxIdempotencyKeyLength = 64

type idempotencyKeyCtx struct{}

// WithIdempotencyKey returns a context that makes the next create call use key
// as its Idempotency-Key instead of a random one. Reuse the same key when you
// retry a request whose outcome you don't know (a timeout, a dropped
// connection): Monime then returns the first result instead of acting twice.
//
//	ctx := monime.WithIdempotencyKey(ctx, monime.IdempotencyKeyFor("payout:"+lineID))
//	payout, err := client.Payout().Create(ctx, params)
//
// The key must be 1 to 64 characters; anything else fails validation before
// the request is sent.
func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idempotencyKeyCtx{}, key)
}

// idempotencyKeyNamespace is the UUIDv5 namespace for IdempotencyKeyFor:
// uuid5(NAMESPACE_URL, "https://github.com/Walon-Foundation/monime-package").
// The TypeScript SDK uses the same one, so both produce the same key for the
// same reference.
var idempotencyKeyNamespace = [16]byte{
	0x32, 0x4f, 0x9f, 0x23, 0x0e, 0x7a, 0x54, 0xba,
	0x91, 0x49, 0x82, 0x40, 0xb1, 0x83, 0x3d, 0xc1,
}

// IdempotencyKeyFor derives a deterministic Idempotency-Key (a UUIDv5) from a
// stable business reference such as "payout:<line id>:<attempt>". The Monime
// docs recommend this so a retry from any process, or after a restart, sends
// the same key.
func IdempotencyKeyFor(reference string) string {
	h := sha1.New()
	h.Write(idempotencyKeyNamespace[:])
	h.Write([]byte(reference))
	sum := h.Sum(nil)

	var u [16]byte
	copy(u[:], sum[:16])
	u[6] = (u[6] & 0x0f) | 0x50 // version 5
	u[8] = (u[8] & 0x3f) | 0x80 // RFC 4122 variant

	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}

// idempotencyKey returns the caller's key from ctx when one was set with
// WithIdempotencyKey, validated, and a fresh random key otherwise.
func idempotencyKey(ctx context.Context) (string, error) {
	if key, ok := ctx.Value(idempotencyKeyCtx{}).(string); ok {
		if key == "" || len(key) > maxIdempotencyKeyLength {
			return "", newValidationError(fmt.Sprintf("idempotency key must be 1 to %d characters", maxIdempotencyKeyLength))
		}
		return key, nil
	}
	key, err := generateIdempotencyKey()
	if err != nil {
		return "", &Error{Message: err.Error()}
	}
	return key, nil
}

// generateIdempotencyKey returns a random hex string suitable for the
// Idempotency-Key header on mutating requests.
func generateIdempotencyKey() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
