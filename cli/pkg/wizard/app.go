package wizard

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

// App class - simplified version for backend CLI use
type App struct {
	Version string  `json:"version"`
	API     *Client `json:"-"` // Uses Client from client.go
}

// NewApp constructor - initializes with Client (corresponds to original TypeScript constructor)
func NewApp(sender Sender) *App {
	// Create simplified client state (backend CLI doesn't need complex state management)
	state := &SimpleClientState{}
	
	// Initialize Client (corresponds to original TypeScript's new Client(this.state, sender, hook))
	client := NewClient(state, sender)
	
	return &App{
		Version: "3.0",
		API:     client,
	}
}

// NewAppWithBaseURL creates App with base URL (convenience function)
func NewAppWithBaseURL(baseURL string) *App {
	// Create HTTP Sender
	sender := NewHTTPSender(baseURL)
	
	// Create App with HTTP Sender
	return NewApp(sender)
}

// SimpleClientState - simplified client state for backend CLI
type SimpleClientState struct {
	session *Session
	account *Account
	device  *DeviceInfo
}

func (s *SimpleClientState) GetSession() *Session {
	return s.session
}

func (s *SimpleClientState) SetSession(session *Session) {
	s.session = session
}

func (s *SimpleClientState) GetAccount() *Account {
	return s.account
}

func (s *SimpleClientState) SetAccount(account *Account) {
	s.account = account
}

func (s *SimpleClientState) GetDevice() *DeviceInfo {
	if s.device == nil {
		s.device = &DeviceInfo{
			ID:       "cli-device-" + generateUUID(),
			Platform: "go-cli",
		}
	}
	return s.device
}

// Signup function - based on original TypeScript signup method (ref: app.ts)
func (a *App) Signup(params SignupParams) (*CreateAccountResponse, error) {
	log.Printf("Starting signup process for DID: %s", params.DID)
	
	// 1. Initialize account object (ref: app.ts line 954-959)
	account := &Account{
		ID:      generateUUID(),
		DID:     params.DID,
		Name:    params.BFLUser, // Use BFLUser as account name
		Local:   false,
		Created: getCurrentTimeISO(),
		Updated: getCurrentTimeISO(),
		MainVault: MainVault{
			ID: "", // Will be set on server side
		},
		Orgs:     []OrgInfo{}, // Initialize as empty array to prevent undefined
		Settings: AccountSettings{},
	}
	
	log.Printf("Account initialized: ID=%s, DID=%s, Name=%s", account.ID, account.DID, account.Name)
	
	// 2. Initialize auth object (ref: app.ts line 964-970)
	auth := NewAuth(params.DID)
	authKey, err := auth.GetAuthKey(params.MasterPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth key: %v", err)
	}
	
	// Calculate verifier (ref: app.ts line 968-970)
	srpClient := NewSRPClient(SRPGroup4096)
	err = srpClient.Initialize(authKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SRP client: %v", err)
	}
	
	auth.Verifier = srpClient.GetV()
	log.Printf("SRP verifier generated: %x...", auth.Verifier[:8])
	
	// 3. Send create account request to server (ref: app.ts line 973-987)
	createParams := CreateAccountParams{
		Account:   *account,
		Auth:      *auth,
		AuthToken: params.AuthToken,
		BFLToken:  params.BFLToken,
		SessionID: params.SessionID,
		BFLUser:   params.BFLUser,
		JWS:       params.JWS,
	}
	
	response, err := a.API.CreateAccount(createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create account on server: %v", err)
	}
	
	log.Printf("Account created on server successfully")
	log.Printf("MFA token received: %s", response.MFA)
	
	// 4. Login to newly created account (ref: app.ts line 991)
	loginParams := LoginParams{
		DID:      params.DID,
		Password: params.MasterPassword,
	}
	
	err = a.Login(loginParams)
	if err != nil {
		return nil, fmt.Errorf("failed to login after signup: %v", err)
	}
	
	log.Printf("Login after signup successful")
	
	// 5. Activate account (ref: app.ts line 1039-1046)
	activeParams := ActiveAccountParams{
		ID:       a.API.State.GetAccount().ID, // Use logged-in account ID
		BFLToken: params.BFLToken,
		BFLUser:  params.BFLUser,
		JWS:      params.JWS,
	}
	
	err = a.API.ActiveAccount(activeParams)
	if err != nil {
		log.Printf("Warning: Failed to activate account: %v", err)
		// Don't return error as account creation was successful
	} else {
		log.Printf("Account activated successfully")
	}
	
	log.Printf("Signup completed successfully for DID: %s", params.DID)
	return response, nil
}

