package monime

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Webhook event names for the caph release, from
// https://docs.monime.io/docs/developer-resources/guides/webhook/structure-and-events.
const (
	EventCheckoutSessionCancelled    = "checkout_session.cancelled"
	EventCheckoutSessionCompleted    = "checkout_session.completed"
	EventCheckoutSessionExpired      = "checkout_session.expired"
	EventFinancialAccountCreated     = "financial_account.created"
	EventFinancialAccountCredited    = "financial_account.credited"
	EventFinancialAccountDebited     = "financial_account.debited"
	EventFinancialTransactionCreated = "financial_transaction.created"
	EventInternalTransferCompleted   = "internal_transfer.completed"
	EventInternalTransferCreated     = "internal_transfer.created"
	EventInternalTransferFailed      = "internal_transfer.failed"
	EventPaymentCreated              = "payment.created"
	EventPaymentProcessingCompleted  = "payment.processing_completed"
	EventPaymentProcessingStarted    = "payment.processing_started"
	EventPaymentCodeCancelled        = "payment_code.cancelled"
	EventPaymentCodeCompleted        = "payment_code.completed"
	EventPaymentCodeExpired          = "payment_code.expired"
	EventPaymentCodeProcessed        = "payment_code.processed"
	EventPayoutCompleted             = "payout.completed"
	EventPayoutFailed                = "payout.failed"
	EventUssdOtpExpired              = "ussd_otp.expired"
	EventUssdOtpVerified             = "ussd_otp.verified"
)

// WebhookSignatureHeader is the header Monime signs each delivery with.
const WebhookSignatureHeader = "Monime-Signature"

// Default freshness limits from Monime's verification guide.
const (
	DefaultWebhookTolerance       = 300 * time.Second // oldest accepted delivery
	DefaultWebhookFutureTolerance = 60 * time.Second  // furthest accepted clock skew ahead
)

// WebhookVerifyOptions says how to check a delivery: Secret for an HS256
// webhook, PublicKeyPEM for an ES256 one (the PEM key shown on the webhook in the
// dashboard, or returned as VerificationMethod.PublicKey). Set both while you
// move an endpoint from one webhook to another; a delivery is then accepted if
// either matches.
type WebhookVerifyOptions struct {
	Secret       string
	PublicKeyPEM string

	// Tolerance and FutureTolerance bound the signed timestamp; zero uses the
	// defaults (300 seconds old, 60 seconds ahead).
	Tolerance       time.Duration
	FutureTolerance time.Duration

	// Now returns the current time; nil uses time.Now. For tests.
	Now func() time.Time
}

// Reasons a delivery fails verification, in WebhookSignatureError.Reason.
const (
	WebhookReasonMissingHeader     = "missing_header"
	WebhookReasonMalformedHeader   = "malformed_header"
	WebhookReasonTimestamp         = "timestamp_out_of_tolerance"
	WebhookReasonSignatureMismatch = "signature_mismatch"
	WebhookReasonInvalidPublicKey  = "invalid_public_key"
	WebhookReasonNoKey             = "no_secret_or_public_key"
	WebhookReasonMalformedPayload  = "malformed_payload"
)

// WebhookSignatureError is returned when a delivery can't be trusted. Answer
// such a request with 401 and don't process it.
type WebhookSignatureError struct {
	Reason  string
	Message string
}

func (e *WebhookSignatureError) Error() string {
	return "monime: webhook " + e.Reason + ": " + e.Message
}

func sigErr(reason, format string, args ...any) error {
	return &WebhookSignatureError{Reason: reason, Message: fmt.Sprintf(format, args...)}
}

