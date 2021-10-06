package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ahojo/go-stripe-ecommerce/internal/cards"
	"github.com/ahojo/go-stripe-ecommerce/internal/models"
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


type TransactionData struct {
	FirstName string
	LastName  string
	Email     string
	PaymentIntentID string
	PaymentMethodID string
	PaymentAmount   int
	PaymentCurrency string
	LastFour       string
	ExpiryMonth    int
	ExpiryYear     int
	BankReturnCode string
}
// GetTransactionData Gets transaction data from post and stripe
func (app *application) GetTransactionData(r *http.Request) (TransactionData, error) {
	var txnData TransactionData

	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	// read the post data from the form
	firstName := r.Form.Get("first-name")
	lastName := r.Form.Get("last-name")
	email := r.Form.Get("cardholder-email")
	paymentIntent := r.Form.Get("payment_intent")
	paymentMethod := r.Form.Get("payment_method")
	paymentAmount := r.Form.Get("payment_amount")
	paymentCurrency := r.Form.Get("payment_currency")

	amount, _ := strconv.Atoi(paymentAmount)


	card := cards.Card{
		Secret: app.config.stripe.secret,
		Key:    app.config.stripe.key,
	}

	// Get the payment intent
	pi, err := card.RetrievePaymentIntent(paymentIntent)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	pm, err := card.GetPaymentMethod(paymentMethod)
	if err != nil {
		app.errorLog.Println(err)
		return txnData, err
	}

	lastFour := pm.Card.Last4
	expiryMonth := pm.Card.ExpMonth
	expiryYear := pm.Card.ExpYear

	txnData = TransactionData{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		PaymentIntentID: paymentIntent,
		PaymentMethodID: paymentMethod,
		PaymentAmount:   amount,
		PaymentCurrency: paymentCurrency,
		LastFour:       lastFour,
		ExpiryMonth:    int(expiryMonth),
		ExpiryYear:     int(expiryYear),
		BankReturnCode: pi.Charges.Data[0].ID,
	}
	return txnData, nil

}

func (app *application) VirtualTerminalPaymentSucceeded(w http.ResponseWriter, r *http.Request) {

	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	// create a new transaction

	
	transac := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:        txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		BankReturnCode:      txnData.BankReturnCode,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		TransactionStatusID: 2,
	}
	_, err = app.SaveTransaction(transac)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// should write this data to session then redirect to page.
	app.Session.Put(r.Context(), "receipt", txnData)
	http.Redirect(w,r, "/virtual-terminal-receipt", http.StatusSeeOther)
}

func (app *application) VirtualTerminalReceipt(w http.ResponseWriter, r *http.Request) {

	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn

	app.Session.Remove(r.Context(), "receipt")

	if err := app.renderTemplate(w, r, "virtual-terminal-receipt", &templateData{Data: data}); err != nil {
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

	
	widgetID, _ := strconv.Atoi(r.Form.Get("product_id"))

	txnData, err := app.GetTransactionData(r)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	// Create a new customer
	customerID, err := app.SaveCustomer(txnData.FirstName, txnData.LastName, txnData.Email)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	// app.infoLog.Print(customerID)
	// create a new transaction

	
	transac := models.Transaction{
		Amount:              txnData.PaymentAmount,
		Currency:            txnData.PaymentCurrency,
		LastFour:            txnData.LastFour,
		ExpiryMonth:        txnData.ExpiryMonth,
		ExpiryYear:          txnData.ExpiryYear,
		BankReturnCode:      txnData.BankReturnCode,
		PaymentIntent:       txnData.PaymentIntentID,
		PaymentMethod:       txnData.PaymentMethodID,
		TransactionStatusID: 2,
	}
	transacID, err := app.SaveTransaction(transac)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
	app.infoLog.Print(transacID)

	// Create a new order
	order := models.Order{
		CustomerID:    customerID,
		TransactionID: transacID,
		WidgetID:      widgetID,
		StatusID:      1,
		Quantity:      1,
		Amount:        txnData.PaymentAmount,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = app.SaveOrder(order)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	
	

	// should write this data to session then redirect to page.
	app.Session.Put(r.Context(), "receipt", txnData)
	http.Redirect(w,r, "/receipt", http.StatusSeeOther)
	
	
}

func (app *application) Receipt(w http.ResponseWriter, r *http.Request) {

	txn := app.Session.Get(r.Context(), "receipt").(TransactionData)
	data := make(map[string]interface{})
	data["txn"] = txn

	app.Session.Remove(r.Context(), "receipt")

	if err := app.renderTemplate(w, r, "receipt", &templateData{Data: data}); err != nil {
		app.errorLog.Println(err)
		return
	}

}
// SaveCustomer saves a customer and returns it's id.
func (app *application) SaveCustomer(firstName, lastName, email string) (int, error) {
	customer := models.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	id, err := app.DB.InsertCustomer(customer)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// SaveTransaction saves a customer and returns it's id.
func (app *application) SaveTransaction(txn models.Transaction) (int, error) {

	id, err := app.DB.InsertTransaction(txn)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// SaveOrder saves a customer and returns it's id.
func (app *application) SaveOrder(order models.Order) (int, error) {

	id, err := app.DB.InsertOrder(order)
	if err != nil {
		return 0, err
	}

	return id, nil
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

func (app *application) BronzePlan(w http.ResponseWriter, r *http.Request) {
	widget,err := app.DB.GetWidget(2)
	if err != nil {
		app.errorLog.Println(err)
		return
	}

	data := make(map[string]interface{})
	data["widget"] = widget

	if err := app.renderTemplate(w,r, "bronze-plan", &templateData{Data: data}); err != nil {
		app.errorLog.Print(err)
	}
}
func (app *application) BronzeReceipt(w http.ResponseWriter, r *http.Request) {
	


	if err := app.renderTemplate(w,r, "receipt-plan", &templateData{}); err != nil {
		app.errorLog.Print(err)
	}
}


// LoginPage displays the login page
func (app *application) LoginPage(w http.ResponseWriter, r *http.Request) {
	if err := app.renderTemplate(w, r, "login", &templateData{}); err != nil {
		app.errorLog.Println(err)
		return
	}
}