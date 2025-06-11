// Package jwk implements a subset of the JSON Web Key spec (https://tools.ietf.org/html/rfc7517)
package jwk

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
)

// JWK represents a JSON Web Key as per RFC7517 (https://tools.ietf.org/html/rfc7517)
// Note that this is a subset of the spec. There are a handful of properties that the
// spec allows for that are not represented here at the moment. This is because we
// only need a subset of the spec for our purposes.
type JWK struct {
	ALG string `json:"alg,omitempty"`
	KID string `json:"kid,omitempty"`
	USE string `json:"use,omitempty"`
	KTY string `json:"kty,omitempty"`
	CRV string `json:"crv,omitempty"`
	D   string `json:"d,omitempty"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
}

// ComputeThumbprint computes the JWK thumbprint as per RFC7638 (https://tools.ietf.org/html/rfc7638)
func (j JWK) ComputeThumbprint() (string, error) {
	thumbprintPayload := map[string]interface{}{
		"crv": j.CRV,
		"kty": j.KTY,
		"x":   j.X,
	}

	if j.Y != "" {
		thumbprintPayload["y"] = j.Y
	}

	bytes, err := json.Marshal(thumbprintPayload)
	if err != nil {
		return "", err
	}

	digest := sha256.Sum256(bytes)
	thumbprint := base64.RawURLEncoding.EncodeToString(digest[:])

	return thumbprint, nil
}
