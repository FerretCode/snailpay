package dashboard

import (
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/deta/deta-go/deta"
	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
	snailStripe "github.com/ferretcode/snail/routes/api/stripe"
	"github.com/stripe/stripe-go/v74"
)

type paymentLinkList struct {
	PaymentLinkList []snailStripe.PaymentLinkRecord
	Host string
}

func PaymentLinkList(w http.ResponseWriter, r *http.Request, db *base.Base) error {
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

	var results []snailStripe.PaymentLinkRecord		

	_, err = linkDb.Fetch(&base.FetchInput{
		Q: base.Query{
			{"user_id": user.UserId, "subscription": false},
		},
		Dest: &results,
	})

	if err != nil {
		return err
	}

	tmpl, err := template.ParseFiles("templates/dashboard/stripe/payments/payment_link_list.html")

	if err != nil {
		return err
	}

	fmt.Println(results)

	tmpl.Execute(w, paymentLinkList{
		PaymentLinkList: results,
		Host: os.Getenv("HOST"),
	})

	return nil
}
