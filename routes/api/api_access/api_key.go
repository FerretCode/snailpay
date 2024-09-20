package apiaccess

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"github.com/deta/deta-go/service/base"
	"github.com/ferretcode/snail/routes/auth"
)

type NewKeyRequest struct {
	Key string `json:"key"`
	Name string `json:"name"`
}

func NewAPIKey(w http.ResponseWriter, r *http.Request, db *base.Base) error {
	user := r.Context().Value("user").(auth.SnailUser)

	bytes, err := io.ReadAll(r.Body)

	if err != nil {
		return err
	}

	newKeyRequest := NewKeyRequest{}

	if err := json.Unmarshal(bytes, &newKeyRequest); err != nil {
		return err
	}

	hash := sha256.Sum256([]byte("snail_" + newKeyRequest.Key))

	encoded := base64.StdEncoding.EncodeToString(hash[:])

	err = db.Update(user.UserId, base.Updates{
		"api_key_hashes": append(
			user.APIKeyHashes, 
			encoded,
		),
		"api_key_names": append(
			user.APIKeyNames,
			newKeyRequest.Name,
		),
	})

	w.WriteHeader(200)
	w.Write([]byte("The API key has been generated."))

	return nil
}
