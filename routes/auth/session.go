package auth

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/redis/go-redis/v9"
)

type Session struct {
	Id string `json:"cookie"`
	Session map[string]interface{} `json:"session"`
}

var ErrNotAuthenticated = errors.New("this user is not authenticated")

func GetSession(id string) (Session, error) {
	ctx := context.Background()

	host := os.Getenv("SESSION_STORAGE_HOST")

	url, err := redis.ParseURL(host)

	if err != nil {
		return Session{}, err
	}

	rdb := redis.NewClient(url)

	defer rdb.Close()

	result, err := rdb.Get(ctx, id).Result()

	if err == redis.Nil {
		return Session{}, ErrNotAuthenticated
	}

	if err != nil {
		return Session{}, err
	}

	session := Session{}

	if err := json.Unmarshal([]byte(result), &session); err != nil {
		return Session{}, nil
	}

	return session, nil
}

func CreateSession(id string, accessToken string, userId string) error {
	ctx := context.Background()

	host := os.Getenv("SESSION_STORAGE_HOST")

	url, err := redis.ParseURL(host)

	if err != nil {
		return err
	}

	rdb := redis.NewClient(url)

	defer rdb.Close()

	session := Session{
		Session: map[string]interface{}{
			"access_token": accessToken,
			"user_id": userId,
		},
	}

	stringified, err := json.Marshal(session)

	if err != nil {
		return err
	}

	err = rdb.Set(ctx, id, string(stringified), 0).Err()

	if err != nil {
		return err
	}

	return nil
} 
