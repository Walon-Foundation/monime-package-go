package monime

import (
	"context"
	"fmt"
	"net/http"
)

// FinancialProvider groups the read-only financial provider directories: the
// supported banks and mobile money (momo) providers.
type FinancialProvider struct {
	client *Client
}

// FinancialProvider returns the financial-provider group, used to access the
// bank and momo provider directories, e.g.
// client.FinancialProvider().Bank().List(ctx).
func (c *Client) FinancialProvider() *FinancialProvider {
	return &FinancialProvider{client: c}
}

// Bank returns the bank provider service.
func (f *FinancialProvider) Bank() *BankService {
	return &BankService{client: f.client}
}

// Momo returns the mobile money (momo) provider service.
func (f *FinancialProvider) Momo() *MomoService {
	return &MomoService{client: f.client}
}

const (
	bankPath = "/banks"
	momoPath = "/momos"
)

// BankFeature describes a single capability of a bank provider.
type BankFeature struct {
	CanPayTo         bool           `json:"canPayTo,omitempty"`
	CanPayFrom       bool           `json:"canPayFrom,omitempty"`
	CanVerifyAccount bool           `json:"canVerifyAccount,omitempty"`
	Schemes          []string       `json:"schemes,omitempty"`
	Metadata         map[string]any `json:"metadata"`
}

// BankFeatureSet groups the capabilities supported by a bank provider.
type BankFeatureSet struct {
	Payout          BankFeature `json:"payout"`
	Payment         BankFeature `json:"payment"`
	KYCVerification BankFeature `json:"kycVerification"`
}

// BankStatus reports whether a bank provider is currently active.
type BankStatus struct {
	Active bool `json:"active"`
}

// Bank is the bank provider resource returned by the API.
type Bank struct {
	ProviderID string         `json:"providerId"`
	Name       string         `json:"name"`
	Country    string         `json:"country"`
	Status     BankStatus     `json:"status"`
	FeatureSet BankFeatureSet `json:"featureSet"`
	CreateTime string         `json:"createTime"`
	UpdateTime string         `json:"updateTime"`
}

// BankList is the paginated bank list response.
type BankList struct {
	Result     []Bank     `json:"result"`
	Pagination Pagination `json:"pagination"`
}

// BankService provides access to the banks API.
type BankService struct {
	client *Client
}

// Retrieve fetches a single bank provider by its provider id.
func (s *BankService) Retrieve(ctx context.Context, providerID string) (*Bank, error) {
	if providerID == "" {
		return nil, newValidationError("providerID is required")
	}
	var out Bank
	if err := s.client.do(ctx, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("%s/%s", bankPath, providerID),
		out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns all bank providers.
func (s *BankService) List(ctx context.Context) (*BankList, error) {
	var out BankList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    bankPath,
		out:     &out,
		rawBody: true,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// MomoFeature describes a single capability of a momo provider.
type MomoFeature struct {
	CanPayTo         bool           `json:"canPayTo,omitempty"`
	CanPayFrom       bool           `json:"canPayFrom,omitempty"`
	CanVerifyAccount bool           `json:"canVerifyAccount,omitempty"`
	Schemes          []string       `json:"schemes,omitempty"`
	Metadata         map[string]any `json:"metadata"`
}

// MomoFeatureSet groups the capabilities supported by a momo provider.
type MomoFeatureSet struct {
	Payout          MomoFeature `json:"payout"`
	Payment         MomoFeature `json:"payment"`
	KYCVerification MomoFeature `json:"kycVerification"`
}

// MomoStatus reports whether a momo provider is currently active.
type MomoStatus struct {
	Active bool `json:"active"`
}

// Momo is the mobile money (momo) provider resource returned by the API.
type Momo struct {
	ProviderID string         `json:"providerId"`
	Name       string         `json:"name"`
	Country    string         `json:"country"`
	Status     MomoStatus     `json:"status"`
	FeatureSet MomoFeatureSet `json:"featureSet"`
	CreateTime string         `json:"createTime"`
	UpdateTime string         `json:"updateTime"`
}

// MomoList is the paginated momo list response.
type MomoList struct {
	Result     []Momo     `json:"result"`
	Pagination Pagination `json:"pagination"`
}

// MomoService provides access to the momos API.
type MomoService struct {
	client *Client
}

// Retrieve fetches a single momo provider by its provider id.
func (s *MomoService) Retrieve(ctx context.Context, providerID string) (*Momo, error) {
	if providerID == "" {
		return nil, newValidationError("providerID is required")
	}
	var out Momo
	if err := s.client.do(ctx, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("%s/%s", momoPath, providerID),
		out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns all momo providers.
func (s *MomoService) List(ctx context.Context) (*MomoList, error) {
	var out MomoList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    momoPath,
		out:     &out,
		rawBody: true,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}