// VerifyWebhookSignature checks that payload, the raw request body exactly as
// received, was signed by Monime. header is the value of the Monime-Signature
// header: "t=<unix seconds>,v1=<base64 signature>". The signed bytes are the
// timestamp, an underscore and the body. HS256 compares
// base64(HMAC-SHA256(secret, signed)) in constant time; ES256 verifies an
// ASN.1 DER ECDSA P-256 signature over SHA-256(signed). Several v1 values are
// accepted, and any one matching is enough, so keys can be rotated.
//
// Read the body before decoding it as JSON: re-encoding it changes the bytes
// and breaks the signature.
func VerifyWebhookSignature(payload []byte, header string, opts WebhookVerifyOptions) error {
	if opts.Secret == "" && opts.PublicKeyPEM == "" {
		return sigErr(WebhookReasonNoKey, "set Secret (HS256) or PublicKeyPEM (ES256)")
	}
	if strings.TrimSpace(header) == "" {
		return sigErr(WebhookReasonMissingHeader, "the %s header is empty", WebhookSignatureHeader)
	}

	var timestamp string
	var signatures [][]byte
	for _, part := range strings.Split(header, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			return sigErr(WebhookReasonMalformedHeader, "%q is not key=value", part)
		}
		switch key {
		case "t":
			timestamp = value
		case "v1":
			sig, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				return sigErr(WebhookReasonMalformedHeader, "v1 is not base64")
			}
			signatures = append(signatures, sig)
		}
	}
	if timestamp == "" || len(signatures) == 0 {
		return sigErr(WebhookReasonMalformedHeader, "the header needs t and v1")
	}

	seconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return sigErr(WebhookReasonMalformedHeader, "t is not a unix timestamp")
	}
	now := time.Now
	if opts.Now != nil {
		now = opts.Now
	}
	tolerance, future := opts.Tolerance, opts.FutureTolerance
	if tolerance <= 0 {
		tolerance = DefaultWebhookTolerance
	}
	if future <= 0 {
		future = DefaultWebhookFutureTolerance
	}
	age := now().Sub(time.Unix(seconds, 0))
	if age > tolerance || -age > future {
		return sigErr(WebhookReasonTimestamp, "signed %s ago, outside the accepted window", age.Round(time.Second))
	}

	signed := make([]byte, 0, len(timestamp)+1+len(payload))
	signed = append(signed, timestamp...)
	signed = append(signed, '_')
	signed = append(signed, payload...)

	if opts.Secret != "" {
		mac := hmac.New(sha256.New, []byte(opts.Secret))
		mac.Write(signed)
		expected := mac.Sum(nil)
		for _, sig := range signatures {
			if hmac.Equal(sig, expected) {
				return nil
			}
		}
	}

	if opts.PublicKeyPEM != "" {
		key, err := parseP256PublicKey(opts.PublicKeyPEM)
		if err != nil {
			return err
		}
		digest := sha256.Sum256(signed)
		for _, sig := range signatures {
			if ecdsa.VerifyASN1(key, digest[:], sig) {
				return nil
			}
		}
	}
	return sigErr(WebhookReasonSignatureMismatch, "no v1 signature matches the secret or public key")
}

func parseP256PublicKey(pemText string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemText))
	if block == nil {
		return nil, sigErr(WebhookReasonInvalidPublicKey, "the public key is not PEM")
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, sigErr(WebhookReasonInvalidPublicKey, "the public key can't be parsed: %v", err)
	}
	key, ok := parsed.(*ecdsa.PublicKey)
	if !ok || key.Curve != elliptic.P256() {
		return nil, sigErr(WebhookReasonInvalidPublicKey, "the public key is not ECDSA P-256")
	}
	return key, nil
}

// WebhookEvent is the envelope of every Monime webhook delivery. Data is a
// partial snapshot of the object: fetch the object by Object.ID when you need
// more than it carries. Deliveries can repeat and arrive out of order, so
// deduplicate on Event.ID.
type WebhookEvent struct {
	APIVersion string             `json:"apiVersion"`
	Event      WebhookEventDetail `json:"event"`
	Object     WebhookEventObject `json:"object"`
	Data       json.RawMessage    `json:"data"`
}

// WebhookEventDetail identifies the occurrence. Timestamp is Unix seconds as a
// string, as Monime sends it; Time parses it.
type WebhookEventDetail struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
}

// Time returns the event's timestamp, or the zero time if it doesn't parse.
func (d WebhookEventDetail) Time() time.Time {
	seconds, err := strconv.ParseInt(d.Timestamp, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(seconds, 0).UTC()
}

// WebhookEventObject is the resource the event is about.
type WebhookEventObject struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// ParseWebhookEvent verifies the delivery with VerifyWebhookSignature and then
// decodes its envelope. Nothing is decoded from an unverified body.
func ParseWebhookEvent(payload []byte, header string, opts WebhookVerifyOptions) (*WebhookEvent, error) {
	if err := VerifyWebhookSignature(payload, header, opts); err != nil {
		return nil, err
	}
	var event WebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, sigErr(WebhookReasonMalformedPayload, "the body is not a webhook envelope: %v", err)
	}
	return &event, nil
}
