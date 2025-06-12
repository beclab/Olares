package jws

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/beclab/Olares/cli/pkg/web5/crypto/dsa"
	"github.com/beclab/Olares/cli/pkg/web5/dids"

	_did "github.com/beclab/Olares/cli/pkg/web5/dids/did"
	"github.com/beclab/Olares/cli/pkg/web5/dids/didcore"
)

// Decode decodes the given JWS string into a [Decoded] type
//
// # Note
//
// The given JWS input is assumed to be a [compact JWS]
//
// [compact JWS]: https://datatracker.ietf.org/doc/html/rfc7515#section-7.1
func Decode(jws string, opts ...DecodeOption) (Decoded, error) {
	o := decodeOptions{}

	for _, opt := range opts {
		opt(&o)
	}

	parts := strings.Split(jws, ".")
	if len(parts) != 3 {
		return Decoded{}, fmt.Errorf("malformed JWS. Expected 3 parts, got %d", len(parts))
	}

	header, err := DecodeHeader(parts[0])
	if err != nil {
		return Decoded{}, fmt.Errorf("malformed JWS. Failed to decode header: %w", err)
	}

	var payload []byte
	if o.payload == nil {
		payload, err = base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return Decoded{}, fmt.Errorf("malformed JWS. Failed to decode payload: %w", err)
		}
	} else {
		payload = o.payload
		parts[1] = base64.RawURLEncoding.EncodeToString(payload)
	}

	signature, err := DecodeSignature(parts[2])
	if err != nil {
		return Decoded{}, fmt.Errorf("malformed JWS. Failed to decode signature: %w", err)
	}

	if header.KID == "" {
		return Decoded{}, errors.New("malformed JWS. Expected header to contain kid")
	}

	signerDID, err := _did.Parse(header.KID)
	if err != nil {
		return Decoded{}, fmt.Errorf("malformed JWS. Failed to parse kid: %w", err)
	}

	return Decoded{
		Header:    header,
		Payload:   payload,
		Signature: signature,
		Parts:     parts,
		SignerDID: signerDID,
	}, nil
}

type decodeOptions struct {
	payload []byte
}

// DecodeOption represents an option that can be passed to [Decode] or [Verify].
type DecodeOption func(opts *decodeOptions)

// Payload can be passed to [Decode] or [Verify] to provide a detached payload.
// More info on detached payloads can be found [here].
//
// [here]: https://datatracker.ietf.org/doc/html/rfc7515#appendix-F
func Payload(p []byte) DecodeOption {
	return func(opts *decodeOptions) {
		opts.payload = p
	}
}

// DecodeHeader decodes the base64url encoded JWS header into a [Header]
func DecodeHeader(base64UrlEncodedHeader string) (Header, error) {
	bytes, err := base64.RawURLEncoding.DecodeString(base64UrlEncodedHeader)
	if err != nil {
		return Header{}, err
	}

	var header Header
	err = json.Unmarshal(bytes, &header)
	if err != nil {
		return Header{}, err
	}

	return header, nil
}

