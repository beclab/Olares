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
	Orgs          []string       `json:"orgs,omitempty"`
}

// OrgProvisioning mirrors the subset of TS OrgProvisioning that can appear
// inside AuthInfo.provisioning.orgs.
type OrgProvisioning struct {
	OrgID         string         `json:"orgId"`
	OrgName       string         `json:"orgName,omitempty"`
	Status        string         `json:"status,omitempty"`
	StatusLabel   string         `json:"statusLabel,omitempty"`
	StatusMessage any            `json:"statusMessage,omitempty"`
	ActionURL     *string        `json:"actionUrl,omitempty"`
	ActionLabel   *string        `json:"actionLabel,omitempty"`
	MetaData      map[string]any `json:"metaData,omitempty"`
	AutoCreate    bool           `json:"autoCreate,omitempty"`
	Quota         map[string]any `json:"quota,omitempty"`
	Features      map[string]any `json:"features,omitempty"`
}

// Provisioning mirrors apps/packages/sdk/src/core/provisioning.ts Provisioning.
type Provisioning struct {
	Account *AccountProvisioning `json:"account,omitempty"`
	Orgs    []OrgProvisioning    `json:"orgs,omitempty"`
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
	Kind           string `json:"kind,omitempty"`    // TS Serializable short kind, e.g. "r"
	Version        string `json:"version,omitempty"` // e.g. "4.0.0"
}

// KeyParams represents PBKDF2 key derivation parameters
type KeyParams struct {
	Algorithm  string `json:"algorithm"`  // "PBKDF2"
	Hash       string `json:"hash"`       // "SHA-256"
	KeySize    int    `json:"keySize"`    // 256
	Iterations int    `json:"iterations"` // 100000
	Salt       string `json:"salt"`       // Base64 encoded salt
	Version    string `json:"version,omitempty"`
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

// Request represents an RPC request.
//
// IMPORTANT: do NOT add `omitempty` to Params. The TS client always
// sends `params: []` on the wire (see apps/packages/sdk/src/core/client.ts
// line 41-52: `typeof input === 'undefined' ? [] : [...]`), and the server
// signs against `JSON.stringify(req.params)` (where `[]` -> "[]" but
// `undefined` -> "undefined"). With `omitempty` Go would drop the field
// for empty param calls (e.g. `getAccount`), causing a signature mismatch:
// client signs `..._[]` while server signs `..._undefined`.
type Request struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
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
		decoded, err = base64.RawURLEncoding.DecodeString(str)
	}
	if err != nil {
		decoded, err = base64.StdEncoding.DecodeString(str)
	}
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(str)
	}
	if err != nil {
		return fmt.Errorf("failed to decode base64url/base64: %w", err)
	}

	*b = Base64Bytes(decoded)
	return nil
}

