package wizard

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
)

// LoginTerminus implements Terminus login functionality (ref: BindTerminusBusiness.ts loginTerminus)
func LoginTerminus(bflUrl, terminusName, localName, password string, needTwoFactor bool) (*Token, error) {
	log.Printf("Starting loginTerminus for user: %s", terminusName)
	
	// 1. Call onFirstFactor to get initial token (ref: loginTerminus line 364-372)
	token, err := OnFirstFactor(bflUrl, terminusName, localName, password, true, needTwoFactor)
	if err != nil {
		return nil, fmt.Errorf("first factor authentication failed: %v", err)
	}
	
	log.Printf("First factor completed, session_id: %s, FA2 required: %t", token.SessionID, token.FA2 || needTwoFactor)
	
	// 2. If second factor authentication is required (ref: loginTerminus line 379-446)
	if token.FA2 || needTwoFactor {
		log.Printf("Second factor authentication required")
		
		// Get TOTP value
		totpValue, err := getTOTPFromMFA()
		if err != nil {
			return nil, fmt.Errorf("failed to get TOTP: %v", err)
		}
		
		log.Printf("Generated TOTP: %s", totpValue)
		
		// Perform second factor authentication
		secondToken, err := performSecondFactor(bflUrl, terminusName, totpValue)
		if err != nil {
			return nil, fmt.Errorf("second factor authentication failed: %v", err)
		}
		
		// Update token information
		token.AccessToken = secondToken.AccessToken
		token.RefreshToken = secondToken.RefreshToken
		token.SessionID = secondToken.SessionID
		
		log.Printf("Second factor completed, updated session_id: %s", token.SessionID)
	}
	
	log.Printf("LoginTerminus completed successfully")
	return token, nil
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

// performSecondFactor performs second factor authentication (ref: loginTerminus line 419-446)
func performSecondFactor(baseURL, terminusName, totpValue string) (*Token, error) {
	log.Printf("Performing second factor authentication")
	
	// Build target URL
	targetURL := fmt.Sprintf("https://desktop.%s/", strings.ReplaceAll(terminusName, "@", "."))
	
	// Build request data
	reqData := map[string]interface{}{
		"targetUrl": targetURL,
		"token":     totpValue,
	}
	
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}
	
	// Send HTTP request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	url := fmt.Sprintf("%s/api/secondfactor/totp", baseURL)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Access-Control-Allow-Origin", "*")
	req.Header.Set("X-Unauth-Error", "Non-Redirect")
	
	log.Printf("Sending second factor request to: %s", url)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}
	
	var response struct {
		Status string `json:"status"`
		Data   Token  `json:"data"`
	}
	
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	
	if response.Status != "OK" {
		return nil, fmt.Errorf("second factor authentication failed: %s", response.Status)
	}
	
	log.Printf("Second factor authentication successful")
	return &response.Data, nil
}

// ResetPassword implements password reset functionality (ref: account.ts reset_password)
func ResetPassword(baseURL, localName, currentPassword, newPassword, accessToken string) error {
	log.Printf("Starting reset password for user: %s", localName)
	
	// Process passwords (salted MD5)
	processedCurrentPassword := passwordAddSort(currentPassword)
	processedNewPassword := passwordAddSort(newPassword)
	
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
		Code    int    `json:"code"`
		Message string `json:"message"`
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
