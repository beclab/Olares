package wizard

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// ============================================================================
// Interface Definitions
// ============================================================================

// Platform interface for authentication operations
type Platform interface {
	StartAuthRequest(opts StartAuthRequestOptions) (*StartAuthRequestResponse, error)
	CompleteAuthRequest(req *StartAuthRequestResponse) (*AuthenticateResponse, error)
}

// AppAPI interface for app-level operations
type AppAPI interface {
	StartAuthRequest(params StartAuthRequestParams) (*StartAuthRequestResponse, error)
	CompleteAuthRequest(params CompleteAuthRequestParams) (*CompleteAuthRequestResponse, error)
}

// ClientState interface for managing client session state
type ClientState interface {
	GetSession() *Session
	SetSession(session *Session)
	GetAccount() *Account
	SetAccount(account *Account)
	GetDevice() *DeviceInfo
}

// Sender interface for network transport
type Sender interface {
	Send(req *Request) (*Response, error)
}

// AuthClient interface for authentication clients
type AuthClient interface {
	PrepareAuthentication(params map[string]any) (map[string]any, error)
}

// ============================================================================
// Type Definitions and Enums
// ============================================================================
type AuthType string

const (
	AuthTypeSSI AuthType = "ssi"
)

type AuthPurpose string

const (
	AuthPurposeSignup            AuthPurpose = "signup"
	AuthPurposeLogin             AuthPurpose = "login"
	AuthPurposeRecover           AuthPurpose = "recover"
	AuthPurposeAccessKeyStore    AuthPurpose = "access_key_store"
	AuthPurposeTestAuthenticator AuthPurpose = "test_authenticator"
	AuthPurposeAdminLogin        AuthPurpose = "admin_login"
)

type AccountStatus string

const (
	AccountStatusUnregistered AccountStatus = "unregistered"
	AccountStatusActive       AccountStatus = "active"
	AccountStatusBlocked      AccountStatus = "blocked"
	AccountStatusDeleted      AccountStatus = "deleted"
)

type AuthRequestStatus string

const (
	AuthRequestStatusStarted  AuthRequestStatus = "started"
	AuthRequestStatusVerified AuthRequestStatus = "verified"
	AuthRequestStatusExpired  AuthRequestStatus = "expired"
)

type ErrorCode string

const (
	ErrorCodeAuthenticationFailed ErrorCode = "email_verification_failed"
	ErrorCodeNotFound             ErrorCode = "not_found"
	ErrorCodeServerError          ErrorCode = "server_error"
)

// AccountProvisioning represents account provisioning information
type AccountProvisioning struct {
	ID            string         `json:"id"`
	DID           string         `json:"did"`
	Name          *string        `json:"name,omitempty"`
	AccountID     *string        `json:"accountId,omitempty"`
	Status        string         `json:"status"`
	StatusLabel   string         `json:"statusLabel"`
	StatusMessage string         `json:"statusMessage"`
	ActionURL     *string        `json:"actionUrl,omitempty"`
	ActionLabel   *string        `json:"actionLabel,omitempty"`
	MetaData      map[string]any `json:"metaData,omitempty"`
	SkipTos       bool           `json:"skipTos"`
	BillingPage   any            `json:"billingPage,omitempty"`
	Quota         map[string]any `json:"quota"`
	Features      map[string]any `json:"features"`
	Orgs          []string       `json:"orgs"`
}

type StartAuthRequestResponse struct {
	ID              string               `json:"id"`
	DID             string               `json:"did"`
	Token           string               `json:"token"`
	Data            map[string]any       `json:"data"`
	Type            AuthType             `json:"type"`
	Purpose         AuthPurpose          `json:"purpose"`
	AuthenticatorID string               `json:"authenticatorId"`
	RequestStatus   AuthRequestStatus    `json:"requestStatus"`
	AccountStatus   *AccountStatus       `json:"accountStatus,omitempty"`
	Provisioning    *AccountProvisioning `json:"provisioning,omitempty"`
	DeviceTrusted   bool                 `json:"deviceTrusted"`
}

type AuthenticateRequest struct {
	DID                string                    `json:"did"`
	Type               AuthType                  `json:"type"`
	Purpose            AuthPurpose               `json:"purpose"`
	AuthenticatorIndex int                       `json:"authenticatorIndex"`
	PendingRequest     *StartAuthRequestResponse `json:"pendingRequest,omitempty"`
	Caller             string                    `json:"caller"`
}

type AuthenticateResponse struct {
	DID           string              `json:"did"`
	Token         string              `json:"token"`
	AccountStatus AccountStatus       `json:"accountStatus"`
	Provisioning  AccountProvisioning `json:"provisioning"`
	DeviceTrusted bool                `json:"deviceTrusted"`
}

