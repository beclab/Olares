package lldap

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	corev1 "k8s.io/client-go/listers/core/v1"
	autht "kubesphere.io/kubesphere/pkg/apiserver/authentication/token"
)

type jwtAuthInterface interface {
	AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error)
}

type jwtAuthenticator struct {
	secretLister corev1.SecretLister
}

func NewJwtAuthenticator(secretLister corev1.SecretLister) authenticator.Token {
	return &jwtAuthenticator{
		secretLister: secretLister,
	}
}

func (j *jwtAuthenticator) AuthenticateToken(ctx context.Context, tokenString string) (*authenticator.Response, bool, error) {
	token, err := jwt.ParseWithClaims(tokenString, &autht.Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		secret, err := j.secretLister.Secrets("os-platform").Get("lldap-credentials")
		if err != nil {
			return nil, err
		}
		jwtSecretKey := secret.Data["lldap-jwt-secret"]
		return jwtSecretKey, nil
	})

	if err != nil {
		return nil, false, err
	}

	if claims, ok := token.Claims.(*autht.Claims); ok && token.Valid {
		return &authenticator.Response{
			User: &user.DefaultInfo{
				Name: claims.Username,
				UID:  claims.Username,
			},
		}, true, nil
	}
	return nil, false, fmt.Errorf("invalid token, or claims not match")

}
