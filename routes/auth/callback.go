package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/deta/deta-go/deta"
	"github.com/deta/deta-go/service/base"
	"github.com/google/go-querystring/query"
	"github.com/google/uuid"
)

type User struct {
	Id string `json:"id"`
	Username string `json:"username"`
	Discriminator string `json:"discriminator"`
}

type SnailUser struct {
	UserId string `json:"key"`
	RefreshToken string `json:"refresh_token"`
	Registered bool `json:"registered"`
	Restricted bool `json:"restricted"`
	AccountId string `json:"account_id"`
	BankAccountId string `json:"bank_account_id"`
	APIKeyHashes []string `json:"api_key_hashes"`
	APIKeyNames []string `json:"api_key_names"` 
}

type TokenRequest struct {
	ClientId string `url:"client_id"`
	ClientSecret string `url:"client_secret"`
	GrantType string `url:"grant_type"`
	Code string `url:"code"`
	RedirectURI string `url:"redirect_uri"`
	RefreshToken string `url:"refresh_token"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn int `json:"expires_in"`
}

func Callback(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	code := r.URL.Query().Get("code")

	tokenRequest := TokenRequest{
		ClientId: os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
		GrantType: "authorization_code",
		Code: code,
		RedirectURI: os.Getenv("DISCORD_REDIRECT_URI"),
	}

	queryParams, err := query.Values(tokenRequest)

	fmt.Println(queryParams.Encode())

	if err != nil {
		return err
	}

	client := http.Client{}

	tokenReq, err := http.NewRequest(
		"POST",
		"https://discord.com/api/oauth2/token",
		strings.NewReader(queryParams.Encode()),
	)

	tokenReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	tokenRes, err := client.Do(tokenReq)

	if err != nil {
		return err
	}

	tokenResponse := TokenResponse{}

	if err := processBody(tokenRes.Body, &tokenResponse); err != nil {
		return err
	}

	if (time.Now().Unix() + int64(tokenResponse.ExpiresIn)) - time.Now().Unix()  < 86400 {
		cookie, err := r.Cookie("snail")

		if err != nil {
			return err
		}

		session, err := GetSession(cookie.Value)

		if err != nil {
			return err
		}

		user := SnailUser{}

		err = db.Get(session.Session["user_id"].(string), &user)

		if err != nil {
			return err
		}

		tokenRequest.RefreshToken = user.RefreshToken

		tokenReq, err := http.NewRequest(
			"POST",
			"https://discord.com/api/oauth2/token",
			strings.NewReader(queryParams.Encode()),
		)

		tokenReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		tokenRes, err := client.Do(tokenReq)

		if err != nil {
			return err
		}

		tr := TokenResponse{}

		if err := processBody(tokenRes.Body, &tr); err != nil {
			return err
		}

		tokenResponse = tr
	}

	userReq, err := http.NewRequest(
		"GET",
		"https://discord.com/api/users/@me",
		nil,
	)

	userReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenResponse.AccessToken))

	userRes, err := client.Do(userReq)

	if err != nil {
		return err
	}

	user := User{}

	if err := processBody(userRes.Body, &user); err != nil {
		return err
	}

	fmt.Println(user)

	snailUser := SnailUser{}

	err = db.Get(user.Id, &snailUser)

	if err != nil {
		if errors.Is(deta.ErrNotFound, err) { 
			snailUser.UserId = user.Id
			snailUser.RefreshToken = tokenResponse.RefreshToken
			snailUser.Registered = false
			snailUser.Restricted = false
			snailUser.AccountId = ""
			snailUser.BankAccountId = ""
			snailUser.APIKeyHashes = []string{}
			snailUser.APIKeyNames = []string{}

			_, err = db.Put(snailUser)
		} else {
			return err
		}
	}

	sid := uuid.NewString()

	cookie := http.Cookie{
		Name: "snail",
		Value: sid,
		Domain: os.Getenv("COOKIE_DOMAIN"),
		Path: "/",
	}

	http.SetCookie(w, &cookie)

	err = CreateSession(sid, tokenResponse.AccessToken, snailUser.UserId)

	if err != nil {
		return err
	}

	http.Redirect(w, r, "/dashboard/home", http.StatusFound)

	return nil	
}

func processBody(b io.ReadCloser, to interface{}) error {
	body, err := io.ReadAll(b)	

	if err != nil {
		return err
	}

	fmt.Println(string(body))

	if err := json.Unmarshal(body, &to); err != nil {
		return err
	}

	return nil
}
