package wizard

import (
	"fmt"
	"time"
)

// SSI authentication client implementation
type SSIAuthClient struct {
	UserStore *UserStore // Direct use of UserStore struct
	// JWSSigner removed as UserStore.SignJWS() is actually used
}

// PrepareAuthentication implements authentication functionality for SSI client
func (p *SSIAuthClient) PrepareAuthentication(params map[string]any) (map[string]any, error) {

	// Extract challenge
	challenge, ok := params["challenge"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid challenge format")
	}

	challengeValue, ok := challenge["value"].(string)
	if !ok {
		return nil, fmt.Errorf("challenge value not found")
	}

	// Build JWS payload
	payload := map[string]any{
		"name":      p.UserStore.GetTerminusName(),
		"did":       p.UserStore.GetDid(),
		"domain":    "http://example.domain",
		"time":      fmt.Sprintf("%d", time.Now().UnixMilli()),
		"challenge": challengeValue,
	}

	// Sign JWS
	jws, err := p.UserStore.SignJWS(payload)
	if err != nil || jws == "" {
		return nil, fmt.Errorf("jws signing failed: %v", err)
	}

	return map[string]any{
		"jws": jws,
	}, nil
}
