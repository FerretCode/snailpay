package dashboard

import (
	"html/template"
	"net/http"
	"os"

	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/subscription"
)

type subscriptionList struct {
	SubscriptionList []stripe.Subscription
}

func SubscriptionList(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	stripe.Key = os.Getenv("STRIPE_TOKEN")

	user := r.Context().Value("user").(auth.SnailUser)

	sList := subscriptionList{}

	subscriptionListParams := stripe.SubscriptionListParams{}

	subscriptionListParams.SetStripeAccount(user.AccountId)
	subscriptionListParams.Expand = []*string{stripe.String("data.customer")}

	subscriptions := subscription.List(&subscriptionListParams)

	for subscriptions.Next() {
		sList.SubscriptionList = append(sList.SubscriptionList, *subscriptions.Subscription())
	}

	tmpl, err := template.ParseFiles("templates/dashboard/stripe/subscriptions/subscription_list.html")

	if err != nil {
		return err
	}

	tmpl.Execute(w, subscriptionList{
		SubscriptionList: sList.SubscriptionList,
	})

	return nil
}