// DecodeSignature decodes the base64url encoded JWS signature into a byte array
func DecodeSignature(base64UrlEncodedSignature string) ([]byte, error) {
	signature, err := base64.RawURLEncoding.DecodeString(base64UrlEncodedSignature)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// options that sign function can take
type signOpts struct {
	selector didcore.VMSelector
	detached bool
	typ      string
}

// SignOpt is a type that represents an option that can be passed to [olares/github.com/beclab/Olares/cli/pkg/web5/jws.Sign].
type SignOpt func(opts *signOpts)

// Purpose is an option that can be passed to [olares/github.com/beclab/Olares/cli/pkg/web5/jws.Sign].
// It is used to select the appropriate key to sign with
func Purpose(p string) SignOpt {
	return func(opts *signOpts) {
		opts.selector = didcore.Purpose(p)
	}
}

// VerificationMethod is an option that can be passed to [olares/github.com/beclab/Olares/cli/pkg/web5/jws.Sign].
// It is used to select the appropriate key to sign with
func VerificationMethod(id string) SignOpt {
	return func(opts *signOpts) {
		opts.selector = didcore.ID(id)
	}
}

// VMSelector is an option that can be passed to [olares/github.com/beclab/Olares/cli/pkg/web5/jws.Sign].
// It is used to select the appropriate key to sign with
func VMSelector(selector didcore.VMSelector) SignOpt {
	return func(opts *signOpts) {
		opts.selector = selector
	}
}

// DetachedPayload is an option that can be passed to [olares/github.com/beclab/Olares/cli/pkg/web5/jws.Sign].
// It is used to indicate whether the payload should be included in the signature.
// More details can be found [here].
//
// [Specification]: https://datatracker.ietf.org/doc/html/rfc7515#appendix-F
func DetachedPayload(detached bool) SignOpt {
	return func(opts *signOpts) {
		opts.detached = detached
	}
}

// Type is an option that can be passed to [olares/github.com/beclab/Olares/cli/pkg/web5/jws.Sign].
// It is used to set the `typ` JWS header value
func Type(typ string) SignOpt {
	return func(opts *signOpts) {
		opts.typ = typ
	}
}

// Sign signs the provided payload with a key associated to the provided DID.
// if no purpose is provided, the default is "assertionMethod". Passing Detached(true)
// will return a compact JWS with detached content
func Sign(payload []byte, did _did.BearerDID, opts ...SignOpt) (string, error) {
	o := signOpts{selector: nil, detached: false}
	for _, opt := range opts {
		opt(&o)
	}

	sign, verificationMethod, err := did.GetSigner(o.selector)
	if err != nil {
		return "", fmt.Errorf("failed to get signer: %w", err)
	}

	jwa, err := dsa.GetJWA(*verificationMethod.PublicKeyJwk)
	if err != nil {
		return "", fmt.Errorf("failed to determine alg: %w", err)
	}

	keyID := did.Document.GetAbsoluteResourceID(verificationMethod.ID)
	header := Header{ALG: jwa, KID: keyID, TYP: o.typ}
	base64UrlEncodedHeader, err := header.Encode()
	if err != nil {
		return "", fmt.Errorf("failed to base64 url encode header: %w", err)
	}

	base64UrlEncodedPayload := base64.RawURLEncoding.EncodeToString(payload)

	toSign := base64UrlEncodedHeader + "." + base64UrlEncodedPayload
	toSignBytes := []byte(toSign)

	signature, err := sign(toSignBytes)
	if err != nil {
		return "", fmt.Errorf("failed to compute signature: %w", err)
	}

	base64UrlEncodedSignature := base64.RawURLEncoding.EncodeToString(signature)

	var compactJWS string
	if o.detached {
		compactJWS = base64UrlEncodedHeader + "." + "." + base64UrlEncodedSignature
	} else {
		compactJWS = toSign + "." + base64UrlEncodedSignature
	}

	return compactJWS, nil
}

// Verify verifies the given compactJWS by resolving the DID Document from the kid header value
// and using the associated public key found by resolving the DID Document
func Verify(compactJWS string, opts ...DecodeOption) (Decoded, error) {
	decodedJWS, err := Decode(compactJWS, opts...)
	if err != nil {
		return decodedJWS, fmt.Errorf("signature verification failed: %w", err)
	}

	err = decodedJWS.Verify()

	return decodedJWS, err
}

// Decoded is a compact JWS decoded into its parts
type Decoded struct {
	Header    Header
	Payload   []byte
	Signature []byte
	Parts     []string
	SignerDID _did.DID
}

// Verify verifies the given compactJWS by resolving the DID Document from the kid header value
// and using the associated public key found by resolving the DID Document
func (jws Decoded) Verify() error {
	if jws.Header.ALG == "" || jws.Header.KID == "" {
		return errors.New("malformed JWS header. alg and kid are required")
	}

	did, err := _did.Parse(jws.Header.KID)
	if err != nil {
		return errors.New("malformed JWS header. kid must be a DID URL")
	}

	resolutionResult, err := dids.Resolve(did.URI)
	if err != nil {
		return fmt.Errorf("failed to resolve DID: %w", err)
	}

	vmSelector := didcore.ID(did.URL())
	verificationMethod, err := resolutionResult.Document.SelectVerificationMethod(vmSelector)
	if err != nil {
		return fmt.Errorf("kid does not match any verification method %w", err)
	}

	toVerify := jws.Parts[0] + "." + jws.Parts[1]

	verified, err := dsa.Verify([]byte(toVerify), jws.Signature, *verificationMethod.PublicKeyJwk)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	if !verified {
		return errors.New("invalid signature")
	}

	return nil
}

// Header represents a JWS (JSON Web Signature) header. See [Specification] for more details.
// [Specification]: https://datatracker.ietf.org/doc/html/rfc7515#section-4
type Header struct {
	// Ide	ntifies the cryptographic algorithm used to secure the JWS. The JWS Signature value is not
	// valid if the "alg" value does not represent a supported algorithm or if there is not a key for
	// use with that algorithm associated with the party that digitally signed or MACed the content.
	//
	// "alg" values should either be registered in the IANA "JSON Web Signature and Encryption
	// Algorithms" registry or be a value that contains a Collision-Resistant Name. The "alg" value is
	// a case-sensitive ASCII string.  This Header Parameter MUST be present and MUST be understood
	// and processed by implementations.
	//
	// [Specification]: https://datatracker.ietf.org/doc/html/rfc7515#section-4.1.1
	ALG string `json:"alg,omitempty"`
	// Key ID Header Parameter https://datatracker.ietf.org/doc/html/rfc7515#section-4.1.4
	KID string `json:"kid,omitempty"`
	// Type Header Parameter https://datatracker.ietf.org/doc/html/rfc7515#section-4.1.9
	TYP string `json:"typ,omitempty"`
}

// Encode returns the base64url encoded header.
func (j Header) Encode() (string, error) {
	bytes, err := json.Marshal(j)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
