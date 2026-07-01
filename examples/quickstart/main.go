// Command quickstart creates a payment code and prints its USSD code.
//
// Set MONIME_SPACE_ID and MONIME_ACCESS_TOKEN in your environment, then:
//
//	go run ./examples/quickstart
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	monime "github.com/Walon-Foundation/monime-package-go"
)

func main() {
	// Credentials are read from MONIME_SPACE_ID / MONIME_ACCESS_TOKEN.
	client, err := monime.New(
		monime.WithVersion(monime.Version20250823),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	code, err := client.PaymentCode().Create(ctx, monime.CreatePaymentCodeParams{
		PaymentName: "Order #1234",
		Amount:      100, // major units; converted to minor units for the API
		Name:        "Jane Doe",
		PhoneNumber: "07600000",
	})
	if err != nil {
		var apiErr *monime.Error
		if errors.As(err, &apiErr) {
			log.Fatalf("monime error: status=%d request=%s msg=%s",
				apiErr.Status, apiErr.RequestID, apiErr.Message)
		}
		log.Fatal(err)
	}

	fmt.Printf("created payment code %s — dial %s\n", code.ID, code.USSDCode)
}
