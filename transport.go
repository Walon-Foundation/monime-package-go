package monime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// requestOptions describes a single API call made through Client.do.
type requestOptions struct {
	method         string
	path           string
	query          url.Values // appended to the URL when non-empty
	body           any        // marshalled to JSON when non-nil
	idempotencyKey string     // set on the Idempotency-Key header when non-empty
	out            any        // pointer the response is decoded into; may be nil

	// rawBody controls how the success body is decoded into out. Monime wraps
	// responses as {"success", "messages", "result", "pagination"}. For
	// single-object endpoints (the default) the "result" object is unwrapped
	// into out. List endpoints set rawBody so the whole envelope is decoded
	// into out, preserving the sibling "pagination" field alongside "result".
	rawBody bool
}

// do executes an API request: it builds the request with the standard Monime
// headers, sends it, unwraps the {"result": ...} envelope into opts.out, and
// maps non-2xx responses to typed errors (*AuthenticationError for 401, *Error
// otherwise).
func (c *Client) do(ctx context.Context, opts requestOptions) error {
	var bodyReader io.Reader
	if opts.body != nil {
		encoded, err := json.Marshal(opts.body)
		if err != nil {
			return &Error{Message: fmt.Sprintf("failed to encode request body: %v", err)}
		}
		bodyReader = bytes.NewReader(encoded)
	}

	endpoint := c.baseURL + opts.path
	if len(opts.query) > 0 {
		endpoint += "?" + opts.query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, opts.method, endpoint, bodyReader)
	if err != nil {
		return &Error{Message: err.Error()}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Monime-Space-Id", c.spaceID)
	if c.version != "" {
		req.Header.Set("Monime-Version", string(c.version))
	}
	if opts.idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", opts.idempotencyKey)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return &Error{Message: err.Error()}
	}
	defer res.Body.Close()

	requestID := requestIDOf(res)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return parseError(res, requestID)
	}

	if res.StatusCode == http.StatusNoContent || opts.out == nil {
		return nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return &Error{Message: fmt.Sprintf("failed to read response: %v", err), RequestID: requestID}
	}

	// List endpoints want the whole envelope (result + pagination). Single
	// endpoints unwrap the "result" object when present, otherwise decode the
	// body directly.
	payload := body
	if !opts.rawBody {
		var env struct {
			Result json.RawMessage `json:"result"`
		}
		if json.Unmarshal(body, &env) == nil && len(env.Result) > 0 {
			payload = env.Result
		}
	}

	if err := json.Unmarshal(payload, opts.out); err != nil {
		return &Error{Message: fmt.Sprintf("failed to decode response: %v", err), RequestID: requestID}
	}
	return nil
}

// requestIDOf reads the request id Monime returns for tracing. The documented
// header is Monime-Request-Id; x-request-id is accepted as a fallback.
func requestIDOf(res *http.Response) string {
	if id := res.Header.Get("Monime-Request-Id"); id != "" {
		return id
	}
	return res.Header.Get("x-request-id")
}

// errorEnvelope is the error body Monime returns across all endpoints:
//
//	{"success": false, "messages": [], "error": {"code", "reason", "message", "details"}}
type errorEnvelope struct {
	Error struct {
		Code    int    `json:"code"`
		Reason  string `json:"reason"`
		Message string `json:"message"`
		Details any    `json:"details"`
	} `json:"error"`

	// Message catches error bodies that are not wrapped in the envelope, such
	// as those produced by a proxy in front of the API.
	Message string `json:"message"`
}

// parseError converts a non-2xx response into a typed error: *RateLimitError
// for 429, *AuthenticationError for 401, and *Error otherwise.
func parseError(res *http.Response, requestID string) error {
	body, _ := io.ReadAll(res.Body)

	base := &Error{
		Message:   fmt.Sprintf("request failed with status %d", res.StatusCode),
		Status:    res.StatusCode,
		RequestID: requestID,
	}

	if len(body) > 0 {
		var env errorEnvelope
		if json.Unmarshal(body, &env) == nil {
			switch {
			case env.Error.Message != "":
				base.Message = env.Error.Message
			case env.Message != "":
				base.Message = env.Message
			}
			base.Code = env.Error.Code
			base.Reason = env.Error.Reason
			base.Details = env.Error.Details
		}
		if base.Details == nil {
			// Not the documented envelope: keep the whole body as details so
			// nothing is lost.
			var raw any
			if json.Unmarshal(body, &raw) == nil {
				base.Details = raw
			}
		}
	}

	switch res.StatusCode {
	case http.StatusUnauthorized:
		return newAuthenticationError(base)
	case http.StatusTooManyRequests:
		return newRateLimitError(base, res.Header)
	}
	return base
}

// newRateLimitError decorates a 429 with its Retry-After delay and the
// Monime-Rate-Limit dimension that was tripped.
func newRateLimitError(base *Error, header http.Header) *RateLimitError {
	err := &RateLimitError{Err: base, Limit: header.Get("Monime-Rate-Limit")}
	if seconds, convErr := strconv.Atoi(header.Get("Retry-After")); convErr == nil && seconds > 0 {
		err.RetryAfter = time.Duration(seconds) * time.Second
	}
	return err
}
