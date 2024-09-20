package stripe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/deta/deta-go/deta"
	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentlink"
)

type DeleteSubscriptionLinkRequest struct {
	PaymentLinks []string `json:"payment_links"`
}

func DeleteSubscriptionLink(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	stripe.Key = os.Getenv("STRIPE_TOKEN")

	deta, err := deta.New(deta.WithProjectKey(os.Getenv("PROJECT_KEY")))

	if err != nil {
		return err
	}

	linkDb, err := base.New(deta, "links")

	if err != nil {
		return err
	}

	user := r.Context().Value("user").(auth.SnailUser)

	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	deleteSubscriptionLinkRequest := DeletePaymentLinkRequest{}

	if err := json.Unmarshal(bytes, &deleteSubscriptionLinkRequest); err != nil {
		return err
	}

	for _, id := range deleteSubscriptionLinkRequest.PaymentLinks {
		subscriptionLinkRecord := SubscriptionLinkRecord{}

		err := linkDb.Get(id, &subscriptionLinkRecord)

		if err != nil {
			return err
		}

		fmt.Println(subscriptionLinkRecord)

		params := &stripe.PaymentLinkParams{
			Active: stripe.Bool(false),
		}

		params.SetStripeAccount(user.AccountId)

		_, err = paymentlink.Update(
			subscriptionLinkRecord.PaymentLinkId,
			params,
		)

		if err != nil {
			return err
		}

		err = linkDb.Delete(id)

		if err != nil {
			return err
		}
	}

	w.WriteHeader(200)
	w.Write([]byte("The selected links were deleted successfully."))

	return nil
}
