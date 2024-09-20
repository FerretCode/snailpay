package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/deta/deta-go/service/base"
)

func CheckAuth(db base.Base) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("snail")

			if err != nil {
				if errors.Is(http.ErrNoCookie, err) {
					apiKey := r.Header.Get("X-API-Key")

					if apiKey == "" {
						http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)		

						return
					}

					hash := sha256.Sum256([]byte(apiKey))

					encoded := base64.StdEncoding.EncodeToString(hash[:])

					var results []map[string]interface{}

					_, err = db.Fetch(&base.FetchInput{
						Q: base.Query{
							{"api_key_hashes?contains": encoded},
						},
						Dest: &results,
					})

					if err != nil {
						fmt.Println(err)

						http.Error(w, "Your API key is incorrect!", http.StatusUnauthorized)

						return
					}

					if len(results) < 1 {
						http.Error(w, "Your API key is incorrect!", http.StatusUnauthorized)

						return
					}

					var hashesSymbols []string
					var namesSymbols []string

					if reflect.ValueOf(results[0]["api_key_hashes"]).Len() > 0 {
						apiKeyHashes := results[0]["api_key_hashes"].([]interface{})
						apiKeyNames := results[0]["api_key_names"].([]interface{})

						hashesSymbols = make([]string, len(apiKeyHashes))
						namesSymbols = make([]string, len(apiKeyNames))

						for i, arg := range apiKeyHashes { hashesSymbols[i] = arg.(string) }
						for i, arg := range apiKeyNames { namesSymbols[i] = arg.(string) }
					} 

					user := SnailUser{
						UserId: results[0]["key"].(string),
						RefreshToken: results[0]["refresh_token"].(string),
						Registered: results[0]["registered"].(bool),
						AccountId: results[0]["account_id"].(string),
						BankAccountId: results[0]["bank_account_id"].(string),
						APIKeyHashes: hashesSymbols,
						APIKeyNames: namesSymbols,
					}

					ctx := context.WithValue(r.Context(), "user", user)

					next.ServeHTTP(w, r.WithContext(ctx))

					return
				}

				fmt.Println(err)

				http.Error(w, "There was an error processing your request.", http.StatusInternalServerError)

				return
			}

			_, err = GetSession(cookie.Value) 

			if err != nil {
				if errors.Is(ErrNotAuthenticated, err) {
					http.Redirect(w, r, "/auth/login", http.StatusTemporaryRedirect)		

					return
				} else {
					fmt.Println(err)

					http.Error(w, "There was an error processing your request.", http.StatusInternalServerError)

					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
