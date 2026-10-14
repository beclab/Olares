package authn

import (
	"errors"
	"fmt"
	"net/http"

	apiRuntime "bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	"bytetrade.io/web3os/bfl/pkg/constants"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserMismatch = errors.New("request user does not match bfl owner")
)

// Identity is the authenticated caller represented by a request. Token is
// retained for handlers that must forward the caller's credentials upstream.
type Identity struct {
	Username string
	Token    string
}

type tokenParser func(string) (string, error)

// Authenticate resolves the caller in the same order as the historical BFL
// filter: a trusted X-BFL-USER header wins, otherwise the caller is read from
// the JWT. WebSocket clients cannot set X-Authorization, so auth_token is also
// accepted as a token source.
func Authenticate(req *http.Request) (Identity, error) {
	return authenticate(req, func(token string) (string, error) {
		claims, err := apiRuntime.ParseToken(token)
		if err != nil {
			return "", err
		}
		return claims.Username, nil
	})
}

func authenticate(req *http.Request, parse tokenParser) (Identity, error) {
	token := req.Header.Get(constants.UserAuthorizationTokenKey)
	if token == "" {
		if cookie, err := req.Cookie("auth_token"); err == nil {
			token = cookie.Value
		}
	}

	username := req.Header.Get(constants.HeaderBflUserKey)
	if token != "" {
		tokenUsername, err := parse(token)
		if err != nil {
			return Identity{}, fmt.Errorf("parse token: %w", err)
		}
		if username != "" && username != tokenUsername {
			return Identity{}, fmt.Errorf("%w: header user %q, token user %q", ErrUserMismatch, username, tokenUsername)
		}
		username = tokenUsername
	}
	if username == "" {
		return Identity{}, ErrUserNotFound
	}

	if constants.Username != "" && username != constants.Username {
		return Identity{}, fmt.Errorf("%w: got %q, want %q", ErrUserMismatch, username, constants.Username)
	}

	return Identity{Username: username, Token: token}, nil
}
