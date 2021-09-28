package cards

import (
	"github.com/stripe/stripe-go/v72/sub"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentintent"
	"github.com/stripe/stripe-go/v72/paymentmethod"
)

type Card struct {
	Secret   string
	Key      string
	Currency string
}

type Transaction struct {
	Amount              int
	TransactionStatusID int
	Currency            string
	LastFour            string
	BankReturnCode      string
}

func (c *Card) Charge(currency string, amount int) (*stripe.PaymentIntent, string, error) {
	return c.CreatePaymentIntent(currency, amount)
}

func (c *Card) CreatePaymentIntent(currency string, Amount int) (*stripe.PaymentIntent, string, error) {
	stripe.Key = c.Secret

	// Create a PaymentIntent with the order amount and currency
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(Amount)),
		Currency: stripe.String(currency),
	}

	// can add more infor to parames like below
	// params.AddMetadata("key", "value")

	pi, err := paymentintent.New(params) // payment intent comes  from stripe
	if err != nil {
		msg := ""

		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = cardErrorMessage(stripeErr.Code)
		}
		return nil, msg, err
	}

	return pi, "", nil
}

// GetPaymentMethod get's the payment method from the paymentintent id
func (c *Card) GetPaymentMethod(s string) (*stripe.PaymentMethod, error) {
	stripe.Key = c.Secret

	pm, err := paymentmethod.Get(s, nil)

	if err != nil {
		return nil, err
	}

	return pm, nil
}

// RetrievePaymentIntent gets an existing payment intentby id
func (c *Card) RetrievePaymentIntent(id string) (*stripe.PaymentIntent, error) {
	stripe.Key = c.Secret

	pi, err := paymentintent.Get(id, nil)

	if err != nil {
		return nil, err
	}
	return pi, nil
}

func (c *Card) SubscribeToPlan(customer *stripe.Customer, plan, email, last4, cardType string) (*stripe.Subscription, error) {

	stripeCustomerID := customer.ID
	items := []*stripe.SubscriptionItemsParams{
		{Plan: stripe.String(plan)},
	}
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(stripeCustomerID),
		Items:    items,
	}

	params.AddMetadata("last_four", last4)
	params.AddMetadata("card_type", cardType)
	params.AddExpand("latest_invoice.payment_intent")
	subscription, err := sub.New(params)
	if err != nil {
		return nil, err
	}

	return subscription, nil 

}
func (c *Card) CreateCustomer(pm, email string) (*stripe.Customer, string, error) {
	stripe.Key = c.Secret

	params := &stripe.CustomerParams{
		PaymentMethod: stripe.String(pm),
		Email:         stripe.String(email),
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(pm),
		},
	}

	customer, err := customer.New(params)
	if err != nil {
		msg := ""
		if stripErr, ok := err.(*stripe.Error); ok {
			msg = cardErrorMessage(stripErr.Code)
		}
		return nil, msg, err
	}

	return customer, "", nil

}
func cardErrorMessage(code stripe.ErrorCode) string {
	var msg = ""
	switch code {
	case stripe.ErrorCodeCardDeclined:
		msg = "Your card was declined"
	case stripe.ErrorCodeExpiredCard:
		msg = "Your card is expired"
	case stripe.ErrorCodeIncorrectCVC:
		msg = "Incorrect CVC code"
	case stripe.ErrorCodeIncorrectZip:
		msg = "Incorrect zip/postal code"
	case stripe.ErrorCodeAmountTooLarge:
		msg = "The amount is too large to charge to your card"
	case stripe.ErrorCodeAmountTooSmall:
		msg = "The amount is too small to charge to your card"
	case stripe.ErrorCodeBalanceInsufficient:
		msg = "Insufficient balance"
	case stripe.ErrorCodePostalCodeInvalid:
		msg = "Your postal code is invalid"
	default:
		msg = "Your card was declined"
	}
	return msg
}
