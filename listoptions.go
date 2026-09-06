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
