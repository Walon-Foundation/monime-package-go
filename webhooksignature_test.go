package monime

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"testing"
	"time"
)

// The same vector is pinned in the TypeScript SDK.
const (
	vectorSecret    = "whsec_test_secret_0123456789abcdef"
	vectorTimestamp = "1773386779"
	vectorBody      = `{"apiVersion":"caph.2025-08-23","event":{"id":"wke-1","name":"payout.completed","timestamp":"1773386779"},"object":{"id":"pyt-1","type":"payout"},"data":{"id":"pyt-1"}}`
	vectorSignature = "oH5YCvgxXM8T9Wqpk17s3ulNxDNMZNpFUsVvxlOkfgQ="
)

func atVector() time.Time { return time.Unix(1773386779, 0).Add(10 * time.Second) }

func wantReason(t *testing.T, err error, reason string) {
	t.Helper()
	var sErr *WebhookSignatureError
	if !errors.As(err, &sErr) || sErr.Reason != reason {
		t.Fatalf("want reason %s, got %v", reason, err)
	}
}

func TestVerifyWebhookSignature_HS256Vector(t *testing.T) {
	header := "t=" + vectorTimestamp + ",v1=" + vectorSignature
	if err := VerifyWebhookSignature([]byte(vectorBody), header, WebhookVerifyOptions{Secret: vectorSecret, Now: atVector}); err != nil {
		t.Fatalf("the documented scheme must verify: %v", err)
	}
}

func TestVerifyWebhookSignature_HS256Rejections(t *testing.T) {
	good := "t=" + vectorTimestamp + ",v1=" + vectorSignature
	opts := WebhookVerifyOptions{Secret: vectorSecret, Now: atVector}

	wantReason(t, VerifyWebhookSignature([]byte(vectorBody+" "), good, opts), WebhookReasonSignatureMismatch)
	wantReason(t, VerifyWebhookSignature([]byte(vectorBody), good, WebhookVerifyOptions{Secret: "another-secret", Now: atVector}), WebhookReasonSignatureMismatch)
	wantReason(t, VerifyWebhookSignature([]byte(vectorBody), "", opts), WebhookReasonMissingHeader)
	wantReason(t, VerifyWebhookSignature([]byte(vectorBody), "v1="+vectorSignature, opts), WebhookReasonMalformedHeader)
	wantReason(t, VerifyWebhookSignature([]byte(vectorBody), "t="+vectorTimestamp+",v1=***", opts), WebhookReasonMalformedHeader)
	wantReason(t, VerifyWebhookSignature([]byte(vectorBody), good, WebhookVerifyOptions{Now: atVector}), WebhookReasonNoKey)

	stale := func() time.Time { return time.Unix(1773386779, 0).Add(301 * time.Second) }
	wantReason(t, VerifyWebhookSignature([]byte(vectorBody), good, WebhookVerifyOptions{Secret: vectorSecret, Now: stale}), WebhookReasonTimestamp)
	early := func() time.Time { return time.Unix(1773386779, 0).Add(-61 * time.Second) }
	wantReason(t, VerifyWebhookSignature([]byte(vectorBody), good, WebhookVerifyOptions{Secret: vectorSecret, Now: early}), WebhookReasonTimestamp)
}

func TestVerifyWebhookSignature_UsesTheUnderscoreSeparator(t *testing.T) {
	// A signature over "t.body" (the separator other providers use) must not
	// verify: Monime signs "t_body".
	mac := hmac.New(sha256.New, []byte(vectorSecret))
	mac.Write([]byte(vectorTimestamp + "." + vectorBody))
	header := "t=" + vectorTimestamp + ",v1=" + base64.StdEncoding.EncodeToString(mac.Sum(nil))
	wantReason(t, VerifyWebhookSignature([]byte(vectorBody), header, WebhookVerifyOptions{Secret: vectorSecret, Now: atVector}), WebhookReasonSignatureMismatch)
}

func TestVerifyWebhookSignature_AcceptsAnyOfSeveralSignatures(t *testing.T) {
	header := "t=" + vectorTimestamp + ",v1=" + base64.StdEncoding.EncodeToString([]byte("old key")) + ",v1=" + vectorSignature
	if err := VerifyWebhookSignature([]byte(vectorBody), header, WebhookVerifyOptions{Secret: vectorSecret, Now: atVector}); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyWebhookSignature_ES256(t *testing.T) {
	private, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, _ := x509.MarshalPKIXPublicKey(&private.PublicKey)
	publicPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))

	digest := sha256.Sum256([]byte(vectorTimestamp + "_" + vectorBody))
	sig, err := ecdsa.SignASN1(rand.Reader, private, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	header := "t=" + vectorTimestamp + ",v1=" + base64.StdEncoding.EncodeToString(sig)
	opts := WebhookVerifyOptions{PublicKeyPEM: publicPEM, Now: atVector}

	event, err := ParseWebhookEvent([]byte(vectorBody), header, opts)
	if err != nil {
		t.Fatal(err)
	}
	if event.Event.ID != "wke-1" || event.Event.Name != EventPayoutCompleted || event.Object.Type != "payout" || event.Event.Time().Unix() != 1773386779 {
		t.Fatalf("event = %+v", event)
	}

	wantReason(t, VerifyWebhookSignature([]byte(vectorBody+"x"), header, opts), WebhookReasonSignatureMismatch)
	wantReason(t, VerifyWebhookSignature([]byte(vectorBody), header, WebhookVerifyOptions{PublicKeyPEM: "not a key", Now: atVector}), WebhookReasonInvalidPublicKey)
}

func TestParseWebhookEvent_DecodesNothingUnverified(t *testing.T) {
	if _, err := ParseWebhookEvent([]byte(vectorBody), "t="+vectorTimestamp+",v1=AAAA", WebhookVerifyOptions{Secret: vectorSecret, Now: atVector}); err == nil {
		t.Fatal("an unverified body must not be decoded")
	}
}
