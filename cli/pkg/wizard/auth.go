package wizard

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Token struct, corresponds to TypeScript Token interface
type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	ExpiresAt    int    `json:"expires_at"`
	SessionID    string `json:"session_id"`
	FA2          bool   `json:"fa2"`
}

// FirstFactorRequest represents first factor request structure
type FirstFactorRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	KeepMeLoggedIn  bool   `json:"keepMeLoggedIn"`
	RequestMethod   string `json:"requestMethod"`
	TargetURL       string `json:"targetURL"`
	AcceptCookie    bool   `json:"acceptCookie"`
}

// FirstFactorResponse represents first factor response structure
type FirstFactorResponse struct {
	Status string `json:"status"`
	Data   Token  `json:"data"`
}

// OnFirstFactor implements first factor authentication (ref: BindTerminusBusiness.ts)
func OnFirstFactor(baseURL, terminusName, osUser, osPwd string, acceptCookie, needTwoFactor bool) (*Token, error) {
	log.Printf("Starting onFirstFactor for user: %s", osUser)
	
	// Process password (salted MD5)
	processedPassword := passwordAddSort(osPwd)
	
	// Build request
	reqData := FirstFactorRequest{
		Username:       osUser,
		Password:       processedPassword,
		KeepMeLoggedIn: false,
		RequestMethod:  "POST",
		TargetURL:      baseURL,
		AcceptCookie:   acceptCookie,
	}
	
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}
	
	// Send HTTP request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	reqURL := fmt.Sprintf("%s/api/firstfactor?hideCookie=true", baseURL)
	req, err := http.NewRequest("POST", reqURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	log.Printf("Sending request to: %s", reqURL)
	
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
	
	var response FirstFactorResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	
	if response.Status != "OK" {
		return nil, fmt.Errorf("authentication failed: %s", response.Status)
	}
	
	log.Printf("First factor authentication successful")
	return &response.Data, nil
}

// passwordAddSort implements salted MD5 (ref: TypeScript version)
func passwordAddSort(password string) string {
	// Salt and MD5
	saltedPassword := password + "@Olares2025"
	hash := md5.Sum([]byte(saltedPassword))
	return fmt.Sprintf("%x", hash)
}


// Main authentication function - corresponds to original TypeScript _authenticate function
func Authenticate(req AuthenticateRequest) (*AuthenticateResponse, error) {
	if platform == nil {
		return nil, NewAuthError(ErrorCodeServerError, "Platform not initialized", nil)
	}

	step := 1
	var authReq *StartAuthRequestResponse = req.PendingRequest

	// Step 1: If no pending request, start new authentication request
	if authReq == nil {
		log.Printf("[%s] Step %d: req is empty, starting auth request...", req.Caller, step)
		
		opts := StartAuthRequestOptions{
			Type:               &req.Type,
			Purpose:            req.Purpose,
			DID:                &req.DID,
			AuthenticatorIndex: &req.AuthenticatorIndex,
		}

		var err error
		authReq, err = platform.StartAuthRequest(opts)
		if err != nil {
			log.Printf("[%s] Step %d: Error occurred while starting auth request: %v", req.Caller, step, err)
			return nil, NewAuthError(
				ErrorCodeAuthenticationFailed,
				fmt.Sprintf("[%s] Step %d: An error occurred: %s", req.Caller, step, err.Error()),
				map[string]any{"error": err},
			)
		}

		reqJSON, _ := json.Marshal(authReq)
		log.Printf("[%s] Step %d: Auth request started successfully. Request details: %s", req.Caller, step, string(reqJSON))
	} else {
		log.Printf("[%s] Step %d: req already exists. Skipping auth request.", req.Caller, step)
	}

	// Step 2: Complete authentication request
	step = 2
	reqJSON, _ := json.Marshal(authReq)
	log.Printf("[%s] Step %d: Completing auth request with req: %s", req.Caller, step, string(reqJSON))

	res, err := platform.CompleteAuthRequest(authReq)
	if err != nil {
		log.Printf("[%s] Step %d: Error occurred while completing auth request: %v", req.Caller, step, err)
		return nil, NewAuthError(
			ErrorCodeAuthenticationFailed,
			fmt.Sprintf("[%s] Step %d: An error occurred: %s", req.Caller, step, err.Error()),
			map[string]any{"error": err},
		)
	}

	resJSON, _ := json.Marshal(res)
	log.Printf("[%s] Step %d: Auth request completed successfully. Response details: %s", req.Caller, step, string(resJSON))

	return res, nil
}


