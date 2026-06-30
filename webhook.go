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

// WebhookVerificationMethod configures how webhook payloads are signed so the
// receiver can verify their authenticity.
type WebhookVerificationMethod struct {
	Type   string `json:"type" validate:"required,oneof=HS256 ES256"`
	Secret string `json:"secret" validate:"required"`
}

// CreateWebhookParams holds the inputs for creating a webhook. It mirrors the
// CreateWebhookRequest body of the TypeScript SDK: name, url and events are
// required, the remaining fields are optional and omitted from the request
// body when left at their zero value.
type CreateWebhookParams struct {
	Name               string                     `json:"name" validate:"required"`
	URL                string                     `json:"url" validate:"required,url"`
	Events             []string                   `json:"events" validate:"required,min=1"`
	Enabled            *bool                      `json:"enabled,omitempty"`
	APIRelease         string                     `json:"apiRelease,omitempty"`
	Headers            map[string]string          `json:"headers,omitempty"`
	AlertEmails        []string                   `json:"alertEmails,omitempty" validate:"omitempty,dive,email"`
	VerificationMethod *WebhookVerificationMethod `json:"verificationMethod,omitempty"`
	Metadata           map[string]any             `json:"metadata,omitempty"`
}

// UpdateWebhookParams holds the inputs for updating a webhook. It mirrors the
// UpdateWebhookRequest body of the TypeScript SDK where every field is
// optional and only the supplied fields are sent.
type UpdateWebhookParams struct {
	Name               string                     `json:"name,omitempty" validate:"omitempty,min=1"`
	URL                string                     `json:"url,omitempty" validate:"omitempty,url"`
	Events             []string                   `json:"events,omitempty" validate:"omitempty,min=1"`
	Enabled            *bool                      `json:"enabled,omitempty"`
	Headers            map[string]string          `json:"headers,omitempty"`
	AlertEmails        []string                   `json:"alertEmails,omitempty" validate:"omitempty,dive,email"`
	VerificationMethod *WebhookVerificationMethod `json:"verificationMethod,omitempty"`
	Metadata           map[string]any             `json:"metadata,omitempty"`
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

	key, err := generateIdempotencyKey()
	if err != nil {
		return nil, &Error{Message: err.Error()}
	}

	var out Webhook
	if err := s.client.do(ctx, requestOptions{
		method:         http.MethodPost,
		path:           webhookPath,
		body:           params,
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

// List returns all webhooks.
func (s *WebhookService) List(ctx context.Context) (*WebhookList, error) {
	var out WebhookList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    webhookPath,
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

	key, err := generateIdempotencyKey()
	if err != nil {
		return nil, &Error{Message: err.Error()}
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
