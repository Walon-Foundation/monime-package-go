package monime

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckoutSession_Create(t *testing.T) {
	var gotMethod, gotPath, gotIdem string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotIdem = r.Method, r.URL.Path, r.Header.Get("Idempotency-Key")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"cs_1","status":"pending","redirectUrl":"https://pay.monime.io/cs_1","lineItems":{"data":[{"name":"Order","price":{"currency":"SLE","value":1000},"quantity":2}]}}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.CheckoutSession().Create(context.Background(), CreateCheckoutSessionParams{
		Name:               "Order",
		SuccessURL:         "https://example.com/success",
		CancelURL:          "https://example.com/cancel",
		Description:        "An order",
		Reference:          "ord-1042",
		FinancialAccountID: "fac-1",
		LineItems: []CheckoutSessionLineItemParams{
			{Name: "Ticket", Price: Amount{Value: 1000}, Quantity: 2, Images: []string{"https://img/1.png"}},
			{Name: "Lunch", Price: Amount{Currency: "SLE", Value: 500}},
		},
		PaymentOptions:  &CheckoutSessionPaymentOptions{Card: &CheckoutSessionCardOption{Disable: true}, Momo: &CheckoutSessionProviderOption{EnabledProviders: []string{"m17"}}},
		BrandingOptions: &CheckoutSessionBrandingOptions{PrimaryColor: "#112233"},
		Metadata:        map[string]any{"order_id": "1042"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "cs_1" || got.Status != "pending" || got.RedirectURL != "https://pay.monime.io/cs_1" {
		t.Fatalf("unexpected response: %+v", got)
	}
	if len(got.LineItems.Data) != 1 || got.LineItems.Data[0].Price.Value != 1000 {
		t.Fatalf("unexpected line items: %+v", got.LineItems)
	}
	if gotMethod != http.MethodPost || gotPath != "/checkout-sessions" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if gotIdem == "" {
		t.Error("expected Idempotency-Key header")
	}

	// top-level body shape
	if gotBody["name"] != "Order" || gotBody["successUrl"] != "https://example.com/success" || gotBody["cancelUrl"] != "https://example.com/cancel" {
		t.Errorf("unexpected top-level body: %+v", gotBody)
	}
	if gotBody["financialAccountId"] != "fac-1" || gotBody["reference"] != "ord-1042" {
		t.Errorf("financialAccountId/reference = %v/%v", gotBody["financialAccountId"], gotBody["reference"])
	}
	if _, present := gotBody["callbackState"]; present {
		t.Errorf("callbackState should be omitted when unset, got %v", gotBody["callbackState"])
	}

	// line items: amounts unscaled, currency defaulted to SLE, type custom
	items, _ := gotBody["lineItems"].([]any)
	if len(items) != 2 {
		t.Fatalf("lineItems = %v", gotBody["lineItems"])
	}
	first := items[0].(map[string]any)
	if first["type"] != "custom" || first["quantity"].(float64) != 2 {
		t.Errorf("unexpected line item: %+v", first)
	}
	price := first["price"].(map[string]any)
	if price["currency"] != "SLE" || price["value"].(float64) != 1000 {
		t.Errorf("price = %+v, want SLE/1000 (unscaled)", price)
	}
	if _, present := items[1].(map[string]any)["quantity"]; present {
		t.Error("quantity should be omitted when zero (the API defaults it to 1)")
	}

	po, _ := gotBody["paymentOptions"].(map[string]any)
	if card, _ := po["card"].(map[string]any); card["disable"] != true {
		t.Errorf("paymentOptions = %v", po)
	}
	if _, present := po["bank"]; present {
		t.Error("an option the caller did not set must be left to the API's default")
	}
	branding := gotBody["brandingOptions"].(map[string]any)
	if branding["primaryColor"] != "#112233" {
		t.Errorf("primaryColor = %v", branding["primaryColor"])
	}
}

func TestCheckoutSession_Create_OmitsUnsetFields(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"cs_2"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.CheckoutSession().Create(context.Background(), CreateCheckoutSessionParams{
		Name: "Order", LineItems: []CheckoutSessionLineItemParams{{Name: "Ticket", Price: Amount{Value: 1000}}},
	}); err != nil {
		t.Fatal(err)
	}
	// The body rejects unknown properties; an empty financialAccountId or a
	// forced empty object is not what the caller asked for.
	for _, k := range []string{"financialAccountId", "description", "reference", "successUrl", "cancelUrl", "paymentOptions", "brandingOptions", "metadata"} {
		if _, present := gotBody[k]; present {
			t.Errorf("%s should be omitted when unset, got %v", k, gotBody[k])
		}
	}
}

func TestCheckoutSession_Create_RejectsMixedCurrencies(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when validation fails")
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.CheckoutSession().Create(context.Background(), CreateCheckoutSessionParams{
		Name: "Order", LineItems: []CheckoutSessionLineItemParams{
			{Name: "A", Price: Amount{Currency: "SLE", Value: 1}},
			{Name: "B", Price: Amount{Currency: "USD", Value: 1}},
		},
	}); err == nil {
		t.Fatal("expected validation error for line items in two currencies")
	}
}

func TestCheckoutSession_Create_ValidationShortCircuits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when validation fails")
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.CheckoutSession().Create(context.Background(), CreateCheckoutSessionParams{Name: ""})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCheckoutSession_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/checkout-sessions" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"result":[{"id":"cs_1"},{"id":"cs_2"}],"pagination":{"count":2,"next":"abc"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.CheckoutSession().List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Result) != 2 || got.Pagination.Count != 2 || got.Pagination.Next != "abc" {
		t.Fatalf("unexpected list: %+v", got)
	}
}

func TestCheckoutSession_Retrieve(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/checkout-sessions/cs_9" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"result":{"id":"cs_9","status":"completed"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.CheckoutSession().Retrieve(context.Background(), "cs_9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "cs_9" || got.Status != "completed" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestCheckoutSession_Retrieve_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if _, err := c.CheckoutSession().Retrieve(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestCheckoutSession_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/checkout-sessions/cs_1" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.CheckoutSession().Delete(context.Background(), "cs_1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckoutSession_Delete_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if err := c.CheckoutSession().Delete(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}
