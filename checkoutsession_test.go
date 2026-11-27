package monime

import (
	"errors"
	"testing"
)

func TestCreateCheckoutSessionParams_Validation(t *testing.T) {
	item := CheckoutSessionLineItemParams{Name: "Ticket", Price: Amount{Value: 1000}}
	tests := []struct {
		name    string
		params  CreateCheckoutSessionParams
		wantErr bool
	}{
		{name: "name and one line item", params: CreateCheckoutSessionParams{Name: "Order", LineItems: []CheckoutSessionLineItemParams{item}}, wantErr: false},
		{name: "urls are optional but must be urls", params: CreateCheckoutSessionParams{Name: "Order", LineItems: []CheckoutSessionLineItemParams{item}, SuccessURL: "not-a-url"}, wantErr: true},
		{name: "missing name", params: CreateCheckoutSessionParams{LineItems: []CheckoutSessionLineItemParams{item}}, wantErr: true},
		{name: "no line items", params: CreateCheckoutSessionParams{Name: "Order"}, wantErr: true},
		{name: "seventeen line items", params: CreateCheckoutSessionParams{Name: "Order", LineItems: make([]CheckoutSessionLineItemParams, 17)}, wantErr: true},
		{name: "line item without a name", params: CreateCheckoutSessionParams{Name: "Order", LineItems: []CheckoutSessionLineItemParams{{Price: Amount{Value: 1}}}}, wantErr: true},
		{name: "quantity over 100000", params: CreateCheckoutSessionParams{Name: "Order", LineItems: []CheckoutSessionLineItemParams{{Name: "T", Price: Amount{Value: 1}, Quantity: 100001}}}, wantErr: true},
		{name: "four images", params: CreateCheckoutSessionParams{Name: "Order", LineItems: []CheckoutSessionLineItemParams{{Name: "T", Price: Amount{Value: 1}, Images: []string{"https://a/1", "https://a/2", "https://a/3", "https://a/4"}}}}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStruct(tt.params)
			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
			if tt.wantErr {
				var valErr *ValidationError
				if !errors.As(err, &valErr) {
					t.Fatalf("expected *ValidationError, got %T", err)
				}
			}
		})
	}
}
