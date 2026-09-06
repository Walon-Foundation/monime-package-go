package monime

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ListOption sets a query parameter on a list request.
//
// WithLimit and WithAfter apply to every list endpoint. Resource-prefixed
// options (WithPayoutStatus, WithPaymentOrderNumber, ...) are only meaningful
// on their own resource's List; passing one to a different endpoint sends a
// parameter that endpoint ignores.
//
// Options with an empty value are dropped, so a cursor can be threaded through
// unconditionally:
//
//	page, err := client.Payout().List(ctx, monime.WithAfter(prev.Pagination.Next))
type ListOption func(*listQuery)

// listQuery accumulates query parameters and the first validation failure.
type listQuery struct {
	values url.Values
	err    error
}

// set records a parameter, ignoring empty values.
func (q *listQuery) set(key, value string) {
	if value == "" {
		return
	}
	q.values.Set(key, value)
}

// setEnum records a parameter only when value is one of allowed, failing the
// whole query otherwise.
func (q *listQuery) setEnum(key, value string, allowed ...string) {
	if value == "" {
		return
	}
	for _, candidate := range allowed {
		if value == candidate {
			q.values.Set(key, value)
			return
		}
	}
	q.fail(fmt.Sprintf("%s must be one of %s", key, strings.Join(allowed, ", ")))
}

// fail records the first validation failure; later ones are ignored.
func (q *listQuery) fail(message string) {
	if q.err == nil {
		q.err = newValidationError(message)
	}
}

// buildListQuery applies opts and reports the first option that failed local
// validation, so an invalid option short-circuits before any network call.
func buildListQuery(opts []ListOption) (url.Values, error) {
	q := &listQuery{values: url.Values{}}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(q)
	}
	if q.err != nil {
		return nil, q.err
	}
	return q.values, nil
}

// WithLimit sets the maximum number of items in a page. The API accepts 1-50
// and defaults to 10.
func WithLimit(limit int) ListOption {
	return func(q *listQuery) {
		if limit < 1 || limit > 50 {
			q.fail("limit must be between 1 and 50")
			return
		}
		q.values.Set("limit", strconv.Itoa(limit))
	}
}

// WithAfter sets the forward pagination cursor. Pass the Next field of the
// previous page's Pagination; an empty cursor is ignored, and Next is empty on
// the last page.
func WithAfter(cursor string) ListOption {
	return func(q *listQuery) { q.set("after", cursor) }
}

// --- Payment code ---------------------------------------------------------

// WithPaymentCodeUssdCode filters payment codes by their assigned USSD code.
func WithPaymentCodeUssdCode(ussdCode string) ListOption {
	return func(q *listQuery) { q.set("ussd_code", ussdCode) }
}

// WithPaymentCodeMode filters payment codes by usage mode: "one_time" or
// "recurrent".
func WithPaymentCodeMode(mode string) ListOption {
	return func(q *listQuery) { q.setEnum("mode", mode, "one_time", "recurrent") }
}

// WithPaymentCodeStatus filters payment codes by status: "pending",
// "cancelled", "processing", "expired" or "completed".
func WithPaymentCodeStatus(status string) ListOption {
	return func(q *listQuery) {
		q.setEnum("status", status, "pending", "cancelled", "processing", "expired", "completed")
	}
}

// --- Payment --------------------------------------------------------------

// WithPaymentOrderNumber filters payments by order number.
func WithPaymentOrderNumber(orderNumber string) ListOption {
	return func(q *listQuery) { q.set("orderNumber", orderNumber) }
}

// WithPaymentFinancialAccountID filters payments by the credited financial
// account.
func WithPaymentFinancialAccountID(accountID string) ListOption {
	return func(q *listQuery) { q.set("financialAccountId", accountID) }
}

// WithPaymentFinancialTransactionReference filters payments by the reference
// grouping their financial transactions.
func WithPaymentFinancialTransactionReference(reference string) ListOption {
	return func(q *listQuery) { q.set("financialTransactionReference", reference) }
}

// --- Payout ---------------------------------------------------------------

// WithPayoutStatus filters payouts by status: "pending", "processing",
// "failed" or "completed".
func WithPayoutStatus(status string) ListOption {
	return func(q *listQuery) {
		q.setEnum("status", status, "pending", "processing", "failed", "completed")
	}
}

// WithPayoutSourceAccount filters payouts by the originating financial account.
func WithPayoutSourceAccount(accountID string) ListOption {
	return func(q *listQuery) { q.set("sourceFinancialAccountId", accountID) }
}

// WithPayoutSourceTransactionReference filters payouts by the reference of the
// transactions debited from the source account.
func WithPayoutSourceTransactionReference(reference string) ListOption {
	return func(q *listQuery) { q.set("sourceTransactionReference", reference) }
}

// WithPayoutDestinationTransactionReference filters payouts by the reference
// assigned at the destination provider.
func WithPayoutDestinationTransactionReference(reference string) ListOption {
	return func(q *listQuery) { q.set("destinationTransactionReference", reference) }
}

// --- Financial account ----------------------------------------------------

// WithFinancialAccountUvan filters financial accounts by Universal Virtual
// Account Number.
func WithFinancialAccountUvan(uvan string) ListOption {
	return func(q *listQuery) { q.set("uvan", uvan) }
}

// WithFinancialAccountReference filters financial accounts by the external
// reference linking them to your own system.
func WithFinancialAccountReference(reference string) ListOption {
	return func(q *listQuery) { q.set("reference", reference) }
}

// WithFinancialAccountBalance asks the API to include each account's balance in
// the response.
func WithFinancialAccountBalance(withBalance bool) ListOption {
	return func(q *listQuery) { q.values.Set("withBalance", strconv.FormatBool(withBalance)) }
}

// --- Financial transaction ------------------------------------------------

// WithFinancialTransactionAccountID filters transactions by financial account.
func WithFinancialTransactionAccountID(accountID string) ListOption {
	return func(q *listQuery) { q.set("financialAccountId", accountID) }
}

// WithFinancialTransactionReference filters transactions by the
// Monime-assigned reference that groups them.
func WithFinancialTransactionReference(reference string) ListOption {
	return func(q *listQuery) { q.set("reference", reference) }
}

// WithFinancialTransactionType filters transactions by direction: "credit" for
// incoming funds, "debit" for outgoing.
func WithFinancialTransactionType(transactionType string) ListOption {
	return func(q *listQuery) { q.setEnum("type", transactionType, "credit", "debit") }
}

// --- Internal transfer ----------------------------------------------------

// WithInternalTransferStatus filters internal transfers by status: "pending",
// "processing", "failed" or "completed".
func WithInternalTransferStatus(status string) ListOption {
	return func(q *listQuery) {
		q.setEnum("status", status, "pending", "processing", "failed", "completed")
	}
}

// WithInternalTransferSourceAccount filters internal transfers by the debited
// financial account.
func WithInternalTransferSourceAccount(accountID string) ListOption {
	return func(q *listQuery) { q.set("sourceFinancialAccountId", accountID) }
}

// WithInternalTransferDestinationAccount filters internal transfers by the
// credited financial account.
func WithInternalTransferDestinationAccount(accountID string) ListOption {
	return func(q *listQuery) { q.set("destinationFinancialAccountId", accountID) }
}

// WithInternalTransferTransactionReference filters internal transfers by the
// reference grouping their financial transactions.
func WithInternalTransferTransactionReference(reference string) ListOption {
	return func(q *listQuery) { q.set("financialTransactionReference", reference) }
}