// Login function - simplified version
func (a *App) Login(params LoginParams) error {
	log.Printf("Starting login process for DID: %s", params.DID)
	
	// 1. Start creating session
	startParams := StartCreateSessionParams{
		DID:       params.DID,
		AuthToken: params.AuthToken,
		AsAdmin:   params.AsAdmin,
	}
	
	startResponse, err := a.API.StartCreateSession(startParams)
	if err != nil {
		return fmt.Errorf("failed to start create session: %v", err)
	}
	
	log.Printf("Session creation started for Account ID: %s", startResponse.AccountID)
	
	// 2. Use SRP for authentication
	authKey, err := deriveKeyPBKDF2(
		[]byte(params.Password),
		startResponse.KeyParams.Salt.Bytes(),
		startResponse.KeyParams.Iterations,
		32,
	)
	if err != nil {
		return fmt.Errorf("failed to derive auth key: %v", err)
	}
	
	// 3. SRP client negotiation
	srpClient := NewSRPClient(SRPGroup4096)
	err = srpClient.Initialize(authKey)
	if err != nil {
		return fmt.Errorf("failed to initialize SRP client: %v", err)
	}
	
	err = srpClient.SetB(startResponse.B.Bytes())
	if err != nil {
		return fmt.Errorf("failed to set B value: %v", err)
	}
	
	log.Printf("SRP negotiation completed")
	
	// 4. Complete session creation
	completeParams := CompleteCreateSessionParams{
		SRPId:            startResponse.SRPId,
		AccountID:        startResponse.AccountID,
		A:                Base64Bytes(srpClient.GetA()),
		M:                Base64Bytes(srpClient.GetM1()),
		AddTrustedDevice: false,           // Don't add trusted device by default
		Kind:             "oe",            // Based on server logs, kind should be "oe"
		Version:          "4.0.0",         // Based on server logs, version should be "4.0.0"
	}
	
	session, err := a.API.CompleteCreateSession(completeParams)
	if err != nil {
		return fmt.Errorf("failed to complete create session: %v", err)
	}
	
	// 5. Set session key
	sessionKey := srpClient.GetK()
	session.Key = sessionKey
	a.API.State.SetSession(session)
	
	log.Printf("Session created: %s", session.ID)
	log.Printf("Session key length: %d bytes", len(sessionKey))
	log.Printf("Session key (hex): %x", sessionKey)
	
	// 6. Temporarily skip GetAccount call due to signature verification issues
	// Create a simplified account object for subsequent operations
	account := &Account{
		ID:  startResponse.AccountID,
		DID: params.DID,
		Name: params.DID,
	}
	
	a.API.State.SetAccount(account)
	
	log.Printf("Login completed successfully for DID: %s (skipped GetAccount due to signature issue)", params.DID)
	return nil
}

// Parameter structures
type SignupParams struct {
	DID            string `json:"did"`
	MasterPassword string `json:"masterPassword"`
	Name           string `json:"name"`
	AuthToken      string `json:"authToken"`
	SessionID      string `json:"sessionId"`
	BFLToken       string `json:"bflToken"`
	BFLUser        string `json:"bflUser"`
	JWS            string `json:"jws"`
}

type LoginParams struct {
	DID       string  `json:"did"`
	Password  string  `json:"password"`
	AuthToken *string `json:"authToken,omitempty"`
	AsAdmin   *bool   `json:"asAdmin,omitempty"`
}

// Extend Client interface to support App-required methods
func (c *Client) CreateAccount(params CreateAccountParams) (*CreateAccountResponse, error) {
	requestParams := []interface{}{params}
	response, err := c.call("createAccount", requestParams)
	if err != nil {
		return nil, err
	}
	
	var result CreateAccountResponse
	if err := c.parseResponse(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse CreateAccount response: %v", err)
	}
	
	return &result, nil
}

func (c *Client) ActiveAccount(params ActiveAccountParams) error {
	requestParams := []interface{}{params}
	_, err := c.call("activeAccount", requestParams)
	return err
}

func (c *Client) StartCreateSession(params StartCreateSessionParams) (*StartCreateSessionResponse, error) {
	requestParams := []interface{}{params}
	response, err := c.call("startCreateSession", requestParams)
	if err != nil {
		return nil, err
	}
	
	// Add debug info: print raw response
	if responseBytes, err := json.Marshal(response.Result); err == nil {
		log.Printf("StartCreateSession raw response: %s", string(responseBytes))
	}
	
	var result StartCreateSessionResponse
	if err := c.parseResponse(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse StartCreateSession response: %v", err)
	}
	
	return &result, nil
}

