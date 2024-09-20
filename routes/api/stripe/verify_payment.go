package stripe

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/deta/deta-go/deta"
	"github.com/deta/deta-go/service/base"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
)

type VerificationCode struct {
	Code string `json:"key"`
	SessionId string `json:"session_id"`
	AccountId string `json:"account_id"`
	Used bool `json:"used"`
}

type Payment struct {
	Created int64 `json:"created"`	
	Customer string `json:"customer"`
	Email string `json:"email"`
	Status string `json:"status"`
	Amount int64 `json:"amount"`
	Product string `json:"product"`
	Subscription bool `json:"subscription"`
}

func VerifyPayment(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	stripe.Key = os.Getenv("STRIPE_TOKEN")

	code := r.URL.Query().Get("code")

	d, err := deta.New(deta.WithProjectKey(os.Getenv("PROJECT_KEY")))

	if err != nil {
		return err
	}

	codeDb, err := base.New(d, "codes")

	if err != nil {
		return err
	}

	verificationCode := VerificationCode{}	

	err = codeDb.Get(code, &verificationCode)

	if err != nil {
		if errors.Is(err, deta.ErrNotFound) {
			w.WriteHeader(403)
			w.Write([]byte("The code is invalid!"))

			return nil
		}

		return err
	}

	if verificationCode.Used {
		w.WriteHeader(403)
		w.Write([]byte("The code has already been used!"))

		return nil
	}

	params := stripe.CheckoutSessionParams{}

	params.SetStripeAccount(verificationCode.AccountId)
	params.Expand = []*string{stripe.String("customer"), stripe.String("line_items"), stripe.String("line_items.data.price.product")}

	session, err := session.Get(verificationCode.SessionId, &params) 

	if err != nil {
		return err
	}

	subscription := false

	if session.Subscription != nil {
		subscription = true
	}

	fmt.Println(json.MarshalIndent(session.LineItems.Data[0].Price, "", "  "))

	payment := Payment{
		Created: session.Created,
		Customer: session.Customer.Name,
		Email: session.Customer.Email,
		Status: string(session.PaymentStatus),
		Amount: session.LineItems.Data[0].Price.UnitAmount,
		Product: session.LineItems.Data[0].Price.Product.Name,
		Subscription: subscription,
	}

	stringified, err := json.MarshalIndent(payment, "", "  ")

	if err != nil {
		return err
	}

	err = codeDb.Update(verificationCode.Code, base.Updates{
		"used": true,	
	})

	w.WriteHeader(200)
	w.Write(stringified)

	return nil
}
