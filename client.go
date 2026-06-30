package monime

import (
	"net/http"
	"os"
	"strings"
)

const defaultBaseURL = "https://api.monime.io/v1"

// Client is the entry point to the Monime API. Construct it with New and access
// resources through its fields, e.g. client.PaymentCode.Create(ctx, ...).
type Client struct {
	accessToken string
	spaceID     string
	version     MonimeVersion
	baseURL     string
	httpClient  *http.Client
}

// Option configures a Client in New.
type Option func(*Client)

// WithAccessToken sets the bearer access token. Falls back to the
// MONIME_ACCESS_TOKEN environment variable when unset.
func WithAccessToken(token string) Option {
	return func(c *Client) { c.accessToken = token }
}

// WithSpaceID sets the Monime space id. Falls back to the MONIME_SPACE_ID
// environment variable when unset.
func WithSpaceID(spaceID string) Option {
	return func(c *Client) { c.spaceID = spaceID }
}

// WithVersion pins the Monime-Version header. Falls back to the MONIME_VERSION
// environment variable when unset.
func WithVersion(version MonimeVersion) Option {
	return func(c *Client) { c.version = version }
}

// WithBaseURL overrides the API base URL. Primarily useful for tests.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(baseURL, "/") }
}

// WithHTTPClient sets the underlying *http.Client. Primarily useful for tests
// or to control timeouts and transport.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) { c.httpClient = httpClient }
}

// New creates a Client from the given options, falling back to environment
// variables for any unset credentials. It returns an error if the space id or
// access token are missing.
func New(opts ...Option) (*Client, error) {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(c)
	}

	if c.accessToken == "" {
		c.accessToken = os.Getenv("MONIME_ACCESS_TOKEN")
	}
	if c.spaceID == "" {
		c.spaceID = os.Getenv("MONIME_SPACE_ID")
	}
	if c.version == "" {
		c.version = MonimeVersion(os.Getenv("MONIME_VERSION"))
	}

	if strings.TrimSpace(c.spaceID) == "" {
		return nil, &Error{Message: "configuration error: 'spaceID' is missing. " +
			"Pass monime.WithSpaceID(\"...\") or set MONIME_SPACE_ID. " +
			"Find your Space ID in the Monime Dashboard (https://monime.io)."}
	}
	if strings.TrimSpace(c.accessToken) == "" {
		return nil, &Error{Message: "configuration error: 'accessToken' is missing. " +
			"Pass monime.WithAccessToken(\"...\") or set MONIME_ACCESS_TOKEN. " +
			"Generate an Access Token in the Monime Dashboard settings."}
	}

	return c, nil
}
