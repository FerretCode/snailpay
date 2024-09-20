package dashboard

import (
	"net/http"
	"os"
	"text/template"

	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/balance"
	"github.com/stripe/stripe-go/v74/payout"
)

type payoutList struct {
	PayoutList []*stripe.Payout 
	Withdrawn float64 
	Balance float64
	Pending float64
}

func Payout(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	stripe.Key = os.Getenv("STRIPE_TOKEN")
	
	user := r.Context().Value("user").(auth.SnailUser)

	tmpl, err := template.ParseFiles("templates/dashboard/stripe/payouts/payouts.html")

	if err != nil {
		return err
	}

	withdrawn := float64(0)

	payoutList := payoutList{}

	payoutListParams := &stripe.PayoutListParams{}

	payoutListParams.SetStripeAccount(user.AccountId)

	pList := payout.List(payoutListParams)

	for pList.Next() {
		payout := pList.Payout()

		payoutList.PayoutList = append(payoutList.PayoutList, payout)

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

	payoutList.Balance = bal / 100
	payoutList.Pending = pending / 100
	payoutList.Withdrawn = withdrawn / 100

	tmpl.Execute(w, payoutList)

	return nil
}