// MarshalJSON encodes like TS bytesToBase64 (url-safe, no padding);
// see apps/packages/sdk/src/core/base64.ts fromByteArray(..., urlSafe=true).
func (b Base64Bytes) MarshalJSON() ([]byte, error) {
	encoded := base64.RawURLEncoding.EncodeToString([]byte(b))
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

// Accessor mirrors apps/packages/sdk/src/core/container.ts Accessor.
type Accessor struct {
	ID           string      `json:"id"`
	EncryptedKey Base64Bytes `json:"encryptedKey"`
	PublicKey    Base64Bytes `json:"publicKey,omitempty"`
	Kind         string      `json:"kind,omitempty"`
	Version      string      `json:"version,omitempty"`
}

// RSAEncryptionParams mirrors the same-named TS class. Only RSA-OAEP /
// SHA-256 is supported (matching the server defaults).
//
// Do not put Serializable "kind"/"version" on the wire: WebCryptoProvider in
// apps/packages/sdk/src/crypto.ts passes keyParams into subtle.importKey/decrypt
// via Object.assign; extra keys (e.g. kind, version) can cause DataError on decrypt.
type RSAEncryptionParams struct {
	Algorithm string `json:"algorithm"` // "RSA-OAEP"
	Hash      string `json:"hash"`      // "SHA-256"
}

// MarshalJSON emits only algorithm+hash so WebCryptoProvider (crypto.ts) never
// sees Serializable extras (kind/version) on keyParams after Object.assign.
func (p RSAEncryptionParams) MarshalJSON() ([]byte, error) {
	type wire struct {
		Algorithm string `json:"algorithm"`
		Hash      string `json:"hash"`
	}
	alg, h := p.Algorithm, p.Hash
	if alg == "" {
		alg = "RSA-OAEP"
	}
	if h == "" {
		h = "SHA-256"
	}
	return json.Marshal(wire{Algorithm: alg, Hash: h})
}

// UnmarshalJSON ignores unknown fields (e.g. legacy kind/version from server).
func (p *RSAEncryptionParams) UnmarshalJSON(data []byte) error {
	type wire struct {
		Algorithm string `json:"algorithm"`
		Hash      string `json:"hash"`
	}
	var w wire
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	p.Algorithm = w.Algorithm
	p.Hash = w.Hash
	if p.Algorithm == "" {
		p.Algorithm = "RSA-OAEP"
	}
	if p.Hash == "" {
		p.Hash = "SHA-256"
	}
	return nil
}

// NewRSAEncryptionParams constructs the canonical params used by TS/WebCrypto.
func NewRSAEncryptionParams() RSAEncryptionParams {
	return RSAEncryptionParams{Algorithm: "RSA-OAEP", Hash: "SHA-256"}
}

// Vault is the Go counterpart of apps/packages/sdk/src/core/vault.ts.
//
// `items` and the cleartext shared key live only in memory after a
// successful Unlock call; they are not serialized to JSON.
type Vault struct {
	Kind             string              `json:"kind"`
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	Owner            string              `json:"owner"`
	Org              *OrgInfo            `json:"org,omitempty"`
	Created          string              `json:"created"`
	Updated          string              `json:"updated"`
	Revision         string              `json:"revision"` // required on updateVault (server revision check)
	KeyParams        RSAEncryptionParams `json:"keyParams"`
	EncryptionParams EncryptionParams    `json:"encryptionParams"`
	Accessors        []Accessor          `json:"accessors"`
	EncryptedData    Base64Bytes         `json:"encryptedData,omitempty"`
	Version          string              `json:"version,omitempty"`

	// Runtime-only state populated by Unlock() / UpdateAccessors() / Commit().
	items  *VaultItemCollection `json:"-"`
	aesKey []byte               `json:"-"`
}

// VaultItemCollection mirrors apps/packages/sdk/src/core/collection.ts.
// Items keyed by id; Changes records the last time an item was modified
// locally so that merge() knows which side wins.
type VaultItemCollection struct {
	Items   map[string]VaultItem `json:"-"`
	Changes map[string]ISOTime   `json:"-"`
}

// NewVaultItemCollection returns an empty collection.
func NewVaultItemCollection() *VaultItemCollection {
	return &VaultItemCollection{
		Items:   map[string]VaultItem{},
		Changes: map[string]ISOTime{},
	}
}

// vaultItemCollectionRaw is the on-disk shape produced by TS
// VaultItemCollection._toRaw: { items: VaultItem[], changes: [string,string][] }.
type vaultItemCollectionRaw struct {
	Items   []VaultItem `json:"items"`
	Changes [][2]string `json:"changes"`
}

// ToBytes serializes the collection in a way compatible with the TS
// implementation (items: array, changes: [[id, iso-time], ...]).
func (c *VaultItemCollection) ToBytes() ([]byte, error) {
	if c == nil {
		c = NewVaultItemCollection()
	}
	raw := vaultItemCollectionRaw{
		Items:   make([]VaultItem, 0, len(c.Items)),
		Changes: make([][2]string, 0, len(c.Changes)),
	}
	for _, item := range c.Items {
		raw.Items = append(raw.Items, item)
	}
	for id, t := range c.Changes {
		raw.Changes = append(raw.Changes, [2]string{id, time.Time(t).UTC().Format(time.RFC3339Nano)})
	}
	return json.Marshal(raw)
}

// FromBytes deserializes a TS-compatible payload back into the collection.
func (c *VaultItemCollection) FromBytes(data []byte) error {
	var raw vaultItemCollectionRaw
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to parse VaultItemCollection: %w", err)
	}
	c.Items = make(map[string]VaultItem, len(raw.Items))
	for _, item := range raw.Items {
		c.Items[item.ID] = item
	}
	c.Changes = make(map[string]ISOTime, len(raw.Changes))
	for _, ch := range raw.Changes {
		t, err := time.Parse(time.RFC3339Nano, ch[1])
		if err != nil {
			t, _ = time.Parse(time.RFC3339, ch[1])
		}
		c.Changes[ch[0]] = ISOTime(t)
	}
	return nil
}

