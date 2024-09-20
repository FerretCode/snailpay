package auth

import (
	"net/http"
	"os"

	"github.com/deta/deta-go/service/base"
)

func Login(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	http.Redirect(w, r, os.Getenv("DISCORD_OAUTH_URL"), http.StatusTemporaryRedirect)

	return nil
}