type StartAuthRequestOptions struct {
	Purpose            AuthPurpose `json:"purpose"`
	Type               *AuthType   `json:"type,omitempty"`
	DID                *string     `json:"did,omitempty"`
	AuthenticatorID    *string     `json:"authenticatorId,omitempty"`
	AuthenticatorIndex *int        `json:"authenticatorIndex,omitempty"`
}

type StartAuthRequestParams struct {
	DID                string      `json:"did"`
	Type               *AuthType   `json:"type,omitempty"`
	SupportedTypes     []AuthType  `json:"supportedTypes"`
	Purpose            AuthPurpose `json:"purpose"`
	AuthenticatorID    *string     `json:"authenticatorId,omitempty"`
	AuthenticatorIndex *int        `json:"authenticatorIndex,omitempty"`
}

type CompleteAuthRequestParams struct {
	ID   string         `json:"id"`
	Data map[string]any `json:"data"`
	DID  string         `json:"did"`
}

type CompleteAuthRequestResponse struct {
	AccountStatus AccountStatus       `json:"accountStatus"`
	DeviceTrusted bool                `json:"deviceTrusted"`
	Provisioning  AccountProvisioning `json:"provisioning"`
}

// Session represents a user session
type Session struct {
	ID  string `json:"id"`
	Key []byte `json:"key,omitempty"`
	// Other session-related fields...
}

// OrgInfo represents organization information
type OrgInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name,omitempty"`
	Revision string `json:"revision,omitempty"`
}

// MainVault represents main vault information
type MainVault struct {
	ID       string `json:"id"`
	Name     string `json:"name,omitempty"`
	Revision string `json:"revision,omitempty"`
}

// AccountSettings represents account settings
type AccountSettings struct {
	// Simplified version, can be extended as needed
}

// EncryptionParams represents AES encryption parameters
type EncryptionParams struct {
	Algorithm      string `json:"algorithm"`      // "AES-GCM"
	TagSize        int    `json:"tagSize"`        // 128
	KeySize        int    `json:"keySize"`        // 256
	IV             string `json:"iv"`             // Base64 encoded initialization vector
	AdditionalData string `json:"additionalData"` // Base64 encoded additional data
	Version        string `json:"version"`        // "3.0.14"
}

// KeyParams represents PBKDF2 key derivation parameters
type KeyParams struct {
	Algorithm  string `json:"algorithm"`  // "PBKDF2"
	Hash       string `json:"hash"`       // "SHA-256"
	KeySize    int    `json:"keySize"`    // 256
	Iterations int    `json:"iterations"` // 100000
	Salt       string `json:"salt"`       // Base64 encoded salt
	Version    string `json:"version"`    // "3.0.14"
}

type Account struct {
	ID               string           `json:"id"`
	DID              string           `json:"did"`
	Name             string           `json:"name"`
	Local            bool             `json:"local,omitempty"`
	Created          string           `json:"created,omitempty"`          // ISO 8601 format
	Updated          string           `json:"updated,omitempty"`          // ISO 8601 format
	PublicKey        string           `json:"publicKey,omitempty"`        // Base64 encoded RSA public key
	EncryptedData    string           `json:"encryptedData,omitempty"`    // Base64 encoded encrypted data
	EncryptionParams EncryptionParams `json:"encryptionParams,omitempty"` // AES encryption parameters
	KeyParams        KeyParams        `json:"keyParams,omitempty"`        // PBKDF2 key derivation parameters
	MainVault        MainVault        `json:"mainVault"`                  // Main vault information
	Orgs             []OrgInfo        `json:"orgs"`                       // Organization list (important: prevent undefined)
	Revision         string           `json:"revision,omitempty"`         // Version control
	Kid              string           `json:"kid,omitempty"`              // Key ID
	Settings         AccountSettings  `json:"settings,omitempty"`         // Account settings
	Version          string           `json:"version,omitempty"`          // Version
}

type DeviceInfo struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
	// Other device-related fields...
}

// Request represents an RPC request
type Request struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params,omitempty"`
	Device *DeviceInfo   `json:"device,omitempty"`
	Auth   *RequestAuth  `json:"auth,omitempty"`
}

type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  *ErrorInfo  `json:"error,omitempty"`
}

// ISOTime is a custom time type that ensures JSON serialization matches JavaScript toISOString() format
type ISOTime time.Time

// MarshalJSON implements JSON serialization using JavaScript toISOString() format
func (t ISOTime) MarshalJSON() ([]byte, error) {
	// JavaScript toISOString() format: 2006-01-02T15:04:05.000Z
	// Ensure milliseconds are always 3 digits
	utcTime := time.Time(t).UTC()
	timeStr := fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d.%03dZ",
		utcTime.Year(), utcTime.Month(), utcTime.Day(),
		utcTime.Hour(), utcTime.Minute(), utcTime.Second(),
		utcTime.Nanosecond()/1000000)
	return json.Marshal(timeStr)
}

