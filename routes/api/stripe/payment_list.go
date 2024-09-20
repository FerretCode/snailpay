package stripe

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
)

type APIPayment struct {
	Amount int64 `json:"amount"`
	Customer string `json:"customer"`
	Email string `json:"email"`
	ID string `json:"id"`
	Timestamp int64 `json:"timestamp"`
	Status string `json:"status"`
}

func PaymentList(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	stripe.Key = os.Getenv("STRIPE_TOKEN")

	user := r.Context().Value("user").(auth.SnailUser)

	var apiPaymentList []APIPayment

	paymentIntentListParams := stripe.PaymentIntentListParams{}

	paymentIntentListParams.SetStripeAccount(user.AccountId)
	paymentIntentListParams.Expand = []*string{stripe.String("data.customer")}

	paymentIntents := paymentintent.List(&paymentIntentListParams)

	for paymentIntents.Next() {
		pay := paymentIntents.PaymentIntent()
		
		var customerName string
		var customerEmail string

		if pay.Customer == nil {
			customerName = "N/A"	
			customerEmail = "N/A"
		} else {
			customerName = pay.Customer.Name
			customerEmail = pay.Customer.Email
		}

		apiPayment := APIPayment{
			Amount: pay.Amount,	
			Customer: customerName,
			Email: customerEmail,
			ID: pay.ID,
			Timestamp: pay.Created,
			Status: string(pay.Status),
		}

		apiPaymentList = append(apiPaymentList, apiPayment)
	}

	if len(apiPaymentList) == 0 {
		w.WriteHeader(200)
		w.Write([]byte("There are currently no payments."))

		return nil
	}

	stringified, err := json.MarshalIndent(apiPaymentList, "", "  ")

	if err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write(stringified)

	return nil
}
