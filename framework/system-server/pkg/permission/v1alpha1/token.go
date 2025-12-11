package permission

import (
	"context"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Options struct {
	JwtSecret string `yaml:"jwtSecret"`
}

type Config struct {
	AuthenticationOptions *Options `yaml:"authentication,omitempty"`
}

type Type string

type Claims struct {
	jwt.StandardClaims
	// Private Claim Names
	// TokenType defined the type of the token
	TokenType Type `json:"token_type,omitempty"`
	// Username user identity, deprecated field
	Username string `json:"username,omitempty"`
	// Extra contains the additional information
	Extra map[string][]string `json:"extra,omitempty"`

	// Used for issuing authorization code
	// Scopes can be used to request that specific sets of information be made available as Claim Values.
	Scopes []string `json:"scopes,omitempty"`

	// The following is well-known ID Token fields

	// End-User's full name in displayable form including all name parts,
	// possibly including titles and suffixes, ordered according to the End-User's locale and preferences.
	Name string `json:"name,omitempty"`
	// String value used to associate a Client session with an ID Token, and to mitigate replay attacks.
	// The value is passed through unmodified from the Authentication Request to the ID Token.
	Nonce string `json:"nonce,omitempty"`
	// End-User's preferred e-mail address.
	Email string `json:"email,omitempty"`
	// End-User's locale, represented as a BCP47 [RFC5646] language tag.
	Locale string `json:"locale,omitempty"`
	// Shorthand name by which the End-User wishes to be referred to at the RP,
	PreferredUsername string `json:"preferred_username,omitempty"`
}

func validateToken(ctx context.Context, kubeConfig *rest.Config, tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		jwtSecretKey, err := getLLdapJwtKey(ctx, kubeConfig)
		if err != nil {
			return nil, err
		}
		return jwtSecretKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.Username, nil
	}
	return "", fmt.Errorf("invalid token, or claims not match")
}

func getLLdapJwtKey(ctx context.Context, kubeConfig *rest.Config) ([]byte, error) {
	kubeClientInService, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	secret, err := kubeClientInService.CoreV1().Secrets("os-platform").Get(ctx, "lldap-credentials", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	jwtSecretKey, ok := secret.Data["lldap-jwt-secret"]
	if !ok {
		return nil, fmt.Errorf("failed to get lldap jwt secret")
	}

	return jwtSecretKey, nil
}
