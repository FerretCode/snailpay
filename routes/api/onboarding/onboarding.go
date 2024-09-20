package onboarding

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/account"
	"github.com/stripe/stripe-go/v74/accountlink"
	"github.com/stripe/stripe-go/v74/bankaccount"
	"github.com/stripe/stripe-go/v74/token"
)

type OnboardingRequest struct {
	Email string `json:"email"`
	CountryCode string `json:"country_code"`
	Name string `json:"name"`
	RoutingNumber string `json:"routing_number"`
	AccountNumber string `json:"account_number"`
}

func StripeOnboarding(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	user := r.Context().Value("user").(auth.SnailUser)

	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	onboardingRequest := OnboardingRequest{}

	if err := json.Unmarshal(bytes, &onboardingRequest); err != nil {
		return err
	}

	stripe.Key = os.Getenv("STRIPE_TOKEN")

	var acct stripe.Account

	if user.AccountId == "" {
		accountParams := &stripe.AccountParams{
			Capabilities: &stripe.AccountCapabilitiesParams{
				Transfers: &stripe.AccountCapabilitiesTransfersParams{
					Requested: stripe.Bool(true),
				},
				CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{
					Requested: stripe.Bool(true),
				},
			},
			Country: stripe.String(onboardingRequest.CountryCode),
			Email: stripe.String(onboardingRequest.Email),
			Type: stripe.String("custom"),
		}	

		account, err := account.New(accountParams)

		if err != nil {
			return err
		}

		acct = *account
	} else {
		account, err := account.GetByID(
			user.AccountId,
			nil,
		)

		if err != nil {
			return err
		}

		acct = *account
	}

	token, err := token.New(&stripe.TokenParams{
		BankAccount: &stripe.BankAccountParams{
			Country: &acct.Country,
			Currency: stripe.String(string(stripe.CurrencyUSD)),		
			AccountHolderName: &onboardingRequest.Name,
			AccountHolderType: stripe.String(string(stripe.BankAccountAccountHolderTypeIndividual)),
			RoutingNumber: stripe.String(onboardingRequest.RoutingNumber),	
			AccountNumber: stripe.String(onboardingRequest.AccountNumber),
		},
	})

	if err != nil {
		return err
	}

	bankAccountParams := &stripe.BankAccountParams{
		Account: stripe.String(acct.ID),
		Token: stripe.String(token.ID),
	}

	bankAccount, err := bankaccount.New(bankAccountParams)

	if err != nil {
		return err
	}

	err = db.Update(user.UserId, base.Updates{
		"account_id": acct.ID,
		"bank_account_id": bankAccount.ID,
	})

	if err != nil {
		return err
	}

	accountLinkParams := &stripe.AccountLinkParams{
		Account: stripe.String(acct.ID),
		Type: stripe.String("account_onboarding"),
		ReturnURL: stripe.String(os.Getenv("STRIPE_RETURN_URL")),
		RefreshURL: stripe.String(os.Getenv("STRIPE_REFRESH_URL")),
	}

	accountLink, err := accountlink.New(accountLinkParams)

	if err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write([]byte(accountLink.URL))

	return nil
}
