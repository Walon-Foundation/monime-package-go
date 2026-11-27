// Command payment_code creates a payment code and prints its USSD code.
//
// Credentials are read from the process environment: MONIME_SPACE_ID and
// MONIME_ACCESS_TOKEN (plus optional MONIME_VERSION). This SDK reads real OS
// environment variables — it does NOT parse a .env file. If you keep secrets in
// a .env file, load it in your app first (e.g. with github.com/joho/godotenv)
// before calling monime.New.
//
//	export MONIME_SPACE_ID=spc-... MONIME_ACCESS_TOKEN=...
//	go run ./examples/payment_code
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	monime "github.com/Walon-Foundation/monime-package-go"
)

func main() {
	client, err := monime.New(
		monime.WithVersion(monime.Version20250823),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// SLE 200.00 is 20000 in minor units. Mode, enable and duration are left to
	// the API's defaults (one_time, enabled, 10 minutes).
	code, err := client.PaymentCode().Create(ctx, monime.CreatePaymentCodeParams{
		Name:                  "Walon order 1042",
		Amount:                &monime.Amount{Currency: monime.CurrencySLE.String(), Value: 20000},
		CustomerName:          "Walon",
		AuthorizedPhoneNumber: "+23276000000",
		Reference:             "order-1042",
	})
	if err != nil {
		var apiErr *monime.Error
		if errors.As(err, &apiErr) {
			log.Fatalf("monime error: status=%d request=%s msg=%s",
				apiErr.Status, apiErr.RequestID, apiErr.Message,
			)
		}
		log.Fatal(err)
	}

	fmt.Printf("payment code is: %s\n", code.USSDCode)
}
