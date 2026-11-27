// Command webhooks registers a webhook endpoint, then serves it and verifies
// every delivery before trusting it.
//
// Credentials are read from the process environment: MONIME_SPACE_ID and
// MONIME_ACCESS_TOKEN. This SDK reads real OS environment variables — it does
// NOT parse a .env file. If you keep secrets in a .env file, load it in your app
// first (e.g. with github.com/joho/godotenv) before calling monime.New.
//
// MONIME_WEBHOOK_SECRET is the HS256 secret (32 to 256 characters) the webhook
// is created with and deliveries are checked against.
//
//	MONIME_SPACE_ID=... MONIME_ACCESS_TOKEN=... MONIME_WEBHOOK_SECRET=... go run ./examples/webhooks
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	monime "github.com/Walon-Foundation/monime-package-go"
)

func main() {
	client, err := monime.New()
	if err != nil {
		log.Fatal(err)
	}
	secret := os.Getenv("MONIME_WEBHOOK_SECRET")

	ctx := context.Background()
	hook, err := client.Webhook().Create(ctx, monime.CreateWebhookParams{
		Name:   "order-events",
		URL:    "https://example.com/monime/webhook",
		Events: []string{monime.EventPaymentProcessingCompleted, monime.EventPayoutCompleted, monime.EventPayoutFailed},
		VerificationMethod: &monime.WebhookVerificationMethod{
			Type:   monime.WebhookVerificationHS256,
			Secret: secret,
		},
	})
	if err != nil {
		log.Fatalf("create webhook: %v", err)
	}
	fmt.Printf("registered webhook %s\n", hook.ID)

	http.HandleFunc("/monime/webhook", func(w http.ResponseWriter, r *http.Request) {
		// Verify the raw body: decoding and re-encoding it would change the
		// bytes and break the signature.
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, "unreadable body", http.StatusBadRequest)
			return
		}
		event, err := monime.ParseWebhookEvent(body, r.Header.Get(monime.WebhookSignatureHeader), monime.WebhookVerifyOptions{Secret: secret})
		var sigErr *monime.WebhookSignatureError
		if errors.As(err, &sigErr) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
		// Deliveries can repeat: deduplicate on event.Event.ID before acting,
		// answer 2xx quickly, and fetch the full object by event.Object.ID.
		log.Printf("event %s %s on %s %s", event.Event.ID, event.Event.Name, event.Object.Type, event.Object.ID)
		w.WriteHeader(http.StatusNoContent)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
