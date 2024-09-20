package stripe

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/subscription"
)

type APISubscription struct {
	Amount int64 `json:"amount"`
	Customer string `json:"customer"`
	Email string `json:"email"`
	ID string `json:"id"`
	Timestamp int64 `json:"timestamp"`
	Status string `json:"status"`
}

func SubscriptionList(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	stripe.Key = os.Getenv("STRIPE_TOKEN")

	user := r.Context().Value("user").(auth.SnailUser)

	var apiSubscriptionList []APISubscription

	subscriptionListParams := stripe.SubscriptionListParams{}

	subscriptionListParams.SetStripeAccount(user.AccountId)
	subscriptionListParams.Expand = []*string{stripe.String("data.customer")}

	subscriptions := subscription.List(&subscriptionListParams)

	for subscriptions.Next() {
		sub := *subscriptions.Subscription()

		apiSubscription := APISubscription{
			Amount: sub.Items.Data[0].Price.UnitAmount,	
			Customer: sub.Customer.Name,
			Email: sub.Customer.Email,
			ID: sub.ID,
			Timestamp: sub.Created,
			Status: string(sub.Status),
		}

		apiSubscriptionList = append(apiSubscriptionList, apiSubscription)
	}

	if len(apiSubscriptionList) == 0 {
		w.WriteHeader(200)
		w.Write([]byte("There are currently no subscriptions."))

		return nil
	}

	stringified, err := json.MarshalIndent(apiSubscriptionList, "", "  ")

	if err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write(stringified)

	return nil
}