// UserBindTerminus main user binding function (ref: TypeScript version)
func UserBindTerminus(mnemonic, bflUrl, vaultUrl, osPwd, terminusName, localName string) error {
	log.Printf("Starting userBindTerminus for user: %s", terminusName)
	
	// 1. Initialize global storage
	if globalUserStore == nil {
		log.Printf("Initializing global stores...")
		err := InitializeGlobalStores(mnemonic, terminusName)
		if err != nil {
			return fmt.Errorf("failed to initialize global stores: %w", err)
		}
		log.Printf("Global stores initialized successfully")
	}
	
	// 2. Initialize platform and App (if not already initialized)
	var app *App
	if platform == nil {
		log.Printf("Initializing platform...")
		
		// Create App using vaultUrl as base URL
		app = NewAppWithBaseURL(vaultUrl)
		
		// Create and set WebPlatform (no need to pass mnemonic, uses global storage)
		webPlatform := NewWebPlatform(app.API)
		SetPlatform(webPlatform)
		
		log.Printf("Platform initialized successfully with base URL: %s", vaultUrl)
	} else {
		// If platform already initialized, create new App instance for signup
		app = NewAppWithBaseURL(vaultUrl)
	}
	
	log.Printf("Using bflUrl: %s", bflUrl)
	
	// 3. Call onFirstFactor to get token (ref: TypeScript implementation)
	token, err := OnFirstFactor(bflUrl, terminusName, localName, osPwd, false, false)
	if err != nil {
		return fmt.Errorf("onFirstFactor failed: %v", err)
	}
	
	log.Printf("First factor authentication successful, session_id: %s", token.SessionID)
	
	// 4. Execute authentication - call _authenticate function from pkg/activate
	authRes, err := Authenticate(AuthenticateRequest{
		DID:                localName,
		Type:               AuthTypeSSI,
		Purpose:            AuthPurposeSignup,
		AuthenticatorIndex: 0,
		Caller:             "E001",
	})
	if err != nil {
		return fmt.Errorf("authentication failed: %v", err)
	}
	
	log.Printf("Authentication successful for DID: %s", authRes.DID)
	
	// 5. Generate JWS - ref: BindTerminusBusiness.ts
	log.Printf("Creating JWS for signup...")
	
	// Extract domain (ref: TypeScript implementation)
	domain := vaultUrl
	if strings.HasPrefix(domain, "http://") {
		domain = domain[7:]
	} else if strings.HasPrefix(domain, "https://") {
		domain = domain[8:]
	}
	
	// Use globalUserStore to sign JWS (ref: userStore.signJWS in TypeScript)
	jws, err := globalUserStore.SignJWS(map[string]any{
		"name":   terminusName,
		"did":    globalUserStore.GetDid(),
		"domain": domain,
		"time":   fmt.Sprintf("%d", time.Now().UnixMilli()),
	})
	if err != nil {
		return fmt.Errorf("JWS signing failed: %v", err)
	}
	
	log.Printf("JWS created successfully: %s...", jws[:50])
	
	// 6. Execute signup (call real implementation in app.go)
	log.Printf("Executing signup...")
	
	// Build SignupParams (ref: app.signup in BindTerminusBusiness.ts)
	signupParams := SignupParams{
		DID:            authRes.DID,
		MasterPassword: mnemonic,
		Name:           terminusName,
		AuthToken:      authRes.Token,
		SessionID:      token.SessionID,
		BFLToken:       token.AccessToken,
		BFLUser:        localName,
		JWS:            jws,
	}
	
	// Call real app.Signup function
	signupResponse, err := app.Signup(signupParams)
	if err != nil {
		return fmt.Errorf("signup failed: %v", err)
	}
	
	log.Printf("Signup successful! MFA: %s", signupResponse.MFA)
	
	// Save MFA token to UserStore for next stage use
	err = globalUserStore.SetMFA(signupResponse.MFA)
	if err != nil {
		log.Printf("Warning: Failed to save MFA token: %v", err)
		// Don't return error as main process has succeeded
	} else {
		log.Printf("MFA token saved to UserStore for future use")
	}
	
	log.Printf("User bind to Terminus completed successfully!")
	
	return nil
}
