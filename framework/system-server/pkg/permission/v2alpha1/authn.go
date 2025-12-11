package v2alpha1

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"bytetrade.io/web3os/system-server/pkg/constants"
	"github.com/brancz/kube-rbac-proxy/pkg/authn"
	"github.com/jellydator/ttlcache/v3"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/request/union"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

var _ authenticator.Request = (*lldapTokenAuthenticator)(nil)

var tokenCacheTTL = time.Minute * 5

type lldapTokenAuthenticator struct {
	tokenCache  *ttlcache.Cache[string, *Claims]
	lldapServer string
}

// AuthenticateRequest implements authenticator.Request.
func (l *lldapTokenAuthenticator) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	cookie, err := req.Cookie(constants.AuthTokenCookieName)
	var token string
	if err != nil {
		if err != http.ErrNoCookie {
			return nil, false, fmt.Errorf("error retrieving cookie: %w", err)
		}
	}

	if err == nil && cookie != nil {
		token = cookie.Value
	}

	if token == "" {
		token = req.Header.Get(constants.AuthorizationTokenKey)
	}

	if token == "" {
		return nil, false, nil // No token found
	}

	claims := l.tokenCache.Get(token)
	if claims != nil {
		klog.Info("found token in cache")
		return &authenticator.Response{
			User: &user.DefaultInfo{
				Name:   claims.Value().Username,
				Groups: claims.Value().Groups,
				UID:    claims.Value().Username,
			},
		}, true, nil
	}

	// verify token
	res, err := TokenVerify(l.lldapServer, token, token)
	if err != nil {
		klog.Errorf("Token verification failed: %v", err)
		return nil, false, fmt.Errorf("token verification failed: %w", err)
	}

	klog.Info("token verified in lldap successfully, ", res)

	c, err := parseToken(token)
	if err != nil {
		klog.Errorf("failed to parse token: %v", err)
		return nil, false, fmt.Errorf("failed to parse token: %w", err)
	}

	response := authenticator.Response{
		User: &user.DefaultInfo{
			Name:   c.Username,
			Groups: c.Groups,
			UID:    c.Username,
		},
	}

	l.tokenCache.Set(token, c, tokenCacheTTL)
	return &response, true, nil
}

var _ authenticator.Request = (*autheliaNonceAuthenticator)(nil)

type autheliaNonceAuthenticator struct {
}

// AuthenticateRequest implements authenticator.Request.
func (a *autheliaNonceAuthenticator) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	nonce := req.Header.Get(constants.AutheliaNonceKey)
	if nonce == "" {
		return nil, false, nil // No nonce found
	}

	if nonce != constants.Nonce {
		return nil, false, fmt.Errorf("invalid nonce")
	}

	u := req.Header.Get(constants.BflUserKey)
	if u == "" {
		return nil, false, fmt.Errorf("no user found in header")
	}

	klog.Info("authelia nonce verified successfully for user: ", u)
	response := authenticator.Response{
		User: &user.DefaultInfo{
			Name:   u,
			Groups: []string{"authelia:backend"},
			UID:    u,
		},
	}

	return &response, true, nil
}

func UnionAllAuthenticators(ctx context.Context, cfg *AuthnConfig, kubeClient kubernetes.Interface) (authenticator.Request, error) {
	var authenticator authenticator.Request

	// If OIDC configuration provided, use oidc authenticator
	if cfg.OIDC.IssuerURL != "" {
		oidcAuthenticator, err := authn.NewOIDCAuthenticator(ctx, cfg.OIDC)
		if err != nil {
			return nil, fmt.Errorf("failed to instantiate OIDC authenticator: %w", err)
		}

		go oidcAuthenticator.Run(ctx)
		authenticator = oidcAuthenticator
	} else {
		//Use Delegating authenticator
		klog.Infof("Valid token audiences: %s", strings.Join(cfg.Token.Audiences, ", "))

		tokenClient := kubeClient.AuthenticationV1()
		delegatingAuthenticator, err := authn.NewDelegatingAuthenticator(tokenClient, &cfg.AuthnConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to instantiate delegating authenticator: %w", err)
		}

		go delegatingAuthenticator.Run(ctx)
		authenticator = delegatingAuthenticator
	}

	return union.New(
		&lldapTokenAuthenticator{ttlcache.New(
			ttlcache.WithTTL[string, *Claims](tokenCacheTTL),
			ttlcache.WithCapacity[string, *Claims](1000),
		), fmt.Sprintf("http://%s:%d", cfg.LLDAP.Server, cfg.LLDAP.Port),
		},
		&autheliaNonceAuthenticator{},
		authenticator), nil
}
