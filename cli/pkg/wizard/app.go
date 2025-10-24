package wizard

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
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
		Version:  "3.0.14",
	}

	// Initialize account with master password (ref: account.ts line 182-190)
	err := a.initializeAccount(account, params.MasterPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize account: %v", err)
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

	// 5. Initialize main vault and create TOTP item (ref: app.ts line 1003-1038)
	err = a.initializeMainVaultWithTOTP(response.MFA)
	if err != nil {
		log.Printf("Warning: Failed to initialize main vault with TOTP: %v", err)
		// Don't return error as account creation was successful
	} else {
		log.Printf("Main vault initialized with TOTP item successfully")
	}

	// 6. Activate account (ref: app.ts line 1039-1046)
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
		AddTrustedDevice: false,   // Don't add trusted device by default
		Kind:             "oe",    // Based on server logs, kind should be "oe"
		Version:          "4.0.0", // Based on server logs, version should be "4.0.0"
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

	// Create a simplified account object for subsequent operations
	account, err := a.API.GetAccount()
	if err != nil {
		return fmt.Errorf("failed to get account: %v", err)
	}

	// account := &Account{
	// 	ID:   startResponse.AccountID,
	// 	DID:  params.DID,
	// 	Name: params.DID,
	// }

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

func (c *Client) UpdateVault(vault Vault) (*Vault, error) {
	requestParams := []interface{}{vault}
	response, err := c.call("updateVault", requestParams)
	if err != nil {
		return nil, err
	}

	var result Vault
	if err := c.parseResponse(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse UpdateVault response: %v", err)
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
	DID       string  `json:"did"`
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

// initializeMainVaultWithTOTP initializes main vault and creates TOTP item (ref: app.ts line 1003-1038)
func (a *App) initializeMainVaultWithTOTP(mfaToken string) error {
	account := a.API.State.GetAccount()
	if account == nil {
		return fmt.Errorf("account is null")
	}

	// 1. Initialize main vault (ref: server.ts line 1573-1579)
	vault := &Vault{
		ID:      account.MainVault.ID, // Use existing vault ID from account
		Name:    "My Vault",
		Owner:   account.ID,
		Created: getCurrentTimeISO(),
		Updated: getCurrentTimeISO(),
		Items:   []VaultItem{}, // Initialize empty items array
	}

	log.Printf("Main vault initialized: ID=%s, Name=%s, Owner=%s", vault.ID, vault.Name, vault.Owner)

	// 2. Get authenticator template (ref: app.ts line 1008-1014)
	template := GetAuthenticatorTemplate()
	if template == nil {
		return fmt.Errorf("authenticator template is null")
	}

	// 3. Set MFA token value (ref: app.ts line 1015)
	template.Fields[0].Value = mfaToken
	log.Printf("TOTP template prepared with MFA token: %s...", mfaToken[:min(8, len(mfaToken))])

	// 4. Create vault item (ref: app.ts line 1024-1033)
	item, err := a.createVaultItem(CreateVaultItemParams{
		Name:   account.Name,
		Vault:  vault,
		Fields: template.Fields,
		Tags:   []string{},
		Icon:   template.Icon,
		Type:   VaultTypeTerminusTotp,
	})
	if err != nil {
		return fmt.Errorf("failed to create vault item: %v", err)
	}

	log.Printf("TOTP vault item created: ID=%s, Name=%s", item.ID, item.Name)
	log.Printf("TOTP field value: %s", item.Fields[0].Value)

	// 5. Add item to vault
	vault.Items = append(vault.Items, *item)

	// 6. Update vault on server (ref: app.ts line 2138: await this.addItems([item], vault))
	err = a.updateVault(vault)
	if err != nil {
		return fmt.Errorf("failed to update vault on server: %v", err)
	}

	log.Printf("Vault updated on server successfully")
	return nil
}

// CreateVaultItemParams parameters for creating a vault item
type CreateVaultItemParams struct {
	Name   string
	Vault  *Vault
	Fields []Field
	Tags   []string
	Icon   string
	Type   VaultType
}

// createVaultItem creates a new vault item (ref: app.ts line 2096-2141)
func (a *App) createVaultItem(params CreateVaultItemParams) (*VaultItem, error) {
	account := a.API.State.GetAccount()
	if account == nil {
		return nil, fmt.Errorf("account is null")
	}

	// Create vault item (ref: item.ts line 451-475)
	item := &VaultItem{
		ID:        generateUUID(),
		Name:      params.Name,
		Type:      params.Type,
		Icon:      params.Icon,
		Fields:    params.Fields,
		Tags:      params.Tags,
		Updated:   getCurrentTimeISO(),
		UpdatedBy: account.ID,
	}

	log.Printf("Vault item created: ID=%s, Name=%s, Type=%d", item.ID, item.Name, item.Type)
	return item, nil
}

// updateVault updates vault on server (ref: app.ts line 1855-2037)
func (a *App) updateVault(vault *Vault) error {
	// Update vault revision
	vault.Revision = generateUUID()
	vault.Updated = getCurrentTimeISO()

	// Call server API to update vault
	updatedVault, err := a.API.UpdateVault(*vault)
	if err != nil {
		return fmt.Errorf("failed to update vault on server: %v", err)
	}

	log.Printf("Vault updated on server: ID=%s, Revision=%s", updatedVault.ID, updatedVault.Revision)
	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// initializeAccount initializes account with RSA keys and encryption parameters (ref: account.ts line 182-190)
func (a *App) initializeAccount(account *Account, masterPassword string) error {
	// 1. Generate RSA key pair (ref: account.ts line 183-186)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate RSA key pair: %v", err)
	}

	// 2. Extract public key and encode it (ref: account.ts line 186)
	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %v", err)
	}
	account.PublicKey = base64.StdEncoding.EncodeToString(publicKeyDER)

	// 3. Set up key derivation parameters (ref: container.ts line 125-133)
	salt := generateRandomBytes(16)
	account.KeyParams = KeyParams{
		Algorithm:  "PBKDF2",
		Hash:       "SHA-256",
		KeySize:    256,
		Iterations: 100000,
		Salt:       base64.StdEncoding.EncodeToString(salt),
		Version:    "3.0.14",
	}

	// 4. Derive encryption key from master password
	encryptionKey := pbkdf2.Key([]byte(masterPassword), salt, account.KeyParams.Iterations, 32, sha256.New)

	// 5. Set up encryption parameters (ref: container.ts line 48-56)
	iv := generateRandomBytes(16)
	additionalData := generateRandomBytes(16)
	account.EncryptionParams = EncryptionParams{
		Algorithm:      "AES-GCM",
		TagSize:        128,
		KeySize:        256,
		IV:             base64.StdEncoding.EncodeToString(iv),
		AdditionalData: base64.StdEncoding.EncodeToString(additionalData),
		Version:        "3.0.14",
	}

	// 6. Create account secrets (private key + signing key)
	privateKeyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	signingKey := generateRandomBytes(32) // HMAC key

	// Combine private key and signing key into account secrets
	accountSecrets := struct {
		SigningKey []byte `json:"signingKey"`
		PrivateKey []byte `json:"privateKey"`
	}{
		SigningKey: signingKey,
		PrivateKey: privateKeyDER,
	}

	accountSecretsBytes, err := json.Marshal(accountSecrets)
	if err != nil {
		return fmt.Errorf("failed to marshal account secrets: %v", err)
	}

	// 7. Encrypt account secrets (ref: container.ts line 59-63)
	encryptedData, err := a.encryptAESGCM(encryptionKey, accountSecretsBytes, iv, additionalData)
	if err != nil {
		return fmt.Errorf("failed to encrypt account secrets: %v", err)
	}
	account.EncryptedData = base64.StdEncoding.EncodeToString(encryptedData)

	log.Printf("Account initialized with RSA key pair and encryption parameters")
	log.Printf("Public key length: %d bytes", len(publicKeyDER))
	log.Printf("Encrypted data length: %d bytes", len(encryptedData))

	return nil
}

// encryptAESGCM encrypts data using AES-GCM
func (a *App) encryptAESGCM(key, plaintext, iv, additionalData []byte) ([]byte, error) {
	// Import crypto/aes and crypto/cipher packages are needed at the top of the file
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, 16)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}

	// Encrypt the plaintext using AES-GCM
	ciphertext := gcm.Seal(nil, iv, plaintext, additionalData)

	return ciphertext, nil
}
