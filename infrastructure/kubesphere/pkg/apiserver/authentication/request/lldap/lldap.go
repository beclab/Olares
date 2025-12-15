package lldap

import (
	"errors"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"net/http"
	"strings"
)

type Authenticator struct {
	auth authenticator.Token
}

func New(auth authenticator.Token) *Authenticator {
	return &Authenticator{auth: auth}
}

var invalidToken = errors.New("invalid jwt token")

func (a *Authenticator) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	tokenString := strings.TrimSpace(req.Header.Get("Authorization"))
	if tokenString == "" {
		return nil, false, nil
	}

	resp, ok, err := a.auth.AuthenticateToken(req.Context(), tokenString)
	if !ok && err == nil {
		err = invalidToken
	}
	return resp, ok, err
}
