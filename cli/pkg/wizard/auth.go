package wizard

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/auth"
)

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
func UserBindTerminus(mnemonic, bflUrl, vaultUrl, authUrl, osPwd, terminusName, localName string) (string, error) {
	log.Printf("Starting userBindTerminus for user: %s", terminusName)

	// 1. Initialize global storage
	if globalUserStore == nil {
		log.Printf("Initializing global stores...")
		err := InitializeGlobalStores(mnemonic, terminusName)
		if err != nil {
			return "", fmt.Errorf("failed to initialize global stores: %w", err)
		}
		log.Printf("Global stores initialized successfully")
	}

	if authUrl != "" {
		globalUserStore.SetAuthURL(authUrl)
		log.Printf("Custom auth URL applied: %s", authUrl)
	}

	// 2. Initialize platform and App (if not already initialized)
	var app *App
	state, err := LoadAppState(globalStorage, globalUserStore.GetDid())
	if err != nil {
		return "", fmt.Errorf("failed to load app state: %w", err)
	}
	if platform == nil {
		log.Printf("Initializing platform...")
		app = NewAppWithState(vaultUrl, state)
		webPlatform := NewWebPlatform(app.API)
		SetPlatform(webPlatform)
		log.Printf("Platform initialized successfully with base URL: %s", vaultUrl)
	} else {
		app = NewAppWithState(vaultUrl, state)
	}

	log.Printf("Using bflUrl: %s", bflUrl)

	// 3. Call /api/firstfactor via the shared pkg/auth implementation.
	//
	//    1:1 mirror of apps/packages/app/src/utils/BindTerminusBusiness.ts
	//    L58-66 (userBindTerminus), which calls onFirstFactor(baseURL,
	//    name, local_name, osPwd, false /*acceptCookie*/, undefined
	//    /*needTwoFactor*/, osVersion) and uses the 1st-factor token
	//    directly without inspecting fa2.
	//
	//    We use auth.FirstFactor (low-level) — NOT auth.Login — because:
	//      - There is no MFA seed yet (it is returned later in
	//        signupResponse.MFA), so even if Authelia echoes fa2=true we
	//        cannot respond to it.
	//      - The first-factor access_token is what the subsequent signup
	//        endpoints need.
	//
	//    NeedTwoFactor=false keeps targetURL = vault.<name>/server (TS
	//    default). AcceptCookie=false matches the explicit `false` arg
	//    in TS L62.
	token, err := auth.FirstFactor(context.TODO(), auth.LoginRequest{
		AuthURL:       bflUrl,
		LocalName:     localName,
		OlaresID:      terminusName,
		Password:      osPwd,
		NeedTwoFactor: false,
		AcceptCookie:  false,
	})
	if err != nil {
		return "", fmt.Errorf("onFirstFactor failed: %v", err)
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
		return "", fmt.Errorf("authentication failed: %v", err)
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
		return "", fmt.Errorf("JWS signing failed: %v", err)
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
		return "", fmt.Errorf("signup failed: %v", err)
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

	return token.AccessToken, nil
}
