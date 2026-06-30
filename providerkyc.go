package monime

import (
	"context"
	"fmt"
	"net/http"
)

// ProviderKycService provides access to the provider-kyc API.
type ProviderKycService struct {
	client *Client
}

// ProviderKyc returns the provider-kyc service.
func (c *Client) ProviderKyc() *ProviderKycService {
	return &ProviderKycService{client: c}
}

const providerKycPath = "/provider-kyc"

// ProviderKyc is the provider-kyc resource returned by the API. It holds the
// KYC details Monime has on record for an account within a provider's system.
type ProviderKyc struct {
	Account  ProviderKycAccount  `json:"account"`
	Provider ProviderKycProvider `json:"provider"`
}

// ProviderKycAccount describes the account within the provider's ecosystem.
type ProviderKycAccount struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	HolderName string         `json:"holderName"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// ProviderKycProvider describes the payment provider.
type ProviderKycProvider struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// Retrieve fetches the provider KYC details for a single provider by id.
func (s *ProviderKycService) Retrieve(ctx context.Context, providerID string) (*ProviderKyc, error) {
	if providerID == "" {
		return nil, newValidationError("providerID is required")
	}
	var out ProviderKyc
	if err := s.client.do(ctx, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("%s/%s", providerKycPath, providerID),
		out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}
