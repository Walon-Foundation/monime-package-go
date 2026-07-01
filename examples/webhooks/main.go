// Command webhooks registers a webhook endpoint and lists existing webhooks.
//
//	MONIME_SPACE_ID=... MONIME_ACCESS_TOKEN=... go run ./examples/webhooks
package main

import (
	"context"
	"fmt"
	"log"

	monime "github.com/Walon-Foundation/monime-package-go"
)

func main() {
	client, err := monime.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	enabled := true
	hook, err := client.Webhook().Create(ctx, monime.CreateWebhookParams{
		Name:    "order-events",
		URL:     "https://example.com/monime/webhook",
		Events:  []string{"payment.completed", "payout.completed"},
		Enabled: &enabled,
	})
	if err != nil {
		log.Fatalf("create webhook: %v", err)
	}
	fmt.Printf("registered webhook %s\n", hook.ID)

	list, err := client.Webhook().List(ctx)
	if err != nil {
		log.Fatalf("list webhooks: %v", err)
	}
	fmt.Printf("you have %d webhook(s)\n", len(list.Result))
}
