package wizard

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/beclab/Olares/cli/pkg/web5/crypto"
	"github.com/beclab/Olares/cli/pkg/web5/dids/did"
	"github.com/beclab/Olares/cli/pkg/web5/dids/didcore"
	"github.com/beclab/Olares/cli/pkg/web5/jwk"
	"github.com/beclab/Olares/cli/pkg/web5/jwt"
)

// Note: DID key-related implementation is now in did_key_utils.go

// UserStore implementation using actual DID keys
type UserStore struct {
	terminusName string
	mnemonic     string
	did          string
	privateJWK   *jwk.JWK // Direct use of Web5 JWK structure
	mfa          string   // Store MFA token
}

func (u *UserStore) GetTerminusName() string {
	return u.terminusName
}

func (u *UserStore) GetDid() string {
	return u.did
}

// SetMFA saves MFA token
func (u *UserStore) SetMFA(mfa string) error {
	u.mfa = mfa
	log.Printf("MFA token saved to UserStore: %s", mfa)
	return nil
}

// GetMFA retrieves MFA token
func (u *UserStore) GetMFA() (string, error) {
	if u.mfa == "" {
		return "", fmt.Errorf("MFA token not found")
	}
	return u.mfa, nil
}

func (u *UserStore) GetPrivateJWK() *jwk.JWK {
	return u.privateJWK
}

// NewUserStore creates user store, generating all keys from mnemonic (using methods from did_key_utils.go)
func NewUserStore(mnemonic, terminusName string) (*UserStore, error) {
	log.Printf("Creating RealUserStore from mnemonic")

	// 1. Generate complete DID key result using methods from did_key_utils.go
	result, err := GetPrivateJWK(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to generate DID key: %w", err)
	}

	log.Printf("Generated DID from mnemonic: %s", result.DID)

	// 2. Direct use of Web5's jwk.JWK, no conversion needed
	privateJWK := &result.PrivateJWK

	return &UserStore{
		terminusName: terminusName,
		mnemonic:     mnemonic,
		did:          result.DID,
		privateJWK:   privateJWK,
	}, nil
}

// UserStore method implementations
func (u *UserStore) GetCurrentID() string {
	return u.did
}

// GetCurrentUser method removed as it was not actually used

func (u *UserStore) GetCurrentUserPrivateKey() (*jwk.JWK, error) {
	return u.privateJWK, nil
}

// createBearerDIDFromPrivateKey creates BearerDID from private key
func (u *UserStore) createBearerDIDFromPrivateKey() (did.BearerDID, error) {
	// 1. Create LocalKeyManager
	keyManager := crypto.NewLocalKeyManager()

	// 2. Direct use of our stored Web5 JWK, no conversion needed
	privateJWK := *u.privateJWK

	// 3. Import private key to KeyManager
	keyID, err := keyManager.ImportKey(privateJWK)
	if err != nil {
		return did.BearerDID{}, fmt.Errorf("failed to import private key: %w", err)
	}

	// 4. Get public key
	publicJWK, err := keyManager.GetPublicKey(keyID)
	if err != nil {
		return did.BearerDID{}, fmt.Errorf("failed to get public key: %w", err)
	}

	// 5. Set public key's KID
	publicJWK.KID = u.did
	publicJWK.USE = "sig"
	publicJWK.ALG = "EdDSA"

	// 6. Parse DID
	parsedDID, err := did.Parse(u.did)
	if err != nil {
		return did.BearerDID{}, fmt.Errorf("failed to parse DID: %w", err)
	}

	// 7. Create DID Document
	document := didcore.Document{
		Context: []string{
			"https://www.w3.org/ns/did/v1",
			"https://w3id.org/security/suites/ed25519-2020/v1",
		},
		ID: u.did,
		VerificationMethod: []didcore.VerificationMethod{
			{
				ID:           u.did,
				Type:         "JsonWebKey2020",
				Controller:   u.did,
				PublicKeyJwk: &publicJWK,
			},
		},
		Authentication:       []string{"#" + u.did},
		AssertionMethod:      []string{"#" + u.did},
		CapabilityDelegation: []string{"#" + u.did},
		CapabilityInvocation: []string{"#" + u.did},
	}

	fmt.Printf("publicJWK: %v", document)
	// 8. Create BearerDID
	bearerDID := did.BearerDID{
		DID:        parsedDID,
		KeyManager: keyManager,
		Document:   document,
	}

	return bearerDID, nil
}

// SignJWS performs real DID key JWS signing (using BearerDID created from private key)
func (u *UserStore) SignJWS(payload map[string]any) (string, error) {
	log.Printf("Creating real JWS signature for DID: %s", u.did)

	// Create BearerDID from private key
	bearerDID, err := u.createBearerDIDFromPrivateKey()
	if err != nil {
		return "", fmt.Errorf("failed to create BearerDID from private key: %w", err)
	}

	// Build JWT Claims (ref: example/main.go)
	claims := jwt.Claims{
		Issuer: bearerDID.URI,
		Misc:   payload, // Direct use of passed payload
	}

	// Ensure payload has necessary fields
	if claims.Misc == nil {
		claims.Misc = make(map[string]interface{})
	}

	// Add timestamp if not present
	if _, exists := claims.Misc["time"]; !exists {
		claims.Misc["time"] = fmt.Sprintf("%d", time.Now().UnixMilli())
	}

	// Use Web5 JWT signing (ref: example/main.go)
	signedJWT, err := jwt.Sign(claims, bearerDID, jwt.Type("JWT"))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	log.Printf("Real JWS created successfully with Web5")
	log.Printf("Bearer DID: %s", bearerDID.URI)
	log.Printf("JWS: %s", signedJWT[:100]+"...")

	return signedJWT, nil
}

const TerminusDefaultDomain = "olares.cn"

func (u *UserStore) GetTerminusURL() string {
	array := strings.Split(u.terminusName, "@")
	localURL := u.getLocalURL()

	if len(array) == 2 {
		return fmt.Sprintf("https://%s%s.%s", localURL, array[0], array[1])
	} else {
		return fmt.Sprintf("https://%s%s.%s", localURL, array[0], TerminusDefaultDomain)
	}
}

func (u *UserStore) GetAuthURL() string {
	array := strings.Split(u.terminusName, "@")
	localURL := u.getLocalURL()

	if len(array) == 2 {
		return fmt.Sprintf("https://auth.%s%s.%s/", localURL, array[0], array[1])
	} else {
		return fmt.Sprintf("https://auth.%s%s.%s/", localURL, array[0], TerminusDefaultDomain)
	}
}

func (u *UserStore) GetVaultURL() string {
	array := strings.Split(u.terminusName, "@")
	localURL := u.getLocalURL()

	if len(array) == 2 {
		return fmt.Sprintf("https://vault.%s%s.%s/server", localURL, array[0], array[1])
	} else {
		return fmt.Sprintf("https://vault.%s%s.%s/server", localURL, array[0], TerminusDefaultDomain)
	}
}

func (u *UserStore) getLocalURL() string {
	return ""
}

func (u *UserStore) GetLocalName() string {
	array := strings.Split(u.terminusName, "@")
	return array[0]
}

func (u *UserStore) GetDomainName() string {
	array := strings.Split(u.terminusName, "@")
	if len(array) == 2 {
		return array[1]
	} else {
		return TerminusDefaultDomain
	}
}
