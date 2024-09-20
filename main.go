package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/dchest/uniuri"
	"github.com/deta/deta-go/deta"
	"github.com/deta/deta-go/service/base"
	apiaccess "github.com/ferretcode/snail/routes/api/api_access"
	"github.com/ferretcode/snail/routes/api/onboarding"
	"github.com/ferretcode/snail/routes/api/stripe"
	"github.com/ferretcode/snail/routes/auth"
	"github.com/ferretcode/snail/routes/dashboard"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	strp "github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/account"
)

type ShortenPaymentLink struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	deta, err := deta.New(deta.WithProjectKey(os.Getenv("PROJECT_KEY")))

	if err != nil {
		log.Fatal(err)
	}

	db, err := base.New(deta, "users")

	if err != nil {
		log.Fatal(err)
	}

	linkDb, err := base.New(deta, "links")

	if err != nil {
		log.Fatal(err)
	}

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.RealIP)

	fileserver := http.FileServer(http.Dir("./static"))
	router.Handle("/static/*", http.StripPrefix("/static/", fileserver))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("templates/index.html")

		if err != nil {
			http.Error(w, "There was an error loading the webpage.", http.StatusInternalServerError)

			return
		}

		tmpl.Execute(w, nil)
	})

	router.Get("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		err := auth.Login(w, r, db)

		if err != nil {
			fmt.Println(err)

			http.Error(w, "There was an issue logging you in.", http.StatusInternalServerError)
		}
	})

	router.Get("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		err := auth.Callback(w, r, db)

		if err != nil {
			fmt.Println(err)

			http.Error(w, "There was an issue logging you in.", http.StatusInternalServerError)
		}
	})

	router.Route("/api", func(r chi.Router) {
		r.Use(auth.CheckAuth(*db))
		r.Use(dashboard.GetUser(*db))

		r.Post("/subscription-link", func(w http.ResponseWriter, r *http.Request) {
			err := stripe.SubscriptionLink(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue creating your payment link.", http.StatusInternalServerError)
			}
		})

		r.Post("/delete-subscription-link", func(w http.ResponseWriter, r *http.Request) {
			err := stripe.DeletePaymentLink(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue deleting the payment link.", http.StatusInternalServerError)
			}
		})

		r.Post("/cancel-subscription", func(w http.ResponseWriter, r *http.Request) {
			err := stripe.CancelSubscription(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue canceling the selected payment link.", http.StatusInternalServerError)
			}
		})

		r.Post("/payment-link", func(w http.ResponseWriter, r *http.Request) {
			err := stripe.PaymentLink(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue creating your payment link.", http.StatusInternalServerError)
			}
		})

		r.Post("/delete-payment-link", func(w http.ResponseWriter, r *http.Request) {
			err := stripe.DeletePaymentLink(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue deleting the payment link.", http.StatusInternalServerError)
			}
		})

		r.Post("/shorten-payment-link", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-API-Key") != "" {
				http.Error(w, "You cannot use this resource!", http.StatusForbidden)

				return
			}

			shortenPaymentLink := ShortenPaymentLink{}

			bytes, err := io.ReadAll(r.Body)

			if err != nil {
				http.Error(w, "There was an error shortening the payment URL.", http.StatusInternalServerError)

				return
			}

			if err := json.Unmarshal(bytes, &shortenPaymentLink); err != nil {
				http.Error(w, "There was an error shortening the payment URL.", http.StatusInternalServerError)

				return
			}

			w.WriteHeader(200)
			w.Write([]byte(
				fmt.Sprintf("%s/%s", os.Getenv("HOST"), shortenPaymentLink.ID),
			))

			router.Get(fmt.Sprintf("/%s", shortenPaymentLink.ID), func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, shortenPaymentLink.URL, http.StatusFound)
			})
		})

		r.Post("/refund-payment", func(w http.ResponseWriter, r *http.Request) {
			err := stripe.RefundPayment(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue refunding the selected payment.", http.StatusInternalServerError)
			}
		})

		r.Post("/new-payout", func(w http.ResponseWriter, r *http.Request) {
			err := stripe.NewPayout(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue creating your payout.", http.StatusInternalServerError)
			}
		})

		r.Post("/new-api-key", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-API-Key") != "" {
				http.Error(w, "You cannot use this resource!", http.StatusForbidden)

				return
			}

			err := apiaccess.NewAPIKey(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an error generating the API key.", http.StatusInternalServerError)
			}
		})

		r.Post("/onboarding", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-API-Key") != "" {
				http.Error(w, "You cannot use this resource!", http.StatusForbidden)

				return
			}

			err := onboarding.StripeOnboarding(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue onboarding you.", http.StatusInternalServerError)
			}
		})

		r.Get("/onboarding/return", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-API-Key") != "" {
				http.Error(w, "You cannot use this resource!", http.StatusForbidden)

				return
			}

			err := onboarding.Return(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue onboarding you.", http.StatusInternalServerError)
			}
		})

		r.Get("/onboarding/refresh", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-API-Key") != "" {
				http.Error(w, "You cannot use this resource!", http.StatusForbidden)

				return
			}

			err := onboarding.Refresh(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue onboarding you.", http.StatusInternalServerError)
			}
		})

		r.Get("/payout", func(w http.ResponseWriter, r *http.Request) {
			err := stripe.Payout(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue getting your payouts.", http.StatusInternalServerError)
			}
		})

		r.Get("/payment-list", func(w http.ResponseWriter, r *http.Request) {
			err := stripe.PaymentList(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue listing your payments.", http.StatusInternalServerError)
			}
		})

		r.Get("/payment-link-list", func(w http.ResponseWriter, r *http.Request) {
			err := stripe.PaymentLinkList(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue listing your payment links.", http.StatusInternalServerError)
			}
		})

		r.Get("/subscription-list", func(w http.ResponseWriter, r *http.Request) {
			err := stripe.SubscriptionList(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue listing your subscriptions.", http.StatusInternalServerError)
			}
		})

		r.Get("/subscription-link-list", func(w http.ResponseWriter, r *http.Request) {
			err := stripe.SubscriptionLinkList(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue listing your subscription links.", http.StatusInternalServerError)
			}
		})

		r.Get("/verify-payment", func(w http.ResponseWriter, r *http.Request) {
			err := stripe.VerifyPayment(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue verifying the payment!", http.StatusInternalServerError)
			}
		})

		r.Post("/delete-api-key", func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-API-Key") != "" {
				http.Error(w, "You cannot use this resource!", http.StatusForbidden)

				return
			}

			err := apiaccess.DeleteAPIKey(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue deleting your API key!", http.StatusInternalServerError)
			}
		})
	})

	router.Route("/dashboard", func(r chi.Router) {
		r.Use(auth.CheckAuth(*db))
		r.Use(dashboard.GetUser(*db))

		r.Get("/home", func(w http.ResponseWriter, r *http.Request) {
			user := r.Context().Value("user").(auth.SnailUser)

			strp.Key = os.Getenv("STRIPE_TOKEN")

			if user.AccountId != "" {
				account, err := account.GetByID(user.AccountId, nil)

				if err != nil {
					fmt.Println(err)

					http.Error(w, "There was an error retreiving your account status.", http.StatusInternalServerError)

					return
				}

				if !user.Restricted && account.Requirements.DisabledReason != "" {
					err = db.Update(user.UserId, base.Updates{
						"restricted": true,
					})

					user.Restricted = true

					ctx := context.WithValue(r.Context(), "user", user)

					r.WithContext(ctx)

					if err != nil {
						fmt.Println(err)

						http.Error(w, "There was an error retreiving your account status.", http.StatusInternalServerError)

						return
					}
				}

				if user.Restricted && account.Requirements.DisabledReason == "" {
					err = db.Update(user.UserId, base.Updates{
						"restricted": false,
					})

					user.Restricted = false

					ctx := context.WithValue(r.Context(), "user", user)

					r.WithContext(ctx)

					if err != nil {
						fmt.Println(err)

						http.Error(w, "There was an error retreiving your account status.", http.StatusInternalServerError)

						return
					}
				}
			}

			tmpl, err := template.ParseFiles("templates/dashboard/index.html")

			if err != nil {
				http.Error(w, "There was an error loading the webpage.", http.StatusInternalServerError)

				return
			}

			tmpl.Execute(w, user)
		})

		r.Get("/onboarding", func(w http.ResponseWriter, r *http.Request) {
			user := r.Context().Value("user").(auth.SnailUser)

			tmpl, err := template.ParseFiles("templates/dashboard/onboarding.html")

			if err != nil {
				http.Error(w, "There was an error loading the webpage.", http.StatusInternalServerError)

				return
			}

			tmpl.Execute(w, user)
		})

		r.Get("/payment-link", func(w http.ResponseWriter, r *http.Request) {
			tmpl, err := template.ParseFiles("templates/dashboard/stripe/payments/payment_link.html")

			if err != nil {
				http.Error(w, "There was an error loading the webpage.", http.StatusInternalServerError)

				return
			}

			tmpl.Execute(w, nil)
		})

		r.Get("/payment-link-list", func(w http.ResponseWriter, r *http.Request) {
			err := dashboard.PaymentLinkList(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue listing your payment links.", http.StatusInternalServerError)
			}
		})

		r.Get("/payment-list", func(w http.ResponseWriter, r *http.Request) {
			err := dashboard.PaymentList(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue listing your payments.", http.StatusInternalServerError)
			}
		})

		r.Get("/subscription-link", func(w http.ResponseWriter, r *http.Request) {
			tmpl, err := template.ParseFiles("templates/dashboard/stripe/subscriptions/subscription_link.html")

			if err != nil {
				http.Error(w, "There was an error loading the webpage.", http.StatusInternalServerError)

				return
			}

			tmpl.Execute(w, nil)
		})

		r.Get("/subscription-link-list", func(w http.ResponseWriter, r *http.Request) {
			err := dashboard.SubscriptionLinkList(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue listing your subscription links.", http.StatusInternalServerError)
			}
		})

		r.Get("/subscription-list", func(w http.ResponseWriter, r *http.Request) {
			err := dashboard.SubscriptionList(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue listing your subscriptions.", http.StatusInternalServerError)
			}
		})

		r.Get("/payout", func(w http.ResponseWriter, r *http.Request) {
			err := dashboard.Payout(w, r, db)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an issue viewing your payouts.", http.StatusInternalServerError)
			}
		})

		r.Get("/api-key", func(w http.ResponseWriter, r *http.Request) {
			user := r.Context().Value("user").(auth.SnailUser)

			tmpl, err := template.ParseFiles("templates/dashboard/api/api_key.html")

			if err != nil {
				http.Error(w, "There was an error loading the webpage.", http.StatusInternalServerError)

				return
			}

			tmpl.Execute(w, user)
		})
	})

	router.Get("/redirect", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("session_id") == "" || r.URL.Query().Get("account_id") == "" {
			http.Error(w, "Not all of the requirements are provided!", http.StatusBadRequest)

			return
		}

		code := uniuri.NewLen(10)

		verificationCode := stripe.VerificationCode{
			Code:      code,
			SessionId: r.URL.Query().Get("session_id"),
			AccountId: r.URL.Query().Get("account_id"),
		}

		codeDb, err := base.New(deta, "codes")

		if err != nil {
			fmt.Println(err)

			http.Error(w, "There was an error loading the webpage.", http.StatusInternalServerError)

			return
		}

		_, err = codeDb.Put(verificationCode)

		if err != nil {
			fmt.Println(err)

			http.Error(w, "There was an error loading the webpage.", http.StatusInternalServerError)

			return
		}

		tmpl, err := template.ParseFiles("templates/redirect.html")

		if err != nil {
			fmt.Println(err)

			http.Error(w, "There was an error loading the webpage.", http.StatusInternalServerError)

			return
		}

		tmpl.Execute(w, code)
	})

	var results []map[string]interface{}

	last, err := linkDb.Fetch(&base.FetchInput{
		Q:    nil,
		Dest: &results,
	})

	if err != nil {
		log.Fatal(err)
	}

	for last != "" {
		_, err = linkDb.Fetch(&base.FetchInput{
			LastKey: last,
			Dest:    &results,
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	for _, v := range results {
		router.Get(fmt.Sprintf("/%s", v["key"].(string)), func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, v["url"].(string), http.StatusFound)
		})
	}

	http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", os.Getenv("PORT")), router)
}