// Update inserts or replaces the given items, marking each as changed.
func (c *VaultItemCollection) Update(items ...VaultItem) {
	now := time.Now().UTC()
	for _, item := range items {
		item.Updated = now.Format(time.RFC3339Nano)
		c.Items[item.ID] = item
		c.Changes[item.ID] = ISOTime(now)
	}
}

// Remove deletes items by id and records the deletion as a change.
func (c *VaultItemCollection) Remove(items ...VaultItem) {
	now := time.Now().UTC()
	for _, item := range items {
		delete(c.Items, item.ID)
		c.Changes[item.ID] = ISOTime(now)
	}
}

// HasChanges reports whether any local changes are still pending sync.
func (c *VaultItemCollection) HasChanges() bool {
	return c != nil && len(c.Changes) > 0
}

// ClearChanges drops change records older than `before` (zero means all).
func (c *VaultItemCollection) ClearChanges(before time.Time) {
	if c == nil {
		return
	}
	for id, t := range c.Changes {
		if before.IsZero() || !time.Time(t).After(before) {
			delete(c.Changes, id)
		}
	}
}

// Merge mirrors VaultItemCollection.merge in TS: locally-changed items
// always win, otherwise the other side's items overwrite.
func (c *VaultItemCollection) Merge(other *VaultItemCollection) {
	if other == nil {
		return
	}
	for id := range c.Items {
		if _, changed := c.Changes[id]; !changed {
			if _, ok := other.Items[id]; !ok {
				delete(c.Items, id)
			}
		}
	}
	for id, item := range other.Items {
		if _, changed := c.Changes[id]; !changed {
			c.Items[id] = item
		}
	}
}

// Items returns the in-memory collection (allocating if needed).
func (v *Vault) ItemsCollection() *VaultItemCollection {
	if v.items == nil {
		v.items = NewVaultItemCollection()
	}
	return v.items
}

// Org is the minimal subset of apps/packages/sdk/src/core/org.ts that
// the CLI needs in order to fetch / iterate vaults shared via an org.
type Org struct {
	ID        string      `json:"id"`
	Name      string      `json:"name,omitempty"`
	Revision  string      `json:"revision,omitempty"`
	PublicKey Base64Bytes `json:"publicKey,omitempty"`
	Vaults    []OrgVault  `json:"vaults,omitempty"`
	Members   []OrgMember `json:"members,omitempty"`
}

type OrgVault struct {
	ID       string `json:"id"`
	Name     string `json:"name,omitempty"`
	Revision string `json:"revision,omitempty"`
	Readonly bool   `json:"readonly,omitempty"`
}

type OrgMember struct {
	ID        string      `json:"id,omitempty"`
	AccountID string      `json:"accountId,omitempty"`
	DID       string      `json:"did"`
	Name      string      `json:"name,omitempty"`
	PublicKey Base64Bytes `json:"publicKey,omitempty"`
	Role      int         `json:"role,omitempty"`
	Status    string      `json:"status,omitempty"`
	Vaults    []OrgVault  `json:"vaults,omitempty"`
}

// AuthInfo is the minimal subset of apps/packages/sdk/src/core/api.ts
// AuthInfo persisted by the CLI.
type AuthInfo struct {
	Provisioning *Provisioning `json:"provisioning,omitempty"`
}

// UnlockedAccount holds the in-memory secrets derived from a successful
// Account.Unlock call (PBKDF2 → AES-GCM decrypt of EncryptedData).
type UnlockedAccount struct {
	Account    *Account
	MasterKey  []byte
	PrivateKey []byte // PKCS#8 DER (matches apps/packages/sdk/src/crypto.ts subtle.exportKey('pkcs8', ...))
	SigningKey []byte // HMAC key
}

// AccountSecrets is the JSON shape that lives inside the account's
// AES-GCM encryptedData blob.
type AccountSecrets struct {
	SigningKey Base64Bytes `json:"signingKey"`
	PrivateKey Base64Bytes `json:"privateKey"`
	Favorites  []string    `json:"favorites,omitempty"`
	Tags       []TagInfo   `json:"tags,omitempty"`
}

// TagInfo mirrors apps/packages/sdk/src/core/item.ts TagInfo.
type TagInfo struct {
	Name     string  `json:"name"`
	Unlisted *bool   `json:"unlisted,omitempty"`
	Color    *string `json:"color,omitempty"`
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
