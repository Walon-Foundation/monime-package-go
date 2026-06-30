package monime

import (
	"context"
	"fmt"
	"net/http"
)

// CheckoutSessionService provides access to the checkout-sessions API.
type CheckoutSessionService struct {
	client *Client
}

// CheckoutSession returns the checkout-sessions service.
func (c *Client) CheckoutSession() *CheckoutSessionService {
	return &CheckoutSessionService{client: c}
}

const checkoutSessionPath = "/checkout-sessions"

// CreateCheckoutSessionParams holds the inputs for creating a checkout session
// used to host a payment page.
type CreateCheckoutSessionParams struct {
	Name       string `validate:"required"`
	Amount     int64  `validate:"gt=0"`
	Quantity   int    `validate:"gt=0"`
	SuccessURL string `validate:"required,url"`
	CancelURL  string `validate:"required,url"`
	// Description, FinancialAccountID, PrimaryColor and Images are optional.
	Description        string
	FinancialAccountID string
	PrimaryColor       string
	Images             []string
}

// CheckoutSession is the checkout-session resource returned by the API.
type CheckoutSession struct {
	ID                 string                         `json:"id"`
	Status             string                         `json:"status"`
	Name               string                         `json:"name"`
	OrderNumber        string                         `json:"orderNumber"`
	Reference          string                         `json:"reference"`
	Description        string                         `json:"description"`
	RedirectURL        string                         `json:"redirectUrl"`
	CancelURL          string                         `json:"cancelUrl"`
	SuccessURL         string                         `json:"successUrl"`
	LineItems          CheckoutSessionLineItems       `json:"lineItems"`
	FinancialAccountID string                         `json:"financialAccountId"`
	BrandingOptions    CheckoutSessionBrandingOptions `json:"brandingOptions"`
	ExpireTime         string                         `json:"expireTime"`
	CreateTime         string                         `json:"createTime"`
	OwnershipGraph     CheckoutSessionOwnershipGraph  `json:"ownershipGraph"`
	Metadata           map[string]any                 `json:"metadata,omitempty"`
}

type CheckoutSessionLineItems struct {
	Data []CheckoutSessionLineItem `json:"data"`
}

type CheckoutSessionLineItem struct {
	Type        string   `json:"type"`
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Price       Amount   `json:"price"`
	Quantity    int      `json:"quantity"`
	Reference   string   `json:"reference"`
	Description string   `json:"description"`
	Images      []string `json:"images"`
}

type CheckoutSessionBrandingOptions struct {
	PrimaryColor string `json:"primaryColor"`
}

type CheckoutSessionOwnershipGraph struct {
	Owner CheckoutSessionOwner `json:"owner"`
}

type CheckoutSessionOwner struct {
	ID    string                `json:"id"`
	Type  string                `json:"type"`
	Owner *CheckoutSessionOwner `json:"owner,omitempty"`
}

// CheckoutSessionList is the paginated checkout-session list response.
type CheckoutSessionList struct {
	Result     []CheckoutSession `json:"result"`
	Pagination Pagination        `json:"pagination"`
}

// Create creates a new checkout session for a hosted payment page.
func (s *CheckoutSessionService) Create(ctx context.Context, params CreateCheckoutSessionParams) (*CheckoutSession, error) {
	if err := validateStruct(params); err != nil {
		return nil, err
	}

	body := map[string]any{
		"name":               params.Name,
		"description":        params.Description,
		"cancelUrl":          params.CancelURL,
		"successUrl":         params.SuccessURL,
		"callbackState":      nil,
		"reference":          nil,
		"financialAccountId": params.FinancialAccountID,
		"lineItems": []map[string]any{
			{
				"type":        "custom",
				"name":        params.Name,
				"price":       Amount{Currency: "SLE", Value: params.Amount},
				"quantity":    params.Quantity,
				"reference":   nil,
				"description": params.Description,
				"images":      params.Images,
			},
		},
		"paymentOptions": map[string]any{
			"card": map[string]any{"disable": false},
			"bank": map[string]any{
				"disable":          false,
				"enabledProviders": []string{"slb001", "slb004", "slb007"},
			},
			"momo": map[string]any{
				"disable":          false,
				"enabledProviders": []string{"m17", "m18"},
			},
			"wallet": map[string]any{
				"disable":          false,
				"enabledProviders": []string{"dw001"},
			},
		},
		"brandingOptions": map[string]any{
			"primaryColor": params.PrimaryColor,
		},
		"metadata": map[string]any{},
	}

	key, err := generateIdempotencyKey()
	if err != nil {
		return nil, &Error{Message: err.Error()}
	}

	var out CheckoutSession
	if err := s.client.do(ctx, requestOptions{
		method:         http.MethodPost,
		path:           checkoutSessionPath,
		body:           body,
		idempotencyKey: key,
		out:            &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns all checkout sessions.
func (s *CheckoutSessionService) List(ctx context.Context) (*CheckoutSessionList, error) {
	var out CheckoutSessionList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    checkoutSessionPath,
		out:     &out,
		rawBody: true,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Retrieve fetches a single checkout session by id.
func (s *CheckoutSessionService) Retrieve(ctx context.Context, checkoutSessionID string) (*CheckoutSession, error) {
	if checkoutSessionID == "" {
		return nil, newValidationError("checkoutSessionID is required")
	}
	var out CheckoutSession
	if err := s.client.do(ctx, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("%s/%s", checkoutSessionPath, checkoutSessionID),
		out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes (cancels) a checkout session by id.
func (s *CheckoutSessionService) Delete(ctx context.Context, checkoutSessionID string) error {
	if checkoutSessionID == "" {
		return newValidationError("checkoutSessionID is required")
	}
	return s.client.do(ctx, requestOptions{
		method: http.MethodDelete,
		path:   fmt.Sprintf("%s/%s", checkoutSessionPath, checkoutSessionID),
	})
}
