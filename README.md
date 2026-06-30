# monime-package-go

An unofficial Go SDK for the [Monime](https://monime.io) API — a feature-for-feature
port of the TypeScript SDK
[`monime-package`](https://github.com/Walon-Foundation/monime-package), written
in idiomatic Go.

- Idiomatic API: every method takes a `context.Context` and returns `(*T, error)`.
- Typed errors you can inspect with `errors.As`.
- Input validation before any network call (`go-playground/validator`).
- One small runtime dependency.

## Install

```sh
go get github.com/Walon-Foundation/monime-package-go
```

```go
import monime "github.com/Walon-Foundation/monime-package-go"
```

Requires Go 1.25+.

## Quick start

```go
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
		monime.WithSpaceID("your-space-id"),         // or env MONIME_SPACE_ID
		monime.WithAccessToken("your-access-token"), // or env MONIME_ACCESS_TOKEN
		monime.WithVersion(monime.Version20250823),  // optional, or env MONIME_VERSION
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

	fmt.Println("USSD code:", code.USSDCode)
}
```

## Configuration

`monime.New` accepts functional options and falls back to environment variables
for any unset credential:

| Option | Env var | Required |
| --- | --- | --- |
| `WithSpaceID` | `MONIME_SPACE_ID` | yes |
| `WithAccessToken` | `MONIME_ACCESS_TOKEN` | yes |
| `WithVersion` | `MONIME_VERSION` | no |
| `WithBaseURL` | — | no (mainly for tests) |
| `WithHTTPClient` | — | no (custom timeouts/transport) |

`New` returns an error if the space id or access token are missing.

## Resources

Each resource is reached through an accessor method on the client. Every method
takes a `context.Context` first.

| Accessor | Methods |
| --- | --- |
| `client.PaymentCode()` | `Create`, `Retrieve`, `List`, `Update`, `Delete` |
| `client.Payment()` | `Retrieve`, `List`, `Update` |
| `client.Payout()` | `Create`, `List`, `Retrieve`, `Update`, `Delete` |
| `client.FinancialAccount()` | `Create`, `Retrieve`, `List`, `Update` |
| `client.FinancialTransaction()` | `List`, `Retrieve` |
| `client.InternalTransfer()` | `Create`, `List`, `Retrieve`, `Update`, `Delete` |
| `client.CheckoutSession()` | `Create`, `List`, `Retrieve`, `Delete` |
| `client.Receipt()` | `Retrieve`, `Redeem` |
| `client.UssdOtp()` | `Create`, `Retrieve`, `List`, `Delete` |
| `client.ProviderKyc()` | `Retrieve` |
| `client.Webhook()` | `Create`, `Retrieve`, `List`, `Update`, `Delete` |
| `client.FinancialProvider().Bank()` | `Retrieve`, `List` |
| `client.FinancialProvider().Momo()` | `Retrieve`, `List` |

### Examples

```go
// List (returns Result + Pagination)
payouts, err := client.Payout().List(ctx)
for _, p := range payouts.Result {
	fmt.Println(p.ID, p.Status)
}

// Retrieve
account, err := client.FinancialAccount().Retrieve(ctx, "fac-123")

// Partial update (PATCH)
_, err = client.PaymentCode().Update(ctx, "pmc-123", map[string]any{
	"enable": false,
})

// Financial providers are grouped
banks, err := client.FinancialProvider().Bank().List(ctx)

// Delete returns just an error
err = client.CheckoutSession().Delete(ctx, "chs-123")
```

## Errors

Every method returns `(*Response, error)`. Errors are typed:

- `*monime.Error` — base API error carrying `Status`, `RequestID`, and `Details`.
- `*monime.AuthenticationError` — returned on `401`; unwraps to `*monime.Error`.
- `*monime.ValidationError` — returned when input fails local validation before
  any network call; unwraps to `*monime.Error`.

```go
resp, err := client.Payment().Retrieve(ctx, id)
if err != nil {
	var authErr *monime.AuthenticationError
	var apiErr *monime.Error
	switch {
	case errors.As(err, &authErr):
		// invalid/expired token
	case errors.As(err, &apiErr):
		log.Printf("status=%d request=%s", apiErr.Status, apiErr.RequestID)
	default:
		// network or decoding error
	}
}
```

## Amounts

Monetary values use the `monime.Amount` type (`Currency`, `Value`) where `Value`
is in **minor units**. Some create helpers (e.g. payment codes) take a major-unit
amount and convert it for you — see each method's documentation.

## Development

```sh
gofmt -l .
go vet ./...
go test ./...        # add -race before pushing
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the project layout and conventions,
and [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

## Disclaimer

This is an **unofficial** SDK and is not affiliated with or endorsed by Monime.
Refer to the official [Monime API documentation](https://docs.monime.io/apis)
for authoritative behavior.

## License

[MIT](LICENSE) © Walon-Foundation
