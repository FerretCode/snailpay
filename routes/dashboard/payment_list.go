package dashboard

import (
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/paymentintent"
)

type paymentList struct {
	PaymentList []stripe.PaymentIntent
}

func PaymentList(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	stripe.Key = os.Getenv("STRIPE_TOKEN")

	user := r.Context().Value("user").(auth.SnailUser)

	pList := paymentList{}

	paymentIntentListParams := stripe.PaymentIntentListParams{}

	paymentIntentListParams.SetStripeAccount(user.AccountId)
	paymentIntentListParams.Expand = []*string{stripe.String("data.customer")}

	paymentIntents := paymentintent.List(&paymentIntentListParams)

	for paymentIntents.Next() {
		pList.PaymentList = append(pList.PaymentList, *paymentIntents.PaymentIntent())
	}

	fmt.Println(pList.PaymentList)

	tmpl, err := template.ParseFiles("templates/dashboard/stripe/payments/payment_list.html")

	if err != nil {
		return err
	}

	tmpl.Execute(w, paymentList{
		PaymentList: pList.PaymentList,
	})

	return nil
}
