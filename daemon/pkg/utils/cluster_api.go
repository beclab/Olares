package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/net/publicsuffix"
	"k8s.io/klog/v2"
)

const (
	clusterDNS  = "10.233.0.3"
	LLDAPPort   = 17170
	LLDAPServer = "lldap-service.os-platform.svc.cluster.local"

	Owner  string = "owner"
	Admin  string = "admin"
	Normal string = "normal"
)

type claims struct {
	jwt.StandardClaims
	// Private Claim Names
	// Username user identity, deprecated field
	Username string `json:"username,omitempty"`

	Groups []string `json:"groups,omitempty"`
	Mfa    int64    `json:"mfa,omitempty"`
}

type ValidToken struct {
	Username string
	Groups   []string
}

func GetClusterHttpClient() *http.Client {
	clusterResolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return net.Dial("udp", clusterDNS+":53")
		},
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{Resolver: clusterResolver}).DialContext,
	}
	client := &http.Client{Transport: transport}
	return client
}

func AccessTokenValidate(token string) (valid bool, tokenData *ValidToken, err error) {
	if len(token) == 0 {
		klog.Error("no token provided for verification")
		return false, nil, errors.New("no token provided for verification")
	}

	lldapAddr := fmt.Sprintf("http://%s:%d", LLDAPServer, LLDAPPort)
	_, err = TokenVerify(lldapAddr, token, token)
	if err != nil {
		klog.Errorf("failed to verify token: %v", err)
		return false, nil, err
	}

	// Token not found in cache, parse and validate the JWT token
	claims, err := parseToken(token)
	if err != nil {
		klog.Errorf("failed to parse token: %v", err)
		return false, nil, err
	}

	// get user groups from cluster
	client, err := GetDynamicClient()
	if err != nil {
		klog.Errorf("failed to get dynamic client: %v", err)
		return false, nil, err
	}
	role, err := GetUserRole(context.Background(), claims.Username, client)
	if err != nil {
		klog.Errorf("failed to get user role: %v", err)
		return false, nil, err
	}

	return true, &ValidToken{
		Username: claims.Username,
		Groups:   []string{role},
	}, nil
}

func parseToken(token string) (*claims, error) {
	if len(token) == 0 {
		return nil, errors.New("token is empty")
	}

	// Parse the JWT token with claims and without claims validation
	parsedToken, err := jwt.ParseWithClaims(token, &claims{}, nil, jwt.WithoutClaimsValidation())

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			switch {
			case ve.Errors&jwt.ValidationErrorMalformed != 0:
				return nil, fmt.Errorf("malformed token: %w", err)
			case ve.Errors&jwt.ValidationErrorExpired != 0:
				return nil, fmt.Errorf("token expired: %w", err)
			case ve.Errors&jwt.ValidationErrorSignatureInvalid != 0:
				return nil, fmt.Errorf("invalid token signature: %w", err)
			case ve.Errors&jwt.ValidationErrorUnverifiable != 0:
				// do not need verify the token signature
			default:
				return nil, fmt.Errorf("token validation error: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse token: %w", err)
		}
	}

	c, ok := parsedToken.Claims.(*claims)
	if !ok {
		return nil, errors.New("failed to extract claims from token")
	}

	return c, nil
}

func TokenVerify(baseURL, accessToken, validToken string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/auth/token/verify", baseURL)
	cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	httpClient := GetClusterHttpClient()
	client := resty.NewWithClient(httpClient)
	client.SetCookieJar(cookieJar)

	resp, err := client.SetTimeout(10*time.Second).R().
		SetHeader("Content-Type", "application/json").SetAuthToken(accessToken).
		SetBody(map[string]string{
			"access_token": validToken,
		}).Post(url)
	if err != nil {
		klog.Infof("send request failed: %v", err)
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		klog.Infof("not 200, %v, body: %v", resp.StatusCode(), string(resp.Body()))
		return nil, errors.New(resp.String())
	}
	var response map[string]interface{}
	err = json.Unmarshal(resp.Body(), &response)
	if err != nil {
		klog.Infof("unmarshal failed: %v", err)
		return nil, err
	}
	klog.Infof("token verify res: %v", response)

	if status, ok := response["status"]; ok && status == "invalid token" {
		klog.Infof("token verify failed, status: %s", status)
		return nil, errors.New("token verification failed")
	}
	return response, nil
}

func (t *ValidToken) IsAdmin() bool {
	for _, group := range t.Groups {
		switch group {
		case Admin, Owner:
			return true
		}
	}
	return false
}

func (t *ValidToken) IsOwner() bool {
	for _, group := range t.Groups {
		if group == Owner {
			return true
		}
	}
	return false
}
