# monime-package-go

An unofficial Go SDK for the [Monime](https://monime.io) API. It is a
feature-for-feature port of the TypeScript SDK
[`monime-package`](https://github.com/Walon-Foundation/monime-package),
following idiomatic Go conventions.

> Status: under active development. Resources are added one at a time. This
> first release ships the core client, HTTP transport, typed errors, and
> validation. Resource APIs (payment codes, payouts, payments, …) follow.

## Install

```sh
go get github.com/Walon-Foundation/monime-package-go
```

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
		monime.WithSpaceID("your-space-id"),       // or env MONIME_SPACE_ID
		monime.WithAccessToken("your-access-token"), // or env MONIME_ACCESS_TOKEN
		monime.WithVersion(monime.Version20250823),  // optional, or env MONIME_VERSION
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = context.Background()
	_ = client
	fmt.Println("client ready")

	// Error handling pattern used by every resource method:
	//
	//   resp, err := client.PaymentCode.Create(ctx, opts)
	//   if err != nil {
	//       var apiErr *monime.Error
	//       if errors.As(err, &apiErr) {
	//           log.Printf("status=%d request=%s", apiErr.Status, apiErr.RequestID)
	//       }
	//       return err
	//   }
	_ = errors.As
}
```

## Configuration

`New` accepts functional options and falls back to environment variables for any
unset credential:

| Option | Env var | Required |
|---|---|---|
| `WithSpaceID` | `MONIME_SPACE_ID` | yes |
| `WithAccessToken` | `MONIME_ACCESS_TOKEN` | yes |
| `WithVersion` | `MONIME_VERSION` | no |
| `WithBaseURL` | — | no (mainly for tests) |
| `WithHTTPClient` | — | no (custom timeouts/transport) |

## Errors

Every method returns `(*Response, error)`. Errors are typed:

- `*monime.Error` — base API error carrying `Status`, `RequestID`, and `Details`.
- `*monime.AuthenticationError` — returned on `401`; unwraps to `*monime.Error`.
- `*monime.ValidationError` — returned when input fails local validation before
  any network call; unwraps to `*monime.Error`.

Use `errors.As` to inspect them.

## Development

```sh
go vet ./...
go test ./...
```

## License

MIT
