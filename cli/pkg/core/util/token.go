package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

var header = JWTHeader{
	Alg: "HS256",
	Typ: "JWT",
}

var payload = JWTPayload{
	Email:     "admin@kubesphere.io",
	Username:  "admin",
	TokenType: "static_token",
}

type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type JWTPayload struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"`
}

func EncryptToken(secret string) (string, error) {
	headerJson, _ := json.Marshal(header)
	headerBase64 := Base64URLEncode(headerJson)

	payloadJson, _ := json.Marshal(payload)
	payloadBase64 := Base64URLEncode(payloadJson)

	headerPayload := fmt.Sprintf("%s.%s", headerBase64, payloadBase64)

	var secretBytes = []byte(secret)

	signature := HMACSHA256([]byte(headerPayload), secretBytes)

	// Encode the signature to base64 URL encoding.
	signatureBase64 := Base64URLEncode(signature)

	return fmt.Sprintf("%s.%s", headerPayload, signatureBase64), nil
}

func Base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")

}

// HMACSHA256 signs a message using a secret key with HMAC SHA256.
func HMACSHA256(message, secret []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(message)
	return h.Sum(nil)
}
