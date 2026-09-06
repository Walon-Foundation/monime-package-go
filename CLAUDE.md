# CLAUDE.md

Guidance for Claude Code when working in this repository.

## What this is

An **unofficial** Go SDK for the [Monime](https://monime.io) API, published as
`github.com/Walon-Foundation/monime-package-go`. It is a deliberate
feature-for-feature port of the TypeScript SDK
[`monime-package`](https://github.com/Walon-Foundation/monime-package): when Go
behavior and TS behavior disagree, that is a bug unless the API docs at
<https://docs.monime.io/apis> say otherwise. Only runtime dependency is
`go-playground/validator/v10`; keep it that way.

## Commands

```sh
gofmt -l .        # must print nothing
go vet ./...
go test ./...     # use -race before pushing; CI runs go test -race ./...
go test -run TestPayout ./...
```

CI (`.github/workflows/go.yml`) runs `go mod verify`, `go vet`, `go test -race`.
There is no linter beyond `gofmt`/`go vet`.

## Layout

One flat package (`package monime`) at the repo root — no subpackages except
`examples/`. Infrastructure files: `client.go` (options + env fallback),
`transport.go` (the single `do()`), `errors.go`, `validate.go`,
`idempotency.go`, `listoptions.go` (`ListOption` + every query-param option),
`types.go` (`Amount`, `Pagination`, `MonimeVersion`).
Each resource is a trio: `<resource>.go`, `<resource>_test.go` (validation /
body shape, no network), `<resource>_e2e_test.go` (`httptest` server asserting
method, path, headers, body, decoding, and the validation short-circuit).

## Conventions that matter

- **Everything routes through `Client.do`** in `transport.go`. Never build an
  `http.Request` inside a resource file.
- **`rawBody`**: Monime wraps responses as
  `{"success","messages","result","pagination"}`. `do()` unwraps `result` into
  `out` by default. **List endpoints must set `rawBody: true`** and decode into
  `struct{ Result []T; Pagination Pagination }`, or pagination is silently lost.
- **Validate before the network call.** Struct tags + `validateStruct(params)`
  for bodies; guard required path params with `newValidationError("...")`.
  Tests assert that an invalid call never reaches the server.
- **Mutating calls (POST/PATCH) set an idempotency key** from
  `generateIdempotencyKey()` via `requestOptions.idempotencyKey`.
- **Prefix every exported type with its resource** (`PayoutOwner`, not
  `Owner`) — the flat package makes collisions easy. Reuse the shared `Amount`
  and `Pagination`; never redefine them.
- **List methods take `opts ...ListOption`** and must call `buildListQuery(opts)`
  first, returning its error before touching the network. New filters go in
  `listoptions.go` as `With<Resource><Param>` and must use the API's exact
  parameter name — several are inconsistent (`ussd_code` is snake_case,
  everything else is camelCase). Enum-valued filters validate locally via
  `setEnum`.
- **Accessors are methods, not fields**: `func (c *Client) Payout()
  *PayoutService`. Every method takes `ctx` first and returns `(*T, error)`, or
  plain `error` for deletes.
- **Errors are typed.** Return `*Error` / `*AuthenticationError` (401) /
  `*RateLimitError` (429, both set by `parseError`) / `*ValidationError`; all
  three unwrap to `*Error` so `errors.As` works for either. Never panic.

## API conformance

The authority is the canonical OpenAPI file for the pinned version, **not** the
rendered docs site:

```
https://raw.githubusercontent.com/monimesl/monime-developer-apis/refs/heads/main/versions/caph/2025-08-23/openapi.yaml
```

The published docs index (`llms.txt`) omits the `/v1/countries` endpoints that
the spec defines, so verifying against the site alone under-reports the API
surface. The SDK covers all 46 operations in that file; re-diff method+path
against it when adding a resource.

Facts that are easy to get wrong and are pinned by tests:

- The request id comes from **`Monime-Request-Id`** (`x-request-id` is only a
  fallback).
- The error body is `{"success":false,"messages":[],"error":{"code","reason",
  "message","details"}}` — not a top-level `message`. `parseError` still falls
  back to a top-level `message` for non-enveloped bodies (e.g. a proxy).
- Lists are cursor-paginated with `limit` (1–50, default 10) and `after`.
- `GET /banks` and `GET /momos` **require** a `country` query param, which is
  why those two `List`s take it as a positional argument.
- `Idempotency-Key` is required on POST only and capped at 64 chars; the
  40-char hex key from `generateIdempotencyKey` sits inside the documented
  25–64 range.

Webhook `Monime-Signature` verification is **not** implemented: the Monime HMAC
docs page is still an unpublished placeholder, so there is no spec to build
against. Don't guess an algorithm.

## Known asymmetry — do not "fix" it

Amount scaling is intentionally inconsistent because the API is:
`PaymentCode.Create` multiplies its major-unit `Amount` by 100
(`paymentcode.go`), while `Payout.Create` forwards the value **as-is** in minor
units. Both behaviors are pinned by tests. Currency is hard-coded `"SLE"` in
those create bodies.

## Tests

`newTestClient(t, srv)` lives in `transport_e2e_test.go` and wires a client to
an `httptest` server. Tests are in-package (`package monime`), so unexported
helpers are fair game. Any new resource or endpoint needs both a unit test and
an e2e test.

## Credentials

`New()` falls back to `MONIME_SPACE_ID`, `MONIME_ACCESS_TOKEN`,
`MONIME_VERSION` via `os.Getenv` and errors if space id or token are missing.
The SDK deliberately does **not** load `.env` — don't add a dotenv dependency.
