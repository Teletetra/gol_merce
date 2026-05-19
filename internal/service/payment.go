package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ecommerce_go/internal/domain"
)

type ChargeInput struct {
	OrderID       string `json:"order_id"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	Method        string `json:"method"`
	CardToken     string `json:"card_token"`
	WalletID      string `json:"wallet_id"`
	CustomerEmail string `json:"customer_email"`
}

type Gateway interface {
	Charge(context.Context, ChargeInput) (domain.Payment, error)
}

type MockGateway struct {
	Provider string
}

func (g MockGateway) Charge(_ context.Context, input ChargeInput) (domain.Payment, error) {
	now := time.Now().UTC()
	payment := domain.Payment{
		ID:            newID("pay"),
		OrderID:       input.OrderID,
		Provider:      g.Provider,
		Method:        strings.ToLower(strings.TrimSpace(input.Method)),
		Amount:        input.Amount,
		Currency:      defaultCurrency(input.Currency),
		Status:        domain.PaymentStatusCaptured,
		Reference:     fmt.Sprintf("%s-%d", g.Provider, time.Now().UnixNano()),
		ProviderTxnID: newID("txn"),
		Metadata: map[string]string{
			"customer_email": input.CustomerEmail,
		},
		CreatedAt:   now,
		ProcessedAt: &now,
	}

	if payment.Method == "" {
		payment.Method = "card"
	}

	if strings.HasPrefix(input.CardToken, "fail_") || input.Amount > 500000 {
		payment.Status = domain.PaymentStatusFailed
		payment.FailureMsg = "payment authorization declined"
		return payment, ErrPaymentFailed
	}

	return payment, nil
}
