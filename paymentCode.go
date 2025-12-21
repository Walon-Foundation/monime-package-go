package monime

import (
	"errors"
	"strings"
)

const URL = "https://api.monime.io/v1/payment-codes"


type PaymentCode struct {
	client *Client
}

type PaymentCodeReturn struct {
	Success bool `json:"success"`
	Error *error `json:"error"`
	Data *any  `json:"data"`
}

func (c *PaymentCode)CreateCode(
	paymentName,name,phoneNumber,financialAccountId string,
	amount int,
) *PaymentCodeReturn{
	if strings.TrimSpace(financialAccountId) == "" {
		financialAccountId = "null"
	}

	if strings.TrimSpace(paymentName) == "" || strings.TrimSpace(phoneNumber) == "" || strings.TrimSpace(name) == "" {
		return &PaymentCodeReturn{Success: false, Error: errors.New("paymentName, name or phoneNumber is missing ")}
	}
}