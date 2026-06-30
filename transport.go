package monime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// requestOptions describes a single API call made through Client.do.
type requestOptions struct {
	method         string
	path           string
	body           any    // marshalled to JSON when non-nil
	idempotencyKey string // set on the Idempotency-Key header when non-empty
	out            any    // pointer the response is decoded into; may be nil

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

	req, err := http.NewRequestWithContext(ctx, opts.method, c.baseURL+opts.path, bodyReader)
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

	requestID := res.Header.Get("x-request-id")

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

// parseError converts a non-2xx response into a typed error.
func parseError(res *http.Response, requestID string) error {
	body, _ := io.ReadAll(res.Body)

	message := fmt.Sprintf("request failed with status %d", res.StatusCode)
	var details any
	if len(body) > 0 {
		var parsed struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &parsed) == nil && parsed.Message != "" {
			message = parsed.Message
		}
		var raw any
		if json.Unmarshal(body, &raw) == nil {
			details = raw
		}
	}

	base := &Error{
		Message:   message,
		Status:    res.StatusCode,
		RequestID: requestID,
		Details:   details,
	}

	if res.StatusCode == http.StatusUnauthorized {
		return newAuthenticationError(base)
	}
	return base
}
