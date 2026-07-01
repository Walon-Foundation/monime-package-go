// Command receipts retrieves a receipt by its order number.
//
// Credentials are read from the process environment: MONIME_SPACE_ID and
// MONIME_ACCESS_TOKEN (plus optional MONIME_VERSION). This SDK reads real OS
// environment variables — it does NOT parse a .env file. If you keep secrets in
// a .env file, load it in your app first (e.g. with github.com/joho/godotenv)
// before calling monime.New.
//
//	export MONIME_SPACE_ID=spc-... MONIME_ACCESS_TOKEN=...
//	go run ./examples/receipts <order-number>
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	monime "github.com/Walon-Foundation/monime-package-go"
)

func main() {
	client, err := monime.New()
	if err != nil {
		log.Fatal(err)
	}

	orderNumber := "your-order-number"
	if len(os.Args) > 1 {
		orderNumber = os.Args[1]
	}

	ctx := context.Background()

	receipt, err := client.Receipt().Retrieve(ctx, orderNumber)
	if err != nil {
		var apiErr *monime.Error
		if errors.As(err, &apiErr) {
			log.Fatalf("monime error: status=%d request=%s msg=%s",
				apiErr.Status, apiErr.RequestID, apiErr.Message)
		}
		log.Fatal(err)
	}

	fmt.Printf("receipt %s — %s — %s %d\n",
		receipt.OrderNumber, receipt.Status,
		receipt.OrderAmount.Currency, receipt.OrderAmount.Value)
}
