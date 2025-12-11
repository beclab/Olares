package v2alpha1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v4"
	"k8s.io/klog/v2"
)

type Claims struct {
	jwt.StandardClaims
	// Private Claim Names
	// Username user identity, deprecated field
	Username string `json:"username,omitempty"`

	Groups []string `json:"groups,omitempty"`
	Mfa    int64    `json:"mfa,omitempty"`
}

func TokenVerify(baseURL, accessToken, validToken string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/auth/token/verify", baseURL)
	client := resty.New()

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

func parseToken(token string) (*Claims, error) {
	if len(token) == 0 {
		return nil, errors.New("token is empty")
	}

	// Parse the JWT token with claims and without claims validation
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, nil, jwt.WithoutClaimsValidation())

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

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok {
		return nil, errors.New("failed to extract claims from token")
	}

	return claims, nil
}
