// Command payouts creates a mobile-money payout and then lists payouts.
//
//	MONIME_SPACE_ID=... MONIME_ACCESS_TOKEN=... go run ./examples/payouts
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

	payout, err := client.Payout().Create(ctx, monime.CreatePayoutParams{
		Amount:        5000, // minor units
		SourceAccount: "fac-your-account-id",
		Destination: monime.PayoutDestination{
			Type:        "momo",
			ProviderID:  "m17",
			PhoneNumber: "07600000",
		},
	})
	if err != nil {
		log.Fatalf("create payout: %v", err)
	}
	fmt.Printf("created payout %s (%s)\n", payout.ID, payout.Status)

	list, err := client.Payout().List(ctx)
	if err != nil {
		log.Fatalf("list payouts: %v", err)
	}
	fmt.Printf("you have %d payout(s)\n", len(list.Result))
	for _, p := range list.Result {
		fmt.Printf("  - %s %s\n", p.ID, p.Status)
	}
}
