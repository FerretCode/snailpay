package stripe

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
	"github.com/stripe/stripe-go/v74/refund"
)

type RefundPaymentRequest struct {
	Payments []string `json:"payments"`
}

//	@title			Refund Payment 
//	@version		1.0
//	@description	Refund a payment 

//	@BasePath	/api

//	@accept		json
//	@produce	json

//	@schemes	https

//	@securityDefinitions.apiKey	ApiKeyAuth

//	@Param	payments	body	[]string	false	"a collection of payment ids to refund"	

func RefundPayment(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	stripe.Key = os.Getenv("STRIPE_TOKEN")

	user := r.Context().Value("user").(auth.SnailUser)

	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	refundPaymentRequest := RefundPaymentRequest{}

	if err := json.Unmarshal(bytes, &refundPaymentRequest); err != nil {
		return err
	}

	for _, id := range refundPaymentRequest.Payments {
		params := &stripe.PaymentIntentParams{}

		params.SetStripeAccount(user.AccountId)

		paymentIntent, err := paymentintent.Get(id, params)

		if err != nil {
			return err
		}

		refundParams := stripe.RefundParams{
			PaymentIntent: stripe.String(paymentIntent.ID),
		}

		refundParams.SetStripeAccount(user.AccountId)

		_, err = refund.New(&refundParams)

		if err != nil {
			return err
		}
	}

	w.WriteHeader(200)
	w.Write([]byte("The selected links were deleted successfully."))

	return nil
}
