package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	validToken, err := MakeJWT(userID, "secret", time.Hour)
	if err != nil {
		fmt.Print(err)
	}

	test := struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserID  uuid.UUID
		wantErr     bool
	}{
		name:        "test01",
		tokenString: validToken,
		tokenSecret: "secret",
		wantUserID:  userID,
		wantErr:     false,
	}

	t.Run(test.name, func(t *testing.T) {
		goUserID, err := ValidateJWT(test.tokenString, test.tokenSecret)
		assert.Equal(t, err != nil, test.wantErr)
		assert.Equal(t, goUserID, test.wantUserID)
	})
}

func TestGetBearerToken(t *testing.T) {
	test := struct {
		name    string
		token   string
		headers map[string][]string
		wantErr bool
	}{
		name:    "test01",
		token:   "abc",
		headers: map[string][]string{"Authorization": {"Bearer abc"}},
		wantErr: false,
	}

	t.Run(test.name, func(t *testing.T) {
		token, err := GetBearerToken(test.headers)
		assert.Equal(t, err != nil, test.wantErr)
		assert.Equal(t, token, test.token)
	})
}