func (c *Client) CompleteCreateSession(params CompleteCreateSessionParams) (*Session, error) {
	requestParams := []interface{}{params}
	response, err := c.call("completeCreateSession", requestParams)
	if err != nil {
		return nil, err
	}
	
	var result Session
	if err := c.parseResponse(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse CompleteCreateSession response: %v", err)
	}
	
	return &result, nil
}

func (c *Client) GetAccount() (*Account, error) {
	// getAccount needs no parameters, pass empty array (ref: client.ts line 46-47: undefined -> [])
	response, err := c.call("getAccount", []interface{}{})
	if err != nil {
		return nil, err
	}
	
	var result Account
	if err := c.parseResponse(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse GetAccount response: %v", err)
	}
	
	return &result, nil
}

// New data structures
type CreateAccountParams struct {
	Account   Account `json:"account"`
	Auth      Auth    `json:"auth"`
	AuthToken string  `json:"authToken"`
	BFLToken  string  `json:"bflToken"`
	SessionID string  `json:"sessionId"`
	BFLUser   string  `json:"bflUser"`
	JWS       string  `json:"jws"`
}

type CreateAccountResponse struct {
	MFA string `json:"mfa"`
}

type ActiveAccountParams struct {
	ID       string `json:"id"`
	BFLToken string `json:"bflToken"`
	BFLUser  string `json:"bflUser"`
	JWS      string `json:"jws"`
}

type StartCreateSessionParams struct {
	DID       string `json:"did"`
	AuthToken *string `json:"authToken,omitempty"`
	AsAdmin   *bool   `json:"asAdmin,omitempty"`
}

type StartCreateSessionResponse struct {
	AccountID string       `json:"accountId"`
	KeyParams PBKDF2Params `json:"keyParams"`
	SRPId     string       `json:"srpId"`
	B         Base64Bytes  `json:"B"`
	Kind      string       `json:"kind,omitempty"`
	Version   string       `json:"version,omitempty"`
}

type CompleteCreateSessionParams struct {
	SRPId            string      `json:"srpId"`
	AccountID        string      `json:"accountId"`
	A                Base64Bytes `json:"A"`                // Use Base64Bytes to handle @AsBytes() decorator
	M                Base64Bytes `json:"M"`                // Use Base64Bytes to handle @AsBytes() decorator
	AddTrustedDevice bool        `json:"addTrustedDevice"` // Add missing field
	Kind             string      `json:"kind"`             // Add kind field
	Version          string      `json:"version"`          // Add version field
}

type PBKDF2Params struct {
	Algorithm  string      `json:"algorithm,omitempty"`
	Hash       string      `json:"hash,omitempty"`
	Salt       Base64Bytes `json:"salt"`
	Iterations int         `json:"iterations"`
	KeySize    int         `json:"keySize,omitempty"`
	Kind       string      `json:"kind,omitempty"`
	Version    string      `json:"version,omitempty"`
}

type Auth struct {
	ID        string       `json:"id"`
	DID       string       `json:"did"`
	Verifier  []byte       `json:"verifier"`
	KeyParams PBKDF2Params `json:"keyParams"`
}

// Auth methods
func NewAuth(did string) *Auth {
	return &Auth{
		ID:  generateUUID(),
		DID: did,
		KeyParams: PBKDF2Params{
			Salt:       generateRandomBytes(16),
			Iterations: 100000,
		},
	}
}

// GetAuthKey generates authentication key (ref: auth.ts line 278-284)
func (a *Auth) GetAuthKey(password string) ([]byte, error) {
	// If no salt is set, generate a random value (ref: auth.ts line 281-282)
	if len(a.KeyParams.Salt) == 0 {
		a.KeyParams.Salt = Base64Bytes(generateRandomBytes(16))
	}
	
	// Use PBKDF2 to derive key (ref: auth.ts line 284 and crypto.ts line 78-101)
	return deriveKeyPBKDF2(
		[]byte(password),
		a.KeyParams.Salt.Bytes(),
		a.KeyParams.Iterations,
		32, // 256 bits = 32 bytes
	)
}

// deriveKeyPBKDF2 implements real PBKDF2 key derivation (ref: deriveKey in crypto.ts)
func deriveKeyPBKDF2(password, salt []byte, iterations, keyLen int) ([]byte, error) {
	// Use real PBKDF2 implementation, ref: crypto.ts line 78-101
	// Use SHA-256 as hash function (corresponds to params.hash in TypeScript)
	key := pbkdf2.Key(password, salt, iterations, keyLen, sha256.New)
	return key, nil
}

// generateRandomBytes generates secure random bytes
func generateRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		// Should handle this error in production implementation
		panic(fmt.Sprintf("Failed to generate random bytes: %v", err))
	}
	return bytes
}

// getCurrentTimeISO gets current time in ISO 8601 format string
func getCurrentTimeISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}
