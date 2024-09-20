package stripe

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/balance"
	"github.com/stripe/stripe-go/v74/payout"
)

type PayoutResponse struct {
	PayoutList []APIPayout
	Withdrawn float64 
	Balance float64
	Pending float64
}

type APIPayout struct {
	Date int64 `json:"date"`
	Amount int64 `json:"amount"`
	ArrivalDate int64 `json:"arrival_date"`
	Status string `json:"status"`
}

//	@title			Payout List 
//	@version		1.0
//	@description	List payouts 

//	@BasePath	/api

//	@accept		json
//	@produce	json

//	@schemes	https

//	@securityDefinitions.apiKey	ApiKeyAuth

func Payout(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	stripe.Key = os.Getenv("STRIPE_TOKEN")
	
	user := r.Context().Value("user").(auth.SnailUser)

	withdrawn := float64(0)

	payoutResponse := PayoutResponse{}

	payoutListParams := &stripe.PayoutListParams{}

	payoutListParams.SetStripeAccount(user.AccountId)

	pList := payout.List(payoutListParams)

	for pList.Next() {
		payout := pList.Payout()

		apiPayout := APIPayout{
			Date: payout.Created,
			Amount: payout.Amount,
			ArrivalDate: payout.ArrivalDate,
			Status: string(payout.Status),
		}

		payoutResponse.PayoutList = append(payoutResponse.PayoutList, apiPayout)

		withdrawn += float64(payout.Amount)
	}

	balanceParams := &stripe.BalanceParams{}

	balanceParams.SetStripeAccount(user.AccountId)

	balance, err := balance.Get(balanceParams)

	if err != nil {
		return err
	}

	bal := float64(0)
	pending := float64(0)

	for _, b := range balance.Available {
		bal += float64(b.Amount)
	}

	for _, b := range balance.Pending {
		pending += float64(b.Amount) 
	}

	payoutResponse.Balance = bal / 100
	payoutResponse.Pending = pending / 100
	payoutResponse.Withdrawn = withdrawn / 100

	stringified, err := json.MarshalIndent(payoutResponse, "", "  ")

	if err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write(stringified)

	return nil
}
