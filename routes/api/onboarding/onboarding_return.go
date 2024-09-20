package onboarding

import (
	"net/http"
	"os"

	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/account"
	"github.com/stripe/stripe-go/v74/accountlink"
)

func Return(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	stripe.Key = os.Getenv("STRIPE_TOKEN")

	user := r.Context().Value("user").(auth.SnailUser)

	err := db.Update(user.UserId, base.Updates{
		"registered": true,
		"payments_verified": false,
	})

	if err != nil {
		return err
	}

	account, err := account.GetByID(
		user.AccountId,
		nil,
	)

	if err != nil {
		return err
	}

	paymentsVerified := true

	if len(account.Requirements.CurrentlyDue) > 0 || len(account.Requirements.EventuallyDue) > 0 || len(account.Requirements.PastDue) > 0 {
		paymentsVerified = false
	}

	if !paymentsVerified {
		accountLinkParams := &stripe.AccountLinkParams{
			Account: stripe.String(account.ID),
			Type: stripe.String("account_onboarding"),
			ReturnURL: stripe.String(os.Getenv("STRIPE_RETURN_URL")),
			RefreshURL: stripe.String(os.Getenv("STRIPE_REFRESH_URL")),
		}

		accountLink, err := accountlink.New(accountLinkParams)

		if err != nil {
			return err
		}

		http.Redirect(w, r, accountLink.URL, http.StatusTemporaryRedirect)

		return nil 
	}

	err = db.Update(user.UserId, base.Updates{
		"registered": true,
		"payments_verified": true,
	})

	if err != nil {
		return err
	}

	http.Redirect(w, r, "/dashboard/home", http.StatusFound)

	return nil
}