// UnmarshalJSON implements JSON deserialization
func (t *ISOTime) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	parsed, err := time.Parse("2006-01-02T15:04:05.000Z", str)
	if err != nil {
		return err
	}

	*t = ISOTime(parsed)
	return nil
}

// Unix returns Unix timestamp for compatibility
func (t ISOTime) Unix() int64 {
	return time.Time(t).Unix()
}

type RequestAuth struct {
	Session   string      `json:"session"`
	Time      ISOTime     `json:"time"`      // Use custom ISOTime type
	Signature Base64Bytes `json:"signature"` // Use Base64Bytes to automatically handle base64 encoding
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Base64Bytes automatically handles base64 encoding/decoding for byte arrays
type Base64Bytes []byte

// UnmarshalJSON implements JSON deserialization, automatically decoding from base64 string
func (b *Base64Bytes) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	// Server uses URL-safe base64 encoding by default (ref: encoding.ts line 366: urlSafe = true)
	// Try base64url decoding first
	decoded, err := base64.URLEncoding.DecodeString(str)
	if err != nil {
		// If base64url fails, try raw base64url decoding
		decoded, err = base64.RawURLEncoding.DecodeString(str)
		if err != nil {
			// Finally try standard base64 decoding
			decoded, err = base64.StdEncoding.DecodeString(str)
			if err != nil {
				return fmt.Errorf("failed to decode base64url/base64: %v", err)
			}
		}
	}

	*b = Base64Bytes(decoded)
	return nil
}

// MarshalJSON implements JSON serialization, automatically encoding to base64 string
func (b Base64Bytes) MarshalJSON() ([]byte, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(b))
	return json.Marshal(encoded)
}

// Bytes returns the underlying byte array
func (b Base64Bytes) Bytes() []byte {
	return []byte(b)
}

// ============================================================================
// Vault and VaultItem Structures
// ============================================================================

// VaultType represents the type of vault item
type VaultType int

const (
	VaultTypeDefault           VaultType = 0
	VaultTypeLogin             VaultType = 1
	VaultTypeCard              VaultType = 2
	VaultTypeTerminusTotp      VaultType = 3
	VaultTypeOlaresSSHPassword VaultType = 4
)

// FieldType represents the type of field in a vault item
type FieldType string

const (
	FieldTypeUsername  FieldType = "username"
	FieldTypePassword  FieldType = "password"
	FieldTypeApiSecret FieldType = "apiSecret"
	FieldTypeMnemonic  FieldType = "mnemonic"
	FieldTypeUrl       FieldType = "url"
	FieldTypeEmail     FieldType = "email"
	FieldTypeDate      FieldType = "date"
	FieldTypeMonth     FieldType = "month"
	FieldTypeCredit    FieldType = "credit"
	FieldTypePhone     FieldType = "phone"
	FieldTypePin       FieldType = "pin"
	FieldTypeTotp      FieldType = "totp"
	FieldTypeNote      FieldType = "note"
	FieldTypeText      FieldType = "text"
)

// Field represents a field in a vault item
type Field struct {
	Name  string    `json:"name"`
	Type  FieldType `json:"type"`
	Value string    `json:"value"`
}

// VaultItem represents an item in a vault
type VaultItem struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      VaultType `json:"type"`
	Icon      string    `json:"icon,omitempty"`
	Fields    []Field   `json:"fields"`
	Tags      []string  `json:"tags"`
	Updated   string    `json:"updated"` // ISO 8601 format
	UpdatedBy string    `json:"updatedBy"`
}

// Vault represents a vault containing items
type Vault struct {
	Kind         string      `json:"kind"`                    // Always "vault" for Vault objects
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Owner        string      `json:"owner"`
	Created      string      `json:"created"`                 // ISO 8601 format
	Updated      string      `json:"updated"`                 // ISO 8601 format
	Revision     string      `json:"revision,omitempty"`
	Items        []VaultItem `json:"items,omitempty"`
	KeyParams    interface{} `json:"keyParams,omitempty"`
	EncryptionParams interface{} `json:"encryptionParams,omitempty"`
	Accessors    interface{} `json:"accessors,omitempty"`
	EncryptedData interface{} `json:"encryptedData,omitempty"`
	Version      string      `json:"version,omitempty"`       // Serialization version
}

// ItemTemplate represents a template for creating vault items
type ItemTemplate struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Icon   string  `json:"icon"`
	Fields []Field `json:"fields"`
}

// GetAuthenticatorTemplate returns the authenticator template for TOTP items
func GetAuthenticatorTemplate() *ItemTemplate {
	return &ItemTemplate{
		ID:   "authenticator",
		Name: "Authenticator",
		Icon: "authenticator",
		Fields: []Field{
			{
				Name:  "One-Time Password",
				Type:  FieldTypeTotp,
				Value: "", // Will be set with MFA token
			},
		},
	}
}

// JWS-related data structures removed, using Web5 library's jwt.Sign() method directly
// UserItem and JWSSignatureInput removed as they were not actually used
