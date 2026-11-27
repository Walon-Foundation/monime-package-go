package monime

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPaymentCode_Create(t *testing.T) {
	var gotMethod, gotPath, gotIdem string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotIdem = r.Method, r.URL.Path, r.Header.Get("Idempotency-Key")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"pc_1","ussdCode":"*123#","amount":{"currency":"SLE","value":10000}}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PaymentCode().Create(context.Background(), CreatePaymentCodeParams{
		Name:                  "Test Payment",
		Amount:                &Amount{Value: 10000},
		FinancialAccountID:    "fac-1",
		CustomerName:          "John Doe",
		AuthorizedPhoneNumber: "+23276000000",
		Reference:             "ord-1",
		Metadata:              map[string]any{"order_id": "ord-1"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "pc_1" || got.USSDCode != "*123#" {
		t.Fatalf("unexpected response: %+v", got)
	}
	if gotMethod != http.MethodPost || gotPath != "/payment-codes" {
		t.Fatalf("method/path = %s %s", gotMethod, gotPath)
	}
	if gotIdem == "" {
		t.Error("expected Idempotency-Key header")
	}
	// The amount is already in minor units and is sent as-is; the currency
	// defaults to SLE.
	amount, _ := gotBody["amount"].(map[string]any)
	if amount["value"].(float64) != 10000 || amount["currency"] != "SLE" {
		t.Errorf("amount = %v, want SLE 10000", amount)
	}
	if gotBody["financialAccountId"] != "fac-1" || gotBody["reference"] != "ord-1" || gotBody["authorizedPhoneNumber"] != "+23276000000" {
		t.Errorf("body = %v", gotBody)
	}
	if customer, _ := gotBody["customer"].(map[string]any); customer["name"] != "John Doe" {
		t.Errorf("customer = %v", gotBody["customer"])
	}
	// Nothing the caller left out is forced: the API defaults mode to one_time,
	// enable to true and duration to 10m.
	for _, k := range []string{"mode", "enable", "duration", "recurrentPaymentTarget", "authorizedProviders"} {
		if _, present := gotBody[k]; present {
			t.Errorf("%s should be omitted when unset, got %v", k, gotBody[k])
		}
	}
}

func TestPaymentCode_Create_Recurrent(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"pc_3","mode":"recurrent"}}`))
	}))
	defer srv.Close()

	enable := false
	c := newTestClient(t, srv)
	if _, err := c.PaymentCode().Create(context.Background(), CreatePaymentCodeParams{
		Name: "Dues", Mode: PaymentCodeModeRecurrent, Enable: &enable, Duration: "1h30m",
		AuthorizedProviders:    []string{"m18"},
		RecurrentPaymentTarget: &PaymentCodeRecurrentTarget{ExpectedPaymentCount: 3, ExpectedPaymentTotal: &Amount{Currency: "USD", Value: 900}},
	}); err != nil {
		t.Fatal(err)
	}
	target, _ := gotBody["recurrentPaymentTarget"].(map[string]any)
	total, _ := target["expectedPaymentTotal"].(map[string]any)
	if gotBody["mode"] != "recurrent" || gotBody["enable"] != false || gotBody["duration"] != "1h30m" ||
		target["expectedPaymentCount"] != float64(3) || total["currency"] != "USD" || total["value"] != float64(900) {
		t.Errorf("body = %v", gotBody)
	}
}

func TestPaymentCode_Create_TargetNeedsRecurrentMode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when validation fails")
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.PaymentCode().Create(context.Background(), CreatePaymentCodeParams{
		Name: "Dues", RecurrentPaymentTarget: &PaymentCodeRecurrentTarget{ExpectedPaymentCount: 2},
	}); err == nil {
		t.Fatal("expected validation error for a target on a one_time code")
	}
	if _, err := c.PaymentCode().Create(context.Background(), CreatePaymentCodeParams{
		Name: "Dues", Amount: &Amount{Value: 0},
	}); err == nil {
		t.Fatal("expected validation error for a zero amount")
	}
}

func TestPaymentCode_Create_OmitsEmptyAccount(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"pc_2"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if _, err := c.PaymentCode().Create(context.Background(), CreatePaymentCodeParams{
		Name: "Pay", Amount: &Amount{Value: 500},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := gotBody["financialAccountId"]; ok {
		t.Error("financialAccountId should be omitted when empty")
	}
}

func TestPaymentCode_Create_ValidationShortCircuits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("server should not be called when validation fails")
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.PaymentCode().Create(context.Background(), CreatePaymentCodeParams{Name: ""})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestPaymentCode_Retrieve(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/payment-codes/pc_9" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"result":{"id":"pc_9","status":"active"}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PaymentCode().Retrieve(context.Background(), "pc_9")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "pc_9" || got.Status != "active" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestPaymentCode_Retrieve_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if _, err := c.PaymentCode().Retrieve(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestPaymentCode_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"result":[{"id":"pc_1"},{"id":"pc_2"}],"pagination":{"count":2,"next":""}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PaymentCode().List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Result) != 2 || got.Pagination.Count != 2 {
		t.Fatalf("unexpected list: %+v", got)
	}
}

func TestPaymentCode_Update(t *testing.T) {
	var gotMethod string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"result":{"id":"pc_1","enable":false}}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	got, err := c.PaymentCode().Update(context.Background(), "pc_1", map[string]any{"enable": false})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPatch {
		t.Errorf("method = %s, want PATCH", gotMethod)
	}
	if got.Enable {
		t.Error("expected enable=false in response")
	}
	if gotBody["enable"] != false {
		t.Errorf("body enable = %v", gotBody["enable"])
	}
}

func TestPaymentCode_Update_Validation(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if _, err := c.PaymentCode().Update(context.Background(), "", map[string]any{"enable": true}); err == nil {
		t.Fatal("expected error for empty id")
	}
	if _, err := c.PaymentCode().Update(context.Background(), "pc_1", nil); err == nil {
		t.Fatal("expected error for empty params")
	}
	if _, err := c.PaymentCode().Update(context.Background(), "pc_1", map[string]any{"mode": "recurrent"}); err == nil {
		t.Fatal("expected error for a field PATCH /payment-codes/{id} does not accept")
	}
}

func TestPaymentCode_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/payment-codes/pc_1" {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.PaymentCode().Delete(context.Background(), "pc_1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPaymentCode_Delete_RequiresID(t *testing.T) {
	c := newTestClient(t, httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	if err := c.PaymentCode().Delete(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty id")
	}
}
