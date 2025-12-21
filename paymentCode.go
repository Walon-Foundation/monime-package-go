package monime

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const URL = "https://api.monime.io/v1/payment-codes"

type PaymentCode struct {
	client *Client
}

type PaymentCodeData struct {
	
}

type PaymentCodeReturn struct {
	Success bool  `json:"success"`
	Error   error `json:"error,omitempty"`
	Data    PaymentCodeData   `json:"data,omitempty"`
}



func (c *PaymentCode) CreateCode(
	paymentName, name, phoneNumber, financialAccountId string,
	amount int,
) *PaymentCodeReturn {
	if strings.TrimSpace(financialAccountId) == "" {
		financialAccountId = "null"
	}

	if strings.TrimSpace(paymentName) == "" || strings.TrimSpace(phoneNumber) == "" || strings.TrimSpace(name) == "" {
		return &PaymentCodeReturn{Success: false, Error: errors.New("paymentName, name or phoneNumber is missing ")}
	}

	payload := fmt.Sprintf()
	req,_ := http.NewRequest("POST",URL, payload)

	value, err := GenerateRandomString(16)
	if err != nil {
		return &PaymentCodeReturn{Success: false, Error: err}
	}

	req.Header.Add("Idempotency-Key", value)
	req.Header.Add("Monime-Space-Id", c.client.spaceID)
	req.Header.Add("Authorization", c.client.accessToken)
	req.Header.Add("Monime-Version", string(*c.client.version))
	req.Header.Add("Content-Type", "application/json")

	res,_ := http.DefaultClient.Do(req)

	defer res.Body.Close()

	var data PaymentCodeData

	if err := json.NewDecoder(res.Body).Decode(&data);err != nil {
		return &PaymentCodeReturn{Success: false, Error: err}
	}


	return &PaymentCodeReturn{Success: true, Error: nil, Data: data}
}
