package wizard

import (
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

// App class - mirrors the TS App in apps/packages/sdk/src/core/app.ts.
//
// Holds a single AppState which is both the in-memory client state used
// by the RPC Client and the persistable representation of the
// account / vaults / orgs known to this client.
type App struct {
	Version string    `json:"version"`
	API     *Client   `json:"-"`
	State   *AppState `json:"-"`
}

// NewApp constructs an App using the given sender and AppState.
// If state is nil, an in-memory only state is created.
func NewApp(sender Sender, state *AppState) *App {
	if state == nil {
		state = NewAppState(nil, "")
	}
	client := NewClient(state, sender)
	return &App{
		Version: "3.0",
		API:     client,
		State:   state,
	}
}

// NewAppWithBaseURL creates App with base URL (convenience function).
// Uses an in-memory state. Prefer NewAppWithState when you need persistence.
func NewAppWithBaseURL(baseURL string) *App {
	sender := NewHTTPSender(baseURL)
	return NewApp(sender, nil)
}

// NewAppWithState creates App with an explicit AppState (typically backed
// by a DirKVStorage rooted at ~/.olares/<did>/).
func NewAppWithState(baseURL string, state *AppState) *App {
	sender := NewHTTPSender(baseURL)
	return NewApp(sender, state)
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

	// 5. Inject MFA TOTP item into the freshly synchronized main vault
	//    (ref: apps/packages/sdk/src/core/app.ts:1003-1038).
	if mv := a.MainVault(); mv != nil && response.MFA != "" {
		log.Printf("Signup: TOTP injection target mv id=%s revision=%s accessors=%d encryptedData=%d aesKey?=%t items?=%t itemCount=%d",
			mv.ID, mv.Revision, len(mv.Accessors), len(mv.EncryptedData),
			mv.aesKey != nil, mv.items != nil, vaultItemCount(mv))

		tpl := GetAuthenticatorTemplate()
		if tpl != nil && len(tpl.Fields) > 0 {
			tpl.Fields[0].Value = response.MFA
			if _, err := a.CreateItem(CreateItemParams{
				Name:   account.Name,
				Vault:  mv,
				Fields: tpl.Fields,
				Tags:   []string{},
				Icon:   tpl.Icon,
				Type:   VaultTypeTerminusTotp,
			}); err != nil {
				log.Printf("Warning: failed to write TOTP item to main vault: %v", err)
				log.Printf("CRITICAL: TOTP item not stored in vault, web second-factor will fail: %v", err)
			} else {
				log.Printf("Main vault updated with TOTP item")

				// Round-trip verification: refetch the vault from server,
				// unlock with the same account, and list the items the
				// server now persists. Diagnostic-only — useful for
				// distinguishing "push succeeded" from "push succeeded
				// but server did not actually persist TOTP".
				if unlocked := a.State.Unlocked(); unlocked != nil {
					verify, gerr := a.API.GetVault(mv.ID)
					if gerr != nil {
						log.Printf("verify: GetVault after TOTP push failed: %v", gerr)
					} else {
						if uerr := verify.Unlock(unlocked); uerr != nil {
							log.Printf("verify: Unlock of refetched vault failed: %v", uerr)
						}
						count := vaultItemCount(verify)
						log.Printf("verify: server vault id=%s revision=%s items=%d accessors=%d encryptedData=%d",
							verify.ID, verify.Revision, count,
							len(verify.Accessors), len(verify.EncryptedData))
						if verify.items != nil {
							for _, it := range verify.items.Items {
								firstFieldType := ""
								if len(it.Fields) > 0 {
									firstFieldType = string(it.Fields[0].Type)
								}
								log.Printf("verify item: id=%s type=%d name=%s fields=%d firstFieldType=%s",
									it.ID, it.Type, it.Name, len(it.Fields), firstFieldType)
							}
						}

						// Vault id alignment: compare the vault id we just
						// wrote into with the one the server records as
						// account.mainVault.id. If they diverge, the web
						// client (which reads account.mainVault.id from
						// server) will fetch a different vault and never
						// see the TOTP item.
						srvAcc, aerr := a.API.GetAccount()
						if aerr != nil {
							log.Printf("post-signup: GetAccount failed: %v", aerr)
						} else if srvAcc == nil {
							log.Printf("post-signup: GetAccount returned nil account")
						} else {
							match := srvAcc.MainVault.ID == mv.ID
							log.Printf("post-signup: account.id=%s account.did=%s account.mainVault.id=%s vs CLI mv.id=%s match=%t",
								srvAcc.ID, srvAcc.DID, srvAcc.MainVault.ID, mv.ID, match)
							for i, ac := range verify.Accessors {
								accMatch := ac.ID == srvAcc.ID
								log.Printf("post-signup: verify.accessors[%d].id=%s vs account.id=%s match=%t",
									i, ac.ID, srvAcc.ID, accMatch)
							}
						}
					}
				} else {
					log.Printf("verify: skipped (no unlocked account in state)")
				}
			}
		}
	} else if response.MFA != "" {
		log.Printf("Warning: skipped TOTP injection (no main vault available yet)")
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

// Login mirrors apps/packages/sdk/src/core/app.ts App.login.
//
// Flow:
//  1. SRP negotiate session
//  2. GetAccount → Account.Unlock(password) → AppState.SetUnlocked
//  3. Persist app state to disk (Save)
//  4. Synchronize (AuthInfo + Account + Orgs + Vaults)
//  5. If a localvault existed before login (from a prior session), merge
//     its items into the (possibly new) main vault.
func (a *App) Login(params LoginParams) error {
	log.Printf("Starting login process for DID: %s", params.DID)

	// 1. SRP — start session
	startResponse, err := a.API.StartCreateSession(StartCreateSessionParams{
		DID:       params.DID,
		AuthToken: params.AuthToken,
		AsAdmin:   params.AsAdmin,
	})
	if err != nil {
		return fmt.Errorf("failed to start create session: %v", err)
	}
	log.Printf("Session creation started for Account ID: %s", startResponse.AccountID)

	authKey, err := deriveKeyPBKDF2(
		[]byte(params.Password),
		startResponse.KeyParams.Salt.Bytes(),
		startResponse.KeyParams.Iterations,
		32,
	)
	if err != nil {
		return fmt.Errorf("failed to derive auth key: %v", err)
	}
	srpClient := NewSRPClient(SRPGroup4096)
	if err := srpClient.Initialize(authKey); err != nil {
		return fmt.Errorf("failed to initialize SRP client: %v", err)
	}
	if err := srpClient.SetB(startResponse.B.Bytes()); err != nil {
		return fmt.Errorf("failed to set B value: %v", err)
	}

	session, err := a.API.CompleteCreateSession(CompleteCreateSessionParams{
		SRPId:            startResponse.SRPId,
		AccountID:        startResponse.AccountID,
		A:                Base64Bytes(srpClient.GetA()),
		M:                Base64Bytes(srpClient.GetM1()),
		AddTrustedDevice: false,
		Kind:             "oe",
		Version:          "4.0.0",
	})
	if err != nil {
		return fmt.Errorf("failed to complete create session: %v", err)
	}
	session.Key = srpClient.GetK()
	a.API.State.SetSession(session)
	log.Printf("Session created: %s", session.ID)

	// 2. Fetch & unlock the server account.
	account, err := a.API.GetAccount()
	if err != nil {
		return fmt.Errorf("failed to fetch account after login: %v", err)
	}
	unlocked, err := account.Unlock(params.Password)
	if err != nil {
		return fmt.Errorf("failed to unlock account: %v", err)
	}

	a.API.State.SetAccount(account)
	if a.State != nil {
		a.State.SetUnlocked(unlocked)
	}

	// 3. Snapshot any pre-existing local main vault so we can merge stale
	//    items in step 5.
	var localvault *Vault
	if a.State != nil && len(a.State.Vaults) > 0 {
		v := a.State.Vaults[0]
		localvault = &v
	}

	// 4. Persist what we have so far, then synchronize.
	if a.State != nil {
		if err := a.State.Save(); err != nil {
			log.Printf("Login: failed to persist app state: %v", err)
		}
	}
	if err := a.Synchronize(); err != nil {
		log.Printf("Login: synchronize failed: %v", err)
		// Synchronize errors are non-fatal — the user can still operate
		// against the local cache.
	}

	// 5. Merge the legacy localvault into the (possibly new) main vault
	//    by re-creating any items that no longer exist remotely.
	if localvault != nil && a.State != nil {
		if mv := a.MainVault(); mv != nil && mv.ID != localvault.ID {
			if err := localvault.Unlock(unlocked); err != nil {
				log.Printf("Login: failed to unlock localvault for merge: %v", err)
			} else if mv.aesKey == nil {
				if err := mv.Unlock(unlocked); err != nil {
					log.Printf("Login: failed to unlock new mainVault for merge: %v", err)
				}
			}
			// Only drop the legacy localvault once we are sure every item
			// has been safely migrated into the new mainVault. If any
			// unlock or CreateItem call failed we keep the localvault
			// around so the next Login attempt can retry the merge.
			mergeReady := mv.items != nil && localvault.items != nil
			migrationFailed := false
			if mergeReady {
				for id, item := range localvault.items.Items {
					if _, exists := mv.items.Items[id]; exists {
						continue
					}
					if item.Name == "" {
						continue
					}
					itemType := item.Type
					if itemType == VaultTypeTerminusTotp {
						itemType = VaultTypeDefault
					}
					if _, err := a.CreateItem(CreateItemParams{
						ID:     id,
						Name:   item.Name,
						Vault:  mv,
						Fields: item.Fields,
						Tags:   item.Tags,
						Icon:   item.Icon,
						Type:   itemType,
					}); err != nil {
						log.Printf("Login: failed to migrate localvault item %s: %v", id, err)
						migrationFailed = true
					}
				}
			} else {
				log.Printf("Login: skipping localvault merge (mv.items=%v, localvault.items=%v); keeping localvault for retry",
					mv.items != nil, localvault.items != nil)
			}
			if mergeReady && !migrationFailed {
				a.State.RemoveVault(localvault.ID)
				if err := a.State.Save(); err != nil {
					log.Printf("Login: failed to persist app state after merge: %v", err)
				}
			}
		}
	}

	log.Printf("Login completed successfully for DID: %s", params.DID)
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
	if vault.Revision == "" {
		if acc := c.State.GetAccount(); acc != nil && vault.ID != "" &&
			acc.MainVault.ID == vault.ID && acc.MainVault.Revision != "" {
			vault.Revision = acc.MainVault.Revision
		}
	}
	if vault.Revision == "" && vault.ID != "" {
		current, err := c.GetVault(vault.ID)
		if err != nil {
			return nil, fmt.Errorf("updateVault: resolve revision: %w", err)
		}
		vault.Revision = current.Revision
		if acc := c.State.GetAccount(); acc != nil && acc.MainVault.ID == vault.ID && current.Revision != "" {
			acc.MainVault.Revision = current.Revision
		}
	}
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

// GetVault fetches a vault by id (mirrors api.getVault in TS).
func (c *Client) GetVault(id string) (*Vault, error) {
	response, err := c.call("getVault", []interface{}{id})
	if err != nil {
		return nil, err
	}
	var result Vault
	if err := c.parseResponse(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse GetVault response: %v", err)
	}
	return &result, nil
}

// GetOrg fetches an org by id (mirrors api.getOrg in TS).
func (c *Client) GetOrg(id string) (*Org, error) {
	response, err := c.call("getOrg", []interface{}{id})
	if err != nil {
		return nil, err
	}
	var result Org
	if err := c.parseResponse(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse GetOrg response: %v", err)
	}
	return &result, nil
}

// GetAuthInfo fetches the AuthInfo for the current session (mirrors
// api.getAuthInfo in TS).
func (c *Client) GetAuthInfo() (*AuthInfo, error) {
	response, err := c.call("getAuthInfo", []interface{}{})
	if err != nil {
		return nil, err
	}
	var result AuthInfo
	if err := c.parseResponse(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse GetAuthInfo response: %v", err)
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
	account.PublicKey = base64.RawURLEncoding.EncodeToString(publicKeyDER)

	// 3. Set up key derivation parameters (ref: container.ts line 125-133)
	salt := generateRandomBytes(16)
	account.KeyParams = KeyParams{
		Algorithm:  "PBKDF2",
		Hash:       "SHA-256",
		KeySize:    256,
		Iterations: 100000,
		Salt:       base64.RawURLEncoding.EncodeToString(salt),
		Version:    "4.0.0",
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
		IV:             base64.RawURLEncoding.EncodeToString(iv),
		AdditionalData: base64.RawURLEncoding.EncodeToString(additionalData),
		Kind:           "r",
		Version:        "4.0.0",
	}

	// 6. Create account secrets (private key + signing key)
	// PKCS#8 matches browser TS WebCryptoProvider in apps/packages/sdk/src/crypto.ts:
	// generateKey(RSA) → subtle.exportKey('pkcs8', ...), and _decryptRSA → importKey('pkcs8', ...).
	// PKCS#1 (MarshalPKCS1PrivateKey) breaks that import and vault accessor unwrap on the client.
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal PKCS8 private key: %w", err)
	}
	signingKey := generateRandomBytes(32) // HMAC key

	// Combine private key and signing key into account secrets (TS AccountSecrets uses @AsBytes → url-safe unpadded).
	accountSecrets := struct {
		SigningKey Base64Bytes `json:"signingKey"`
		PrivateKey Base64Bytes `json:"privateKey"`
	}{
		SigningKey: Base64Bytes(signingKey),
		PrivateKey: Base64Bytes(privateKeyDER),
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
	account.EncryptedData = base64.RawURLEncoding.EncodeToString(encryptedData)

	log.Printf("Account initialized with RSA key pair and encryption parameters")
	log.Printf("Public key length: %d bytes", len(publicKeyDER))
	log.Printf("Encrypted data length: %d bytes", len(encryptedData))

	return nil
}

// encryptAESGCM is kept for callers that already have a *App receiver; it
// just delegates to the package-level aesGCMEncrypt helper.
func (a *App) encryptAESGCM(key, plaintext, iv, additionalData []byte) ([]byte, error) {
	return aesGCMEncrypt(key, plaintext, iv, additionalData)
}
