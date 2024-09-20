package stripe

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/deta/deta-go/deta"
	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
	"github.com/stripe/stripe-go/v74"
)

type paymentLinkList struct {
	PaymentLinkList []PaymentLinkRecord
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

	var results []PaymentLinkRecord		

	_, err = linkDb.Fetch(&base.FetchInput{
		Q: base.Query{
			{"user_id": user.UserId, "subscription": false},
		},
		Dest: &results,
	})

	if err != nil {
		return err
	}

	for i, result := range results {
		result.URL = fmt.Sprintf("%s/%s", os.Getenv("HOST"), result.ID)

		results[i] = result
	}

	stringified, err := json.MarshalIndent(results, "", "  ")

	if err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write(stringified)

	return nil
}
