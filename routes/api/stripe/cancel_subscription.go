package stripe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
	"github.com/stripe/stripe-go/v74"
	sub "github.com/stripe/stripe-go/v74/subscription"
)

type CancelSubscriptionRequest struct {
	Subscriptions []string `json:"subscriptions"`
}

func CancelSubscription(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	stripe.Key = os.Getenv("STRIPE_TOKEN")

	user := r.Context().Value("user").(auth.SnailUser)

	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	cancelSubscriptionRequest := CancelSubscriptionRequest{}

	if err := json.Unmarshal(bytes, &cancelSubscriptionRequest); err != nil {
		return err
	}

	fmt.Println(cancelSubscriptionRequest)

	for _, id := range cancelSubscriptionRequest.Subscriptions {
		params := &stripe.SubscriptionParams{}

		params.SetStripeAccount(user.AccountId)

		subscription, err := sub.Get(id, params)

		if err != nil {
			return err
		}

		subscriptionCancelParams := &stripe.SubscriptionCancelParams{}

		subscriptionCancelParams.SetStripeAccount(user.AccountId)

		_, err = sub.Cancel(subscription.ID, subscriptionCancelParams)

		if err != nil {
			return err
		}
	}

	w.WriteHeader(200)
	w.Write([]byte("The selected links were deleted successfully."))

	return nil
}
