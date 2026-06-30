# Contributing to monime-package-go

Thanks for your interest in contributing to **monime-package-go**, an unofficial
Go SDK for the [Monime](https://monime.io) API. Bug fixes, new endpoints,
documentation, and ideas are all welcome.

This guide is specific to **this** project — its layout, conventions, and the
workflow that keeps the SDK consistent.

---

## Getting Started

1. **Fork & clone**

   ```bash
   git clone https://github.com/Walon-Foundation/monime-package-go.git
   cd monime-package-go
   ```

2. **Requirements**

   - Go **1.25+** (see the `go` directive in `go.mod`).
   - That's it — the only runtime dependency is
     `github.com/go-playground/validator/v10`. Modules are fetched on first
     `go build`/`go test`.

3. **Build & test**

   ```bash
   go build ./...
   go vet ./...
   go test ./...        # add -race in CI / before pushing
   ```

---

## Project Layout

The SDK is a single flat package (`package monime`), one file per resource.

| File | Responsibility |
| --- | --- |
| `client.go` | `Client`, functional options, env fallback |
| `transport.go` | the one `do()` that builds requests, unwraps the `result` envelope, and maps errors |
| `errors.go` | `Error`, `AuthenticationError`, `ValidationError` |
| `validate.go` | shared `go-playground/validator` instance + `validateStruct` |
| `idempotency.go` | `crypto/rand` idempotency-key generator |
| `types.go` | shared `Amount`, `Pagination`, `MonimeVersion` |
| `<resource>.go` | one resource: its service, request/response types, and methods |
| `<resource>_test.go` | unit tests (validation, body construction) |
| `<resource>_e2e_test.go` | `httptest` end-to-end tests |

---

## Adding or Changing a Resource

New endpoints should match the official docs at <https://docs.monime.io/apis>.
Follow the patterns the existing resources already use:

1. **Accessor method** on `*Client`, returning a small service struct:

   ```go
   type WidgetService struct{ client *Client }

   func (c *Client) Widget() *WidgetService { return &WidgetService{client: c} }
   ```

2. **Context first.** Every method takes `ctx context.Context` as its first
   argument and returns `(*T, error)` (or `error` for deletes).

3. **Validate before the network call.** Use struct tags + `validateStruct(...)`
   for request bodies; guard required path params and return
   `newValidationError("...")`. Validation must short-circuit before any HTTP
   request.

4. **Mutating calls** (POST/PATCH) generate an idempotency key via
   `generateIdempotencyKey()` and pass it through `requestOptions`.

5. **List endpoints** set `rawBody: true` and decode into a
   `struct{ Result []T; Pagination Pagination }`, because Monime returns
   `result` and `pagination` as siblings.

6. **Type names must be resource-prefixed** (e.g. `WidgetOwner`, not `Owner`).
   Everything lives in one package, so unprefixed names collide. Reuse the
   shared `Amount` and `Pagination` — don't redefine them.

7. **Tests are required** — both a unit test (validation / body shape) and an
   `httptest` e2e test (method, path, headers, body, response decoding, and the
   validation short-circuit).

---

## Pull Requests

- **One focused change per PR** (a single resource, endpoint, or fix). Keep it
  small and reviewable.
- Branch naming: `feat/<resource>`, `fix/<thing>`, or `docs/<thing>`.
- **Before opening a PR**, make sure these pass locally:

  ```bash
  gofmt -l .       # should print nothing
  go vet ./...
  go test ./...
  ```

- Describe what changed and link any relevant issue. If a change reflects a
  Monime API doc update, link the doc.
- CI (`.github/workflows/go.yml`) runs `go vet` and `go test -race` on every PR;
  it must be green before merge.

---

## Coding Conventions

- Run `gofmt` (or `goimports`) — formatting is not optional.
- Use clear, idiomatic Go names; exported identifiers get doc comments.
- Return typed errors (`*Error`, `*AuthenticationError`, `*ValidationError`) so
  callers can use `errors.As`. Never panic on a recoverable error.
- Keep each resource file focused and self-contained.

---

## Reporting Issues

Check the [issue tracker](https://github.com/Walon-Foundation/monime-package-go/issues)
first; if it's new, open an issue with a clear description, a minimal repro, and
your environment (Go version, OS). For **security** issues, do not open a public
issue — follow [SECURITY.md](SECURITY.md).

---

Thank you for contributing to **monime-package-go**!
