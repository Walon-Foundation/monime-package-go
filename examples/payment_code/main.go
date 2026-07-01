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

	code, err := client.PaymentCode().Create(ctx, monime.CreatePaymentCodeParams{
		PaymentName:        "walon",
		Amount:             200,
		Name:               "walon",
		PhoneNumber:        "070000000",
		FinancialAccountID: "hellloo",
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
