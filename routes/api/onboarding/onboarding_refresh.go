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

func Refresh(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	stripe.Key = os.Getenv("STRIPE_TOKEN")

	user := r.Context().Value("user").(auth.SnailUser)

	account, err := account.GetByID(
		user.AccountId,
		nil,
	)

	if err != nil {
		return err
	}

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
