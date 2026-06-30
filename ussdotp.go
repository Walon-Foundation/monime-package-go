package monime

import (
	"context"
	"fmt"
	"net/http"
)

// UssdOtpService provides access to the ussd-otps API.
type UssdOtpService struct {
	client *Client
}

// UssdOtp returns the ussd-otps service.
func (c *Client) UssdOtp() *UssdOtpService {
	return &UssdOtpService{client: c}
}

const ussdOtpPath = "/ussd-otps"

// CreateUssdOtpParams holds the inputs for creating a USSD OTP. It mirrors the
// TypeScript SDK's create request, whose zod schema requires only phoneNumber.
type CreateUssdOtpParams struct {
	PhoneNumber string `validate:"required"`
}

// UssdOtp is the one-time passcode session resource returned by the API.
type UssdOtp struct {
	ID                    string         `json:"id"`
	Status                string         `json:"status"`
	DialCode              string         `json:"dialCode"`
	AuthorizedPhoneNumber string         `json:"authorizedPhoneNumber"`
	VerificationMessage   string         `json:"verificationMessage"`
	CreateTime            string         `json:"createTime"`
	ExpireTime            string         `json:"expireTime"`
	Metadata              map[string]any `json:"metadata,omitempty"`
}

// UssdOtpList is the paginated ussd-otp list response.
type UssdOtpList struct {
	Result     []UssdOtp  `json:"result"`
	Pagination Pagination `json:"pagination"`
}

// Create creates a new USSD OTP session. This mirrors the TypeScript SDK, which
// posts the validated options (phoneNumber) with a generated idempotency key.
func (s *UssdOtpService) Create(ctx context.Context, params CreateUssdOtpParams) (*UssdOtp, error) {
	if err := validateStruct(params); err != nil {
		return nil, err
	}

	body := map[string]any{
		"phoneNumber": params.PhoneNumber,
	}

	key, err := generateIdempotencyKey()
	if err != nil {
		return nil, &Error{Message: err.Error()}
	}

	var out UssdOtp
	if err := s.client.do(ctx, requestOptions{
		method:         http.MethodPost,
		path:           ussdOtpPath,
		body:           body,
		idempotencyKey: key,
		out:            &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Retrieve fetches a single USSD OTP session by id. This endpoint is net-new
// relative to the TypeScript SDK (added from the Monime API docs).
func (s *UssdOtpService) Retrieve(ctx context.Context, ussdOtpID string) (*UssdOtp, error) {
	if ussdOtpID == "" {
		return nil, newValidationError("ussdOtpID is required")
	}
	var out UssdOtp
	if err := s.client.do(ctx, requestOptions{
		method: http.MethodGet,
		path:   fmt.Sprintf("%s/%s", ussdOtpPath, ussdOtpID),
		out:    &out,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns all USSD OTP sessions. This endpoint is net-new relative to the
// TypeScript SDK (added from the Monime API docs).
func (s *UssdOtpService) List(ctx context.Context) (*UssdOtpList, error) {
	var out UssdOtpList
	if err := s.client.do(ctx, requestOptions{
		method:  http.MethodGet,
		path:    ussdOtpPath,
		out:     &out,
		rawBody: true,
	}); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a USSD OTP session by id. This endpoint is net-new relative to
// the TypeScript SDK (added from the Monime API docs).
func (s *UssdOtpService) Delete(ctx context.Context, ussdOtpID string) error {
	if ussdOtpID == "" {
		return newValidationError("ussdOtpID is required")
	}
	return s.client.do(ctx, requestOptions{
		method: http.MethodDelete,
		path:   fmt.Sprintf("%s/%s", ussdOtpPath, ussdOtpID),
	})
}
