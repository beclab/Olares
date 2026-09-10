// Package olarescli implements the platform side of permission.loginOlaresCLI:
// deriving a long-lived Olares credential for the user installing an app,
// writing it into the app's namespace for the pod webhook to mount, and
// revoking it on uninstall.
//
// The credential is a refresh token minted by lldap. It deliberately does not
// include an access token: those live for a day, and one sitting in a mounted
// directory would be stale by the time most consumers read it.
package olarescli

import (
	"context"
	"fmt"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
)

const (
	// lldapBaseURL is the in-cluster address of lldap, the same one the user
	// controller talks to.
	lldapBaseURL = "http://lldap-service.os-platform:17170"

	// deriveTTLDays is the lifetime we ask lldap for. lldap clamps it to its
	// own configured maximum, so a platform that wants shorter grants can say
	// so without app-service changing.
	deriveTTLDays = 3650

	requestTimeout = 10 * time.Second
)

// Grant is one derived credential. RefreshToken is the revocation handle
// and, together with OlaresID and AppName, the JSON file mounted into the app.
type Grant struct {
	RefreshToken string `json:"refresh_token"`
	Username     string `json:"username"`
	ExpiresAt    string `json:"expires_at"`
	OlaresID     string `json:"olaresId,omitempty"`
	AppName      string `json:"appName,omitempty"`
}

// Client talks to lldap's /auth/token/derive endpoints. Authentication is
// app-service's own Kubernetes ServiceAccount token: lldap checks it with a
// TokenReview and matches the resulting identity against its allowlist.
//
// The user is named in the request rather than proving possession of their
// lldap token, because app-service never holds one: its own callers are
// authenticated by the X-Bfl-User header the gateway sets. lldap already
// extends the same trust to the CLI's ServiceAccount for password resets.
type Client struct {
	baseURL string
	saToken string
	http    *resty.Client
}

func NewClient(saToken string) *Client {
	return &Client{
		baseURL: lldapBaseURL,
		saToken: saToken,
		http:    resty.New().SetTimeout(requestTimeout),
	}
}

// Derive mints a long-lived refresh token belonging to username.
func (c *Client) Derive(ctx context.Context, username, label string) (*Grant, error) {
	var grant Grant
	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetAuthToken(c.saToken).
		SetBody(map[string]interface{}{
			"username": username,
			"ttl_days": deriveTTLDays,
			"label":    label,
		}).
		SetResult(&grant).
		Post(c.baseURL + "/auth/token/derive")
	if err != nil {
		return nil, fmt.Errorf("derive olares-cli credential: %w", err)
	}
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("derive olares-cli credential for %s: lldap returned %d: %s",
			username, resp.StatusCode(), resp.String())
	}
	if grant.RefreshToken == "" {
		return nil, fmt.Errorf("derive olares-cli credential: incomplete response from lldap")
	}
	return &grant, nil
}

// Revoke drops one grant by presenting the refresh token itself. lldap treats
// an unknown token as success, so callers on a retry path do not need to
// distinguish "already revoked".
func (c *Client) Revoke(ctx context.Context, refreshToken string) error {
	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetAuthToken(c.saToken).
		SetBody(map[string]string{"refresh_token": refreshToken}).
		Post(c.baseURL + "/auth/token/derive/revoke")
	if err != nil {
		return fmt.Errorf("revoke olares-cli credential: %w", err)
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("revoke olares-cli credential: lldap returned %d: %s",
			resp.StatusCode(), resp.String())
	}
	return nil
}
