package apiaccess

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
)

type DeleteKeyRequest struct {
	Name string `json:"name"`
}

func DeleteAPIKey(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	user := r.Context().Value("user").(auth.SnailUser)

	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	deleteKeyRequest := DeleteKeyRequest{}

	if err := json.Unmarshal(bytes, &deleteKeyRequest); err != nil {
		return err
	}

	index := 0

	for i, v := range user.APIKeyNames {
		if v == deleteKeyRequest.Name { index = i } 
	}

	err = db.Update(user.UserId, base.Updates{
		"api_key_hashes": append(
			user.APIKeyHashes[:index], 
			user.APIKeyHashes[index + 1:]...,
		),
		"api_key_names": append(
			user.APIKeyNames[:index],
			user.APIKeyNames[index + 1:]...,
		),
	})

	w.WriteHeader(200)
	w.Write([]byte("The API key has been deleted."))

	return nil
}
