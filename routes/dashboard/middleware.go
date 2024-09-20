package dashboard

import (
	"context"
	"fmt"
	"net/http"

	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
)

func GetUser(db base.Base) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contextUser := r.Context().Value("user")

			if contextUser != nil {
				ctx := context.WithValue(r.Context(), "user", contextUser)

				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}

			cookie, err := r.Cookie("snail")		

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an error processing your request.", http.StatusInternalServerError)

				return
			}

			session, err := auth.GetSession(cookie.Value)

			user := auth.SnailUser{}

			err = db.Get(session.Session["user_id"].(string), &user)

			if err != nil {
				fmt.Println(err)

				http.Error(w, "There was an error processing your request.", http.StatusInternalServerError)

				return
			}

			ctx := context.WithValue(r.Context(), "user", user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
