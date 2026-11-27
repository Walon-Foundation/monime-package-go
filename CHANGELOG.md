# Changelog

All notable changes to this SDK are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project uses
[Semantic Versioning](https://semver.org/). Until 1.0.0, a minor release may
break the API; each break is marked **BREAKING** with what to change.

## [Unreleased]

A conformance pass against Monime's published OpenAPI spec
(`https://docs.monime.io/apis/openapi.json`, release `caph.2025-08-23`) and its
guides, checked on 2026-11-27. Intended as v0.3.0.

### Added

- `WithIdempotencyKey(ctx, key)`: sets the `Idempotency-Key` for a create call,
  so a retry after a timeout sends the same key and Monime returns the first
  result instead of acting twice. Keys are 1 to 64 characters.
- `IdempotencyKeyFor(reference)` derives a deterministic UUIDv5 key from a
  business reference, as the Monime docs recommend. It uses the same namespace
  as the TypeScript SDK, so both give the same key.
- Webhook verification:
  - `VerifyWebhookSignature` and `ParseWebhookEvent` check the
    `Monime-Signature` header for HS256 (shared secret) and ES256 (PEM P-256
    public key) webhooks, with the documented 300-second and 60-second
    freshness window.
  - `WebhookEvent` is the decoded envelope, and `Event…` constants cover the
    caph event names.
- Financial accounts: `Reference` on create, and `Retrieve(ctx, id,
  WithFinancialAccountBalance(true))` to include the balance.
- Payouts and internal transfers take a `Currency` (`SLE` by default, or `USD`).
- Payouts:
  - constants for statuses, failure codes and the provider ids in the spec;
  - `metadata` on fees and ownership-graph owners.
- Payments: `PaymentChannel` now carries:
  - `Provider`, `Reference`, `Fingerprint`, `Metadata`;
  - `PhoneNumber` (momo), `AccountNumber` (bank), `WalletID` (wallet);
  - `Scheme` and `Last4` (card).
- USSD OTPs: `VerificationMessage`, `Duration` and `Metadata` on create.
- Webhooks:
  - `APIRelease` on update;
  - the ES256 `PublicKey` in responses;
  - constants for signing methods and API releases.
- Status constants for payment codes, checkout sessions, internal transfers,
  receipts, payments and USSD OTPs.
- Metadata is validated before sending on every request: at most 64 pairs, the
  key format and reserved prefixes, and string values of at most 100
  characters.
- Every `Update(..., map[string]any)` now rejects fields that the endpoint
  doesn't accept, and checks their length limits.

### Changed

- **BREAKING:** `CreatePaymentCodeParams` mirrors `POST /v1/payment-codes`.
  - `Amount` is an `*Amount` in **minor units**, sent as-is. Before, it was a
    whole-leone `int64` multiplied by 100, which made cents impossible.
  - `PaymentName` is now `Name` (the code's own name). The old `Name` (the
    payer) is now `CustomerName`, and `PhoneNumber` is now
    `AuthorizedPhoneNumber`, which is optional.
  - New fields: `Mode`, `Enable`, `Duration`, `Reference`,
    `AuthorizedProviders`, `RecurrentPaymentTarget`, `FinancialAccountID`,
    `Metadata`. Nothing is hard-coded any more; before, every code was
    recurrent, 1h30m, with an empty reference and metadata.
  - Migration: `PaymentName: "X", Amount: 200, Name: "Y", PhoneNumber: "076…"`
    becomes `Name: "X", Amount: &Amount{Value: 20000}, CustomerName: "Y",
    AuthorizedPhoneNumber: "+23276…"`. Add `Mode: PaymentCodeModeRecurrent` and
    `Duration: "1h30m"` only if you relied on them.
- **BREAKING:** `CreateCheckoutSessionParams` mirrors `POST
  /v1/checkout-sessions`.
  - It takes `LineItems` (1 to 16; each has a name, a unit `Price` in minor
    units, a quantity, a reference, a description and up to three images)
    instead of `Amount`, `Quantity` and `Images`.
  - `PrimaryColor` moved to `BrandingOptions`.
  - New fields: `Reference`, `CallbackState`, `PaymentOptions`, `Metadata`.
  - `SuccessURL` and `CancelURL` are optional.
- **BREAKING:** `ProviderKyc().Retrieve(ctx, providerID, accountID)`. The API
  requires the `accountId` query, so every call made without it was malformed.
- **BREAKING:** `Receipt().Redeem` takes `RedeemReceiptParams` and returns a
  `ReceiptRedemption{Redeem, Receipt}`, the shape the API returns.
- **BREAKING:** `UpdateWebhookParams` no longer has `VerificationMethod`, which
  `PATCH /v1/webhooks/{id}` does not accept. Create a new webhook to change how
  deliveries are signed.
- `WebhookVerificationMethod.Secret` is required for HS256 (32 to 256
  characters) and must be empty for ES256.
- `Monime-Version` is sent by default as `caph.2025-08-23` (`DefaultVersion`);
  `WithVersion` and `MONIME_VERSION` still override it.
- Payouts: `SourceAccount` is optional. When it's left out, the API pays from
  the Space's main account.
- `Pagination.Total` is deprecated. The API returns `count` and `next` only, so
  it was always zero.

### Fixed

- Creates used a random `Idempotency-Key` made inside the call, so a retried
  payout could be paid twice. See `WithIdempotencyKey` above.
- Payout destinations:
  - now send only the fields their type allows;
  - require `accountNumber` for bank and `phoneNumber` for momo;
  - never send the read-only `transactionReference`.
- Optional fields left unset are omitted instead of being sent as `""`, `{}` or
  `null`:
  - financial accounts: `description`, `metadata`;
  - payouts: `metadata`;
  - internal transfers: `description`, `metadata`;
  - checkout sessions: `financialAccountId`, `reference`, `callbackState`,
    `metadata`, `paymentOptions`, `brandingOptions`;
  - payment codes: every optional field.
- Webhooks: `apiRelease`, which the API requires, is always sent (default
  `caph`). ES256 webhooks are created without a secret.
- The webhooks example subscribed to `payment.completed`, which is not a caph
  event. It now uses real event names and verifies deliveries.

## [0.2.0] - 2026-09-06

Aligned with the `caph.2025-08-23` API.

### Added

- `ListOption` cursor paging (`WithLimit`, `WithAfter`) on every list endpoint,
  plus per-resource filters.
- The countries resource.
- A typed `RateLimitError` on 429, with `RetryAfter` and `Limit`.
- Query parameters in the transport.
- `CLAUDE.md` with API conformance notes, and docs for pagination, filtering and
  the error types.

### Fixed

- The request id is read from `Monime-Request-Id`.
- API errors are parsed from the documented error envelope.
- The bank and momo lists send the required `country`.

## [0.1.0] - 2026-07-01

Initial release.

### Added

- The client (options with environment fallback), the transport, typed errors
  and validation.
- Payment codes, payments, payouts, financial accounts, financial transactions,
  checkout sessions, USSD OTPs, provider KYC, receipts, webhooks, financial
  providers (bank and momo) and internal transfers, with unit and `httptest`
  end-to-end tests.
- README, CONTRIBUTING, LICENSE, SECURITY and runnable examples.

[Unreleased]: https://github.com/Walon-Foundation/monime-package-go/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/Walon-Foundation/monime-package-go/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/Walon-Foundation/monime-package-go/releases/tag/v0.1.0
