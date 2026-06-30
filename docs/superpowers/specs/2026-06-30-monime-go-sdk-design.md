# Monime Go SDK — Design

Date: 2026-06-30
Status: Approved

## Goal

Build an unofficial Go SDK for the Monime API, porting the existing TypeScript
SDK at `/home/walon/Code/Walon-Foundation/monime-package` feature-for-feature
into idiomatic Go. Verify each resource against the official docs at
<https://docs.monime.io/apis> and mirror any discovered fixes back into the TS
package.

## Key decisions

- **API style:** Idiomatic Go. Every method returns `(*Response, error)`.
  Errors are typed and inspected with `errors.As`. No `Result{success,...}`
  wrapper.
- **Validation:** `github.com/go-playground/validator/v10` driven by struct
  tags on request types.
- **E2E tests:** `httptest.Server` — no real network, no credentials, runs in
  CI. Mirrors the TS approach of mocking the transport.
- **PR granularity:** One foundation PR (core infra), then one branch + PR per
  resource (~14 PRs total).
- **Merge ownership:** The author opens each PR and stops. The **user** reviews
  and merges. After merge, the author runs `git checkout main && git pull` and
  starts the next feature on top of the merged result.

## Package layout

Single flat Go package `monime` (module
`github.com/Walon-Foundation/monime-package-go`), one file per resource.

```
client.go            Client, functional options, env fallback, baseURL override
transport.go         internal do(ctx, method, path, body, idemKey, out)
errors.go            Error, AuthenticationError, ValidationError
types.go             core: Amount, Pagination, MonimeVersion consts
validate.go          shared validator instance + helper
idempotency.go       crypto/rand key generator (port of utils.go)
paymentcode.go       PaymentCodeService + request/response types + validate tags
payout.go payment.go financialaccount.go ... (one per resource)
financialprovider.go FinancialProvider{ Bank, Momo }
<resource>_test.go      unit tests (validation, body construction)
<resource>_e2e_test.go  httptest server tests
README.md
.github/workflows/go.yml  go vet + go test on push/PR
```

Rationale for a flat package: the SDK is small-to-mid sized; a flat package
preserves `client.Resource.Method` ergonomics without per-resource import churn.

## Client & configuration

```go
client, err := monime.New(
    monime.WithSpaceID("..."),       // or env MONIME_SPACE_ID
    monime.WithAccessToken("..."),   // or env MONIME_ACCESS_TOKEN
    monime.WithVersion(monime.Version20250823),
    // monime.WithHTTPClient(...), monime.WithBaseURL(...)  // tests / advanced
)
```

- `New` resolves options, falls back to env vars, and validates that
  `spaceID` and `accessToken` are present. Returns `(*Client, error)` with the
  same friendly guidance messages as the TS SDK.
- All resource methods take `context.Context` as the first argument — the one
  idiomatic addition over the TS SDK.
- `WithBaseURL` / `WithHTTPClient` exist mainly so e2e tests can point the
  client at an `httptest.Server`. Default base URL: `https://api.monime.io/v1`.

## Transport & errors

`transport.go` centralizes a single `do()`:

- Builds the request and sets headers: `Content-Type: application/json`,
  `Authorization: Bearer <token>`, `Monime-Space-Id`, optional `Monime-Version`,
  optional `Idempotency-Key`.
- Unwraps the `{ "result": ... }` envelope into the caller's `out` value.
- Maps non-2xx responses to typed errors: `401 → *AuthenticationError`,
  otherwise `*Error` carrying `Status`, `RequestID` (from `x-request-id`), and
  `Details`. `204 No Content` returns no body.

```go
type Error struct {
    Message   string
    Status    int
    RequestID string
    Details   any
}
func (e *Error) Error() string { return e.Message }

type AuthenticationError struct{ *Error }
type ValidationError      struct{ *Error }
```

Embedding `*Error` keeps `errors.As(err, &apiErr)` working for both the base and
specialized types.

## Resources

Faithful port of every method, request body shape, and response type from the TS
source. Notable body logic preserved: `amount * 100` minor units, payment-code
`mode`/`duration`/`recurrentPaymentTarget`, etc.

| Resource | Methods | Accessor |
|---|---|---|
| paymentCode | create, retrieve, list, delete | `client.PaymentCode` |
| payout | create, list, retrieve, delete | `client.Payout` |
| payment | retrieve, list, update | `client.Payment` |
| financialAccount | create, retrieve, list | `client.FinancialAccount` |
| financialTransaction | list, retrieve | `client.FinancialTransaction` |
| internalTransfer | create, list, retrieve, delete | `client.InternalTransfer` |
| checkoutSession | create, list, retrieve, delete | `client.CheckoutSession` |
| receipt | retrieve, redeem | `client.Receipt` |
| ussdOtp | create | `client.UssdOtp` |
| providerKyc | retrieve | `client.ProviderKyc` |
| webhook | create, retrieve, list, update, delete | `client.Webhook` |
| bank | retrieve, list | `client.FinancialProvider.Bank` |
| momo | retrieve, list | `client.FinancialProvider.Momo` |

## Docs sync & TS mirroring

For each resource, before/while implementing:

1. `WebFetch` the matching page under <https://docs.monime.io/apis>.
2. Confirm path, HTTP method, and field names against the TS source.
3. If the live docs differ, the **Go code follows the docs**.
4. If that same discrepancy is an outdated/buggy field in the TS package, fix it
   there too in a clearly-labelled commit. If there is no discrepancy, the TS
   package stays untouched.

## Testing

- **Unit tests** (`_test.go`, package `monime`): validation rules and request
  body construction, table-driven.
- **E2E tests** (`_e2e_test.go`, black-box where practical): `httptest.Server`
  asserting method, path, headers, request body, and response/error decoding —
  including the auth-error and network-failure paths the TS e2e tests cover.
- Gate for every PR: `go vet ./...` and `go test ./...` must pass.

## Per-feature workflow

1. `git checkout main && git pull`
2. `git checkout -b feat/<resource>`
3. Implement resource + unit + e2e tests; verify against docs.
4. `go vet ./... && go test ./...` (must pass).
5. Commit, `git push -u origin feat/<resource>`.
6. `gh pr create` with a summary of the resource and its methods.
7. **Stop. User reviews and merges the PR.**
8. After merge, return to step 1 for the next resource.

The **core-infrastructure PR** (client, transport, errors, core types, validate
helper, idempotency, go.mod deps, README, CI workflow) lands first; it replaces
the existing sketch files `monime.go`, `paymentCode.go`, and `utils.go`.

## Out of scope

- Real-network integration tests against the live Monime API.
- Retries/backoff, rate-limit handling, pagination auto-iteration (can be added
  later; not present in the TS source).
