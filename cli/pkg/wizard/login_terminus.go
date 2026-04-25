package wizard

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/auth"
)

// LoginTerminus performs first-factor (and, when needed, second-factor TOTP)
// authentication against the Authelia backend. The actual HTTP work is
// delegated to pkg/auth.Login so the wizard never owns its own copy of the
// `passwordAddSort` salt math, the cookie-jar / 2FA wiring, or the response
// parser — keeping wire-format quirks centralised in pkg/auth.
//
// The wizard-specific bit that pkg/auth deliberately does not know about is
// where the TOTP code comes from: during activation it has to be computed
// locally from the MFA seed stored in globalUserStore (see getTOTPFromMFA).
// We therefore:
//
//  1. pre-compute TOTP eagerly when the caller already knows 2FA is on
//     (`needTwoFactor=true`), so we can submit both factors in one call;
//  2. fall back to the same TOTP source if pkg/auth.Login surfaces
//     ErrTOTPRequired (caller passed false but server says fa2 is needed) —
//     this matches the old wizard behaviour of branching on
//     `token.FA2 || needTwoFactor`.
func LoginTerminus(bflUrl, terminusName, localName, password string, needTwoFactor bool) (*auth.Token, error) {
	log.Printf("Starting loginTerminus for user: %s", terminusName)

	// 1:1 mirror of apps/packages/app/src/utils/BindTerminusBusiness.ts
	// L364-372 (loginTerminus): onFirstFactor is invoked with
	// `acceptCookie=true, needTwoFactor=<arg>`. NeedTwoFactor here flips
	// the targetURL onto desktop.<name>/ so Authelia's 2FA-policy fires
	// (matching TS L21-25 in account.ts) and is also OR'd with tok.FA2
	// inside auth.Login to gate the second-factor POST (TS L379).
	req := auth.LoginRequest{
		AuthURL:       bflUrl,
		LocalName:     localName,
		TerminusName:  terminusName,
		Password:      password,
		NeedTwoFactor: needTwoFactor,
		AcceptCookie:  true,
	}
	if needTwoFactor {
		totp, err := getTOTPFromMFA()
		if err != nil {
			return nil, fmt.Errorf("get totp: %w", err)
		}
		log.Printf("Generated TOTP (eager, needTwoFactor=true)")
		req.TOTP = totp
	}

	tok, err := auth.Login(context.TODO(), req)
	if errors.Is(err, auth.ErrTOTPRequired) {
		// Caller asserted no 2FA but the server disagreed. Pull the TOTP
		// from the MFA seed and retry once.
		log.Printf("Server reported fa2 even though caller passed needTwoFactor=false; retrying with TOTP")
		totp, ferr := getTOTPFromMFA()
		if ferr != nil {
			return nil, fmt.Errorf("get totp: %w", ferr)
		}
		req.TOTP = totp
		tok, err = auth.Login(context.TODO(), req)
	}
	if err != nil {
		return nil, err
	}

	log.Printf("LoginTerminus completed successfully, session_id: %s", tok.SessionID)
	return tok, nil
}

// getTOTPFromMFA generates TOTP from stored MFA (ref: loginTerminus line 380-403)
func getTOTPFromMFA() (string, error) {
	// Get MFA token from global storage
	mfa, err := globalUserStore.GetMFA()
	if err != nil {
		return "", fmt.Errorf("MFA token not found: %v", err)
	}

	log.Printf("Using MFA token for TOTP generation: %s", mfa)

	// Generate TOTP (ref: TypeScript hotp function)
	currentTime := time.Now().Unix()
	interval := int64(30) // 30 second interval
	counter := currentTime / interval

	totp, err := generateHOTP(mfa, counter)
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP: %v", err)
	}

	return totp, nil
}

// generateHOTP generates HOTP (ref: TypeScript hotp function)
func generateHOTP(secret string, counter int64) (string, error) {
	// Process base32 string: remove spaces, convert to uppercase, handle padding
	cleanSecret := strings.ToUpper(strings.ReplaceAll(secret, " ", ""))

	// Add padding characters if needed
	padding := len(cleanSecret) % 8
	if padding != 0 {
		cleanSecret += strings.Repeat("=", 8-padding)
	}

	// Decode base32 encoded secret to bytes
	secretBytes, err := base32.StdEncoding.DecodeString(cleanSecret)
	if err != nil {
		return "", fmt.Errorf("failed to decode base32 secret: %v", err)
	}

	// Convert counter to 8-byte big-endian
	counterBytes := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		counterBytes[i] = byte(counter & 0xff)
		counter >>= 8
	}

	// Use HMAC-SHA1 to calculate hash (consistent with TypeScript version)
	h := hmac.New(sha1.New, secretBytes)
	h.Write(counterBytes)
	hash := h.Sum(nil)

	// Dynamic truncation (consistent with TypeScript getToken function)
	offset := hash[len(hash)-1] & 0xf
	code := ((int(hash[offset]) & 0x7f) << 24) |
		((int(hash[offset+1]) & 0xff) << 16) |
		((int(hash[offset+2]) & 0xff) << 8) |
		(int(hash[offset+3]) & 0xff)

	// Generate 6-digit number
	otp := code % int(math.Pow10(6))

	return fmt.Sprintf("%06d", otp), nil
}

// ResetPassword implements password reset functionality (ref: account.ts reset_password)
func ResetPassword(baseURL, localName, currentPassword, newPassword, accessToken string) error {
	log.Printf("Starting reset password for user: %s", localName)

	// Process passwords (salted MD5) — reuse pkg/auth so wizard never owns
	// its own copy of the salt; see auth.PasswordSalt for rationale.
	processedCurrentPassword := auth.PasswordSalt(currentPassword)
	processedNewPassword := auth.PasswordSalt(newPassword)

	// Build request data (ref: account.ts line 138-141)
	reqData := map[string]interface{}{
		"current_password": processedCurrentPassword,
		"password":         processedNewPassword,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP client (ref: account.ts line 128-135)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Build request URL (ref: account.ts line 136-137)
	url := fmt.Sprintf("%s/bfl/iam/v1alpha1/users/%s/password", baseURL, localName)
	req, err := http.NewRequest("PUT", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set request headers (ref: account.ts line 131-134)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Authorization", accessToken)

	log.Printf("Sending reset password request to: %s", url)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// Check HTTP status code (ref: account.ts line 144-146)
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// Check response status (ref: account.ts line 148-155)
	if response.Code != 0 {
		if response.Message != "" {
			return fmt.Errorf("password reset failed: %s", response.Message)
		}
		return fmt.Errorf("password reset failed: network error")
	}

	log.Printf("Password reset completed successfully")
	return nil
}
