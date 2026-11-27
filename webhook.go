package monime

import (
	"context"
	"fmt"
	"net/http"
)

// WebhookService provides access to the webhooks API.
type WebhookService struct {
	client *Client
}

// Webhook returns the webhooks service.
func (c *Client) Webhook() *WebhookService {
	return &WebhookService{client: c}
}

const webhookPath = "/webhooks"

// Webhook signing methods and API releases.
const (
	WebhookVerificationHS256 = "HS256"
	WebhookVerificationES256 = "ES256"

	WebhookAPIReleaseCaph    = "caph"
	WebhookAPIReleaseSiriusB = "siriusb"
)

// WebhookVerificationMethod configures how Monime signs deliveries.
//
//   - HS256: HMAC-SHA256 with a shared Secret of 32 to 256 characters.
//   - ES256: ECDSA P-256. Monime generates the key pair; send no Secret, and
//     read the PEM PublicKey from the created webhook to verify deliveries.
//
// PublicKey is read-only: it is filled in responses and never sent.
type WebhookVerificationMethod struct {
	Type      string `json:"type" validate:"required,oneof=HS256 ES256"`
	Secret    string `json:"secret,omitempty" validate:"required_if=Type HS256,excluded_if=Type ES256,omitempty,min=32,max=256"`
	PublicKey string `json:"publicKey,omitempty"`
}

// requestBody is the verification method as the create body accepts it.
func (m *WebhookVerificationMethod) requestBody() map[string]any {
	body := map[string]any{"type": m.Type}
	if m.Type == WebhookVerificationHS256 {
		body["secret"] = m.Secret
	}
	return body
}

// CreateWebhookParams holds the inputs for POST /v1/webhooks. Name, URL and
// Events are required. APIRelease defaults to "caph". The remaining fields are
// optional and omitted from the request body when left at their zero value.
type CreateWebhookParams struct {
	Name               string                     `json:"name" validate:"required,min=1,max=100"`
	URL                string                     `json:"url" validate:"required,url,max=255"`
	Events             []string                   `json:"events" validate:"required,min=1,max=100"`
	Enabled            *bool                      `json:"enabled,omitempty"`
	APIRelease         string                     `json:"apiRelease,omitempty" validate:"omitempty,oneof=caph siriusb"`
	Headers            map[string]string          `json:"headers,omitempty"`
	AlertEmails        []string                   `json:"alertEmails,omitempty" validate:"omitempty,max=2,dive,email"`
	VerificationMethod *WebhookVerificationMethod `json:"verificationMethod,omitempty"`
	Metadata           map[string]any             `json:"metadata,omitempty"`
}

// UpdateWebhookParams holds the inputs for PATCH /v1/webhooks/{id}. Every field
// is optional and only the supplied fields are sent. The verification method
// can't be changed: create a new webhook to rotate keys or secrets.
type UpdateWebhookParams struct {
	Name        string            `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	URL         string            `json:"url,omitempty" validate:"omitempty,url,max=255"`
	Events      []string          `json:"events,omitempty" validate:"omitempty,min=1,max=100"`
	Enabled     *bool             `json:"enabled,omitempty"`
	APIRelease  string            `json:"apiRelease,omitempty" validate:"omitempty,oneof=caph siriusb"`
	Headers     map[string]string `json:"headers,omitempty"`
	AlertEmails []string          `json:"alertEmails,omitempty" validate:"omitempty,max=2,dive,email"`
	Metadata    map[string]any    `json:"metadata,omitempty"`
}

// Webhook is the webhook resource returned by the API.
type Webhook struct {
	ID                 string                    `json:"id"`
	Name               string                    `json:"name"`
	URL                string                    `json:"url"`
	Enabled            bool                      `json:"enabled"`
	Events             []string                  `json:"events"`
	APIRelease         string                    `json:"apiRelease"`
	VerificationMethod WebhookVerificationMethod `json:"verificationMethod"`
	Headers            map[string]string         `json:"headers"`
	AlertEmails        []string                  `json:"alertEmails"`
	CreateTime         string                    `json:"createTime"`
	UpdateTime         string                    `json:"updateTime"`
	Metadata           map[string]any            `json:"metadata,omitempty"`
}

// WebhookList is the paginated webhook list response.
type WebhookList struct {
	Result     []Webhook  `json:"result"`
	Pagination Pagination `json:"pagination"`
}

// Create creates a new webhook.
func (s *WebhookService) Create(ctx context.Context, params CreateWebhookParams) (*Webhook, error) {
	if err := validateStruct(params); err != nil {
		return nil, err
	}
	if params.VerificationMethod != nil {
		if err := validateStruct(params.VerificationMethod); err != nil {
			return nil, err
		}
	}
	if err := validateMetadata(params.Metadata, false); err != nil {
		return nil, err
	}

	body := map[string]any{
		"name":   params.Name,
		"url":    params.URL,
		"events": params.Events,
		// apiRelease is required by the API; caph is its default and current.
		"apiRelease": WebhookAPIReleaseCaph,
	}
	putIfSet(body, "apiRelease", params.APIRelease)
	if params.Enabled != nil {
		body["enabled"] = *params.Enabled
	}
	if len(params.Headers) > 0 {
		body["headers"] = params.Headers
	}
	if len(params.AlertEmails) > 0 {
		body["alertEmails"] = params.AlertEmails
	}
	if params.VerificationMethod != nil {
		body["verificationMethod"] = params.VerificationMethod.requestBody()
	}
	if params.Metadata != nil {
		body["metadata"] = params.Metadata
	}

	key, err := idempotencyKey(ctx)
	if err != nil {
		return nil, err
	}

	var out Webhook
	if err := s.client.do(ctx, requestOptions{
		method:         http.MethodPost,
		path:           webhookPath,
		body:           body,
		idempotencyKey: key,
		out:            &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Retrieve fetches a single webhook by id.
func (s *WebhookService) Retrieve(ctx context.Context, webhookID string) (*Webhook, error) {
	if webhookID == "" {
		return nil, newValidationError("webhookID is required")
	}
	var out Webhook
	if err := s.client.do(ctx, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("%s/%s", webhookPath, webhookID),
		out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns a page of webhooks. Use WithLimit and WithAfter to page through
// the collection.
func (s *WebhookService) List(ctx context.Context, opts ...ListOption) (*WebhookList, error) {
	query, err := buildListQuery(opts)
	if err != nil {
		return nil, err
	}

	var out WebhookList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    webhookPath,
		query:   query,
		out:     &out,
		rawBody: true,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update applies a partial update to a webhook.
func (s *WebhookService) Update(ctx context.Context, webhookID string, params UpdateWebhookParams) (*Webhook, error) {
	if webhookID == "" {
		return nil, newValidationError("webhookID is required")
	}
	if err := validateStruct(params); err != nil {
		return nil, err
	}
	if err := validateMetadata(params.Metadata, true); err != nil {
		return nil, err
	}

	key, err := idempotencyKey(ctx)
	if err != nil {
		return nil, err
	}

	var out Webhook
	if err := s.client.do(ctx, requestOptions{
		method:         http.MethodPatch,
		path:           fmt.Sprintf("%s/%s", webhookPath, webhookID),
		body:           params,
		idempotencyKey: key,
		out:            &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a webhook by id.
func (s *WebhookService) Delete(ctx context.Context, webhookID string) error {
	if webhookID == "" {
		return newValidationError("webhookID is required")
	}
	return s.client.do(ctx, requestOptions{
		method: http.MethodDelete,
		path:   fmt.Sprintf("%s/%s", webhookPath, webhookID),
	})
}
