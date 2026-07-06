package callerjwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	IssuerURL           = "https://caller-jwt.olares.system/"
	Audience            = "app-gateway-data"
	ClaimAppRef         = "olares.caller.appRef"
	ClaimEntrance       = "olares.entrance"
	ClaimViewer         = "olares.viewer"
	AppJWTSecretName    = "caller-jwt"
	AppJWTSecretDataKey = "token"
	IssuerKeysSecretName = "caller-jwt-issuer-keys"
	SigningKeyPEM       = "signing.pem"
	SigningKeyIDKey     = "signing.kid"
	PreviousKeyPEM      = "previous.pem"
	PreviousKeyIDKey    = "previous.kid"
	MaxTTL              = time.Hour
)

// IssueRequest carries workload identity inputs for a caller JWT-SVID (WI-OC-C2-01).
type IssueRequest struct {
	Namespace          string
	ServiceAccountName string
	AppRef             string
	Entrance           string
	Viewer             string
	TTL                time.Duration
}

// Claims is the frozen C2 v1 JWT claim schema.
type Claims struct {
	jwt.RegisteredClaims
	AppRef   string `json:"olares.caller.appRef"`
	Entrance string `json:"olares.entrance,omitempty"`
	Viewer   string `json:"olares.viewer,omitempty"`
}

// KeyPair holds one RS256 signing key and its JWKS key ID.
type KeyPair struct {
	KID        string
	PrivateKey *rsa.PrivateKey
}

// KeyRing supports overlapping RS256 keys for rotation (WI-OC-C2-01 §8).
type KeyRing struct {
	Active   KeyPair
	Previous *KeyPair
}

// Issuer signs caller JWT-SVIDs with RS256.
type Issuer struct {
	keys KeyRing
}

// NewIssuer constructs an issuer from the supplied key ring.
func NewIssuer(keys KeyRing) (*Issuer, error) {
	if keys.Active.PrivateKey == nil || keys.Active.KID == "" {
		return nil, errors.New("callerjwt: active signing key is required")
	}
	return &Issuer{keys: keys}, nil
}

// Keys exposes the public signing keys for JWKS publication.
func (i *Issuer) Keys() KeyRing {
	if i == nil {
		return KeyRing{}
	}
	return i.keys
}

// Issue signs a caller JWT-SVID for the given workload.
func (i *Issuer) Issue(req IssueRequest) (string, error) {
	if i == nil {
		return "", errors.New("callerjwt: issuer is nil")
	}
	ns := strings.TrimSpace(req.Namespace)
	sa := strings.TrimSpace(req.ServiceAccountName)
	appRef := strings.TrimSpace(req.AppRef)
	if ns == "" || sa == "" || appRef == "" {
		return "", errors.New("callerjwt: namespace, service account, and appRef are required")
	}
	ttl := req.TTL
	if ttl <= 0 {
		ttl = MaxTTL
	}
	if ttl > MaxTTL {
		ttl = MaxTTL
	}
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    IssuerURL,
			Subject:   SPIFFESubject(ns, sa),
			Audience:  jwt.ClaimStrings{Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		AppRef: appRef,
	}
	if v := strings.TrimSpace(req.Entrance); v != "" {
		claims.Entrance = v
	}
	if v := strings.TrimSpace(req.Viewer); v != "" {
		claims.Viewer = v
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = i.keys.Active.KID
	return token.SignedString(i.keys.Active.PrivateKey)
}

// ParseClaims validates a token against the issuer key ring and returns claims.
func (i *Issuer) ParseClaims(tokenString string) (*Claims, error) {
	if i == nil {
		return nil, errors.New("callerjwt: issuer is nil")
	}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
	var claims Claims
	_, err := parser.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		return i.publicKeyForKID(kid)
	})
	if err != nil {
		return nil, err
	}
	if claims.Issuer != IssuerURL {
		return nil, fmt.Errorf("callerjwt: unexpected issuer %q", claims.Issuer)
	}
	if !audienceContains(claims.Audience, Audience) {
		return nil, fmt.Errorf("callerjwt: audience must include %q", Audience)
	}
	return &claims, nil
}

func (i *Issuer) publicKeyForKID(kid string) (*rsa.PublicKey, error) {
	if kid == i.keys.Active.KID {
		return &i.keys.Active.PrivateKey.PublicKey, nil
	}
	if i.keys.Previous != nil && kid == i.keys.Previous.KID {
		return &i.keys.Previous.PrivateKey.PublicKey, nil
	}
	return nil, fmt.Errorf("callerjwt: unknown kid %q", kid)
}

func audienceContains(aud jwt.ClaimStrings, want string) bool {
	for _, v := range aud {
		if v == want {
			return true
		}
	}
	return false
}

// SPIFFESubject returns the SPIFFE-style workload subject for C2 v1.
func SPIFFESubject(namespace, serviceAccount string) string {
	return fmt.Sprintf("spiffe://olares/ns/%s/sa/%s", namespace, serviceAccount)
}

// GenerateKeyPair creates a new RSA-2048 signing key pair.
func GenerateKeyPair() (KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return KeyPair{}, fmt.Errorf("generate rsa key: %w", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	return KeyPair{KID: keyIDFromPEM(pemBytes), PrivateKey: privateKey}, nil
}

// KeyPairFromPEM loads a signing key pair from PKCS#1 PEM bytes.
func KeyPairFromPEM(pemBytes []byte, kid string) (KeyPair, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return KeyPair{}, errors.New("callerjwt: invalid pem")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return KeyPair{}, fmt.Errorf("parse private key: %w", err)
	}
	if kid == "" {
		kid = keyIDFromPEM(pemBytes)
	}
	return KeyPair{KID: kid, PrivateKey: privateKey}, nil
}

func keyIDFromPEM(pemBytes []byte) string {
	h := fnv.New32a()
	_, _ = h.Write(pemBytes)
	return fmt.Sprintf("%08x", h.Sum32())
}

// NewKeyRingForTest returns a key ring with an optional previous key for rotation tests.
func NewKeyRingForTest(withPrevious bool) (KeyRing, error) {
	active, err := GenerateKeyPair()
	if err != nil {
		return KeyRing{}, err
	}
	ring := KeyRing{Active: active}
	if withPrevious {
		prev, err := GenerateKeyPair()
		if err != nil {
			return KeyRing{}, err
		}
		ring.Previous = &prev
	}
	return ring, nil
}
