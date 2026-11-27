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

// Checkout session statuses.
const (
	CheckoutSessionStatusPending   = "pending"
	CheckoutSessionStatusCancelled = "cancelled"
	CheckoutSessionStatusExpired   = "expired"
	CheckoutSessionStatusCompleted = "completed"
)

// CreateCheckoutSessionParams holds the inputs for creating a checkout session,
// as POST /v1/checkout-sessions defines them. Optional fields left empty are
// not sent.
type CreateCheckoutSessionParams struct {
	// Name is the order's name, up to 150 characters.
	Name        string `validate:"required,max=150"`
	Description string `validate:"max=1000"`
	// SuccessURL and CancelURL are where the payer is sent afterwards; both are
	// optional.
	SuccessURL string `validate:"omitempty,url,max=255"`
	CancelURL  string `validate:"omitempty,url,max=255"`
	// CallbackState is echoed back to you, up to 255 characters.
	CallbackState string `validate:"max=255"`
	// Reference is your own order identifier, up to 255 characters.
	Reference string `validate:"max=255"`
	// FinancialAccountID is the account the money goes to; when empty the
	// Space's main account is used.
	FinancialAccountID string `validate:"max=100"`
	// LineItems lists 1 to 16 items, all in one currency.
	LineItems       []CheckoutSessionLineItemParams `validate:"required,min=1,max=16,dive"`
	PaymentOptions  *CheckoutSessionPaymentOptions
	BrandingOptions *CheckoutSessionBrandingOptions
	Metadata        map[string]any
}

// CheckoutSessionLineItemParams is one item on the order. Price is the unit
// price in minor units; Currency defaults to SLE.
type CheckoutSessionLineItemParams struct {
	Name  string `validate:"required,max=100"`
	Price Amount
	// Quantity is 1 to 100,000; zero means 1, the API's default.
	Quantity    int    `validate:"omitempty,min=1,max=100000"`
	Reference   string `validate:"max=100"`
	Description string `validate:"max=100"`
	// Images are up to three picture URLs.
	Images []string `validate:"max=3,dive,url"`
}

// CheckoutSessionPaymentOptions turns payment methods on or off on the hosted
// page. A nil option leaves that method at the API's default (enabled).
type CheckoutSessionPaymentOptions struct {
	Card   *CheckoutSessionCardOption     `json:"card,omitempty"`
	Bank   *CheckoutSessionProviderOption `json:"bank,omitempty"`
	Momo   *CheckoutSessionProviderOption `json:"momo,omitempty"`
	Wallet *CheckoutSessionProviderOption `json:"wallet,omitempty"`
}

// CheckoutSessionCardOption disables card payments when Disable is true.
type CheckoutSessionCardOption struct {
	Disable bool `json:"disable"`
}

// CheckoutSessionProviderOption disables a method, or narrows it to some
// providers (for example bank "slb001", momo "m17", wallet "dw001").
type CheckoutSessionProviderOption struct {
	Disable           bool     `json:"disable"`
	EnabledProviders  []string `json:"enabledProviders,omitempty"`
	DisabledProviders []string `json:"disabledProviders,omitempty"`
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
	PrimaryColor string `json:"primaryColor,omitempty" validate:"max=255"`
}

type CheckoutSessionOwnershipGraph struct {
	Owner CheckoutSessionOwner `json:"owner"`
}

type CheckoutSessionOwner struct {
	ID       string                `json:"id"`
	Type     string                `json:"type"`
	Metadata map[string]any        `json:"metadata,omitempty"`
	Owner    *CheckoutSessionOwner `json:"owner,omitempty"`
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

	if err := validateMetadata(params.Metadata, false); err != nil {
		return nil, err
	}
	if params.BrandingOptions != nil {
		if err := validateStruct(params.BrandingOptions); err != nil {
			return nil, err
		}
	}

	items := make([]map[string]any, 0, len(params.LineItems))
	var currency Currency
	for i, item := range params.LineItems {
		price, err := normalizeAmount(fmt.Sprintf("lineItems[%d].price", i), &item.Price)
		if err != nil {
			return nil, err
		}
		if i == 0 {
			currency = Currency(price.Currency)
		} else if Currency(price.Currency) != currency {
			return nil, newValidationError("all line items must be in the same currency")
		}
		line := map[string]any{"type": "custom", "name": item.Name, "price": price}
		if item.Quantity > 0 {
			line["quantity"] = item.Quantity
		}
		putIfSet(line, "reference", item.Reference)
		putIfSet(line, "description", item.Description)
		if len(item.Images) > 0 {
			line["images"] = item.Images
		}
		items = append(items, line)
	}

	body := map[string]any{"name": params.Name, "lineItems": items}
	putIfSet(body, "description", params.Description)
	putIfSet(body, "successUrl", params.SuccessURL)
	putIfSet(body, "cancelUrl", params.CancelURL)
	putIfSet(body, "callbackState", params.CallbackState)
	putIfSet(body, "reference", params.Reference)
	putIfSet(body, "financialAccountId", params.FinancialAccountID)
	if params.PaymentOptions != nil {
		body["paymentOptions"] = params.PaymentOptions
	}
	if params.BrandingOptions != nil && params.BrandingOptions.PrimaryColor != "" {
		body["brandingOptions"] = params.BrandingOptions
	}
	if params.Metadata != nil {
		body["metadata"] = params.Metadata
	}

	key, err := idempotencyKey(ctx)
	if err != nil {
		return nil, err
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

// List returns a page of checkout sessions. Use WithLimit and WithAfter to
// page through the collection.
func (s *CheckoutSessionService) List(ctx context.Context, opts ...ListOption) (*CheckoutSessionList, error) {
	query, err := buildListQuery(opts)
	if err != nil {
		return nil, err
	}

	var out CheckoutSessionList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    checkoutSessionPath,
		query:   query,
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
