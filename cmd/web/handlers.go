package main

import (
	"net/http"
	"strconv"

	"github.com/ahojo/go-stripe-ecommerce/internal/cards"
	"github.com/go-chi/chi/v5"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "home", &templateData{}); err != nil {
		app.errorLog.Println(err)
		return
	}
}

func (app *application) VirtualTerminal(w http.ResponseWriter, r *http.Request) {

	if err := app.renderTemplate(w, r, "terminal", &templateData{}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
		return
	}
}

func (app *application) PaymentSucceeded(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// read the post data from the form
	cardHolder := r.Form.Get("cardholder-name")
	email := r.Form.Get("cardholder-email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")

	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	// Get the payment intent
	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear


	// Create a new customer
	// Create a new order
	// create a new transaction


	data := make(map[string]interface{})
	data["cardHolder"] = cardHolder
	data["email"] = email
	data["paymentIntent"] = paymentIntent
	data["paymentMethod"] = paymentMethod
	data["paymentAmount"] = paymentAmount
	data["paymentCurrency"] = paymentCurrency
	data["last_four"] = lastFour
	data["expiry_month"] = expiryMonth
	data["expiry_year"] = expiryYear
	data["bank_return_code"] = pi.Charges.Data[0].ID

	// should write this data to session then redirect to page.

	
	if err = app.renderTemplate(w, r, "succeeded", &templateData{Data: data}); err != nil {
		app.errorLog.Println(err)
		return
	}
}

//ChargeOnce Displays the page to buy one widget
func (app *application) ChargeOnce(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	widgetID, _ := strconv.Atoi(id)

	widget, err := app.DB.GetWidget(widgetID)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	//widget := models.Widget{
	//	ID:             1,
	//	Name:           "Custome Widget",
	//	Description:    "A very nice widget",
	//	InventoryLevel: 10,
	//	Price:          1000,
	//}

	data := make(map[string]interface{})
	data["widget"] = widget
	if err := app.renderTemplate(w, r, "buy-once", &templateData{
		Data: data,
	}, "stripe-js"); err != nil {
		app.errorLog.Println(err)
		return
	}
}
