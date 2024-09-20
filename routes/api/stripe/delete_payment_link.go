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

type DeletePaymentLinkRequest struct {
	PaymentLinks []string `json:"payment_links"`
}

//	@title			Delete Payment Link
//	@version		1.0
//	@description	Delete a payment link

//	@BasePath	/api

//	@accept		json
//	@produce	json

//	@schemes	https

//	@securityDefinitions.apiKey	ApiKeyAuth

//	@Param	payment_links	body	[]string	false	"the list of payment link ids to delete"	

func DeletePaymentLink(w http.ResponseWriter, r *http.Request, db *base.Base) error {
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

	deletePaymentLinkRequest := DeletePaymentLinkRequest{}

	if err := json.Unmarshal(bytes, &deletePaymentLinkRequest); err != nil {
		return err
	}

	for _, id := range deletePaymentLinkRequest.PaymentLinks {
		paymentLinkRecord := PaymentLinkRecord{}

		err := linkDb.Get(id, &paymentLinkRecord)

		if err != nil {
			return err
		}

		fmt.Println(paymentLinkRecord)

		params := &stripe.PaymentLinkParams{
			Active: stripe.Bool(false),
		}

		params.SetStripeAccount(user.AccountId)

		_, err = paymentlink.Update(
			paymentLinkRecord.PaymentLinkId,
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
