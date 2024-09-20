package stripe

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/payout"
)

type PayoutRequest struct {
	Amount int `json:"amount"`
}

func NewPayout(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	user := r.Context().Value("user").(auth.SnailUser)

	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	payoutRequest := PayoutRequest{}

	if err := json.Unmarshal(bytes, &payoutRequest); err != nil {
		return err
	}

	payoutParams := &stripe.PayoutParams{
		Amount: stripe.Int64(int64(payoutRequest.Amount)),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
	}

	payoutParams.SetStripeAccount(user.AccountId)

	_, err = payout.New(payoutParams)

	if err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write([]byte("Your payout has been initiated!"))

	return nil
}
