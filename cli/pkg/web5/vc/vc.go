package vc

import (
	"fmt"
	"time"

	"github.com/beclab/Olares/cli/pkg/web5/dids/did"
	"github.com/beclab/Olares/cli/pkg/web5/jwt"

	"github.com/google/uuid"
)

// these constants are defined in the W3C Verifiable Credential Data Model specification for:
//   - [Context]
//   - [Type]
//
// [Context]: https://www.w3.org/TR/vc-data-model/#contexts
// [Type]: https://www.w3.org/TR/vc-data-model/#dfn-type
const (
	BaseContext = "https://www.w3.org/2018/credentials/v1"
	BaseType    = "VerifiableCredential"
)

// DataModel represents the W3C Verifiable Credential Data Model defined [here]
//
// [here]: https://www.w3.org/TR/vc-data-model/
type DataModel[T CredentialSubject] struct {
	Context           []string           `json:"@context"`                   // https://www.w3.org/TR/vc-data-model/#contexts
	Type              []string           `json:"type"`                       // https://www.w3.org/TR/vc-data-model/#dfn-type
	Issuer            string             `json:"issuer"`                     // https://www.w3.org/TR/vc-data-model/#issuer
	CredentialSubject T                  `json:"credentialSubject"`          // https://www.w3.org/TR/vc-data-model/#credential-subject
	ID                string             `json:"id,omitempty"`               // https://www.w3.org/TR/vc-data-model/#identifiers
	IssuanceDate      string             `json:"issuanceDate"`               // https://www.w3.org/TR/vc-data-model/#issuance-date
	ExpirationDate    string             `json:"expirationDate,omitempty"`   // https://www.w3.org/TR/vc-data-model/#expiration
	CredentialSchema  []CredentialSchema `json:"credentialSchema,omitempty"` // https://www.w3.org/TR/vc-data-model-2.0/#data-schemas
	Evidence          []Evidence         `json:"evidence,omitempty"`         // https://www.w3.org/TR/vc-data-model/#evidence
}

// Evidence represents the evidence property of a Verifiable Credential.
type Evidence struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	// todo is `AdditionalFields` the right name?
	AdditionalFields map[string]interface{}
}

// CredentialSubject is implemented by any type that can be used as the CredentialSubject
// of a Verifiable Credential.
//
// # Note
//
// The VC Data Model specification states that id is not a required field for [CredentialSubject]. However,
// we've chosen to require it in order to necessitate that all credential's include a subject as we were unable
// to find a use case where a credential would not be issued to a single subject. Further, the spec states that
// [vc-jwt] requires the sub be set to the id of the CredentialSubject which becomes difficult to assert while
// also providing the ability to leverage strongly typed claims.
//
// [CredentialSubject]: https://www.w3.org/TR/vc-data-model/#credential-subject
// [vc-jwt]: https://www.w3.org/TR/vc-data-model/#json-web-token
type CredentialSubject interface {
	GetID() string
	SetID(id string)
}

// CredentialSchema represents the credentialSchema property of a Verifiable Credential.
// more information can be found [here]
//
// [here]: https://www.w3.org/TR/vc-data-model-2.0/#data-schemas
type CredentialSchema struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// Claims is a type alias for a map[string]any that can be used to represent the claims of a Verifiable Credential
// when the structure of the claims is not known at compile time.
type Claims map[string]any

// GetID returns the id of the CredentialSubject. used to set the sub claim of a vc-jwt in [vcjwt.Sign]
func (c Claims) GetID() string {
	id, _ := c["id"].(string)
	return id
}

// SetID sets the id of the CredentialSubject. used to set the sub claim of a vc-jwt in [vcjwt.Verify]
func (c Claims) SetID(id string) {
	c["id"] = id
}

// createOptions contains all of the options that can be passed to [Create]
type createOptions struct {
	contexts       []string
	types          []string
	id             string
	issuanceDate   time.Time
	expirationDate time.Time
	schemas        []CredentialSchema
	evidence       []Evidence
}

// CreateOption is the return type of all Option functions that can be passed to [Create]
type CreateOption func(*createOptions)

// Contexts can be used to add additional contexts to the Verifiable Credential created by [Create]
func Contexts(contexts ...string) CreateOption {
	return func(o *createOptions) {
		if o.contexts != nil {
			o.contexts = append(o.contexts, contexts...)
		} else {
			o.contexts = contexts
		}
	}
}

// Schemas can be used to include JSON Schemas within the Verifiable Credential created by [Create]
// more information can be found [here]
//
// [here]: https://www.w3.org/TR/vc-data-model-2.0/#data-schemas
func Schemas(schemas ...string) CreateOption {
	return func(o *createOptions) {
		if o.schemas != nil {
			o.schemas = make([]CredentialSchema, 0, len(schemas))
		}

		for _, schema := range schemas {
			o.schemas = append(o.schemas, CredentialSchema{Type: "JsonSchema", ID: schema})
		}
	}
}

// Types can be used to add additional types to the Verifiable Credential created by [Create]
func Types(types ...string) CreateOption {
	return func(o *createOptions) {
		if o.types != nil {
			o.types = append(o.types, types...)
		} else {
			o.types = types
		}
	}
}

// ID can be used to override the default ID generated by [Create]
func ID(id string) CreateOption {
	return func(o *createOptions) {
		o.id = id
	}
}

// IssuanceDate can be used to override the default issuance date generated by [Create]
func IssuanceDate(issuanceDate time.Time) CreateOption {
	return func(o *createOptions) {
		o.issuanceDate = issuanceDate
	}
}

// ExpirationDate can be used to set the expiration date of the Verifiable Credential created by [Create]
func ExpirationDate(expirationDate time.Time) CreateOption {
	return func(o *createOptions) {
		o.expirationDate = expirationDate
	}
}

// Evidences can be used to set the evidence array of the Verifiable Credential created by [Create]
func Evidences(evidence ...Evidence) CreateOption {
	return func(o *createOptions) {
		o.evidence = evidence
	}
}

// Create returns a new Verifiable Credential with the provided claims and options.
// if no options are provided, the following defaults will be used:
//   - ID: urn:vc:uuid:<uuid>
//   - Contexts: ["https://www.w3.org/2018/credentials/v1"]
//   - Types: ["VerifiableCredential"]
//   - IssuanceDate: time.Now()
//
// # Note
//
// Any additional contexts or types provided will be appended to the defaults in order to remain conformant with
// the W3C Verifiable Credential Data Model specification
func Create[T CredentialSubject](claims T, opts ...CreateOption) DataModel[T] {
	o := createOptions{
		id:           "urn:vc:uuid:" + uuid.New().String(),
		contexts:     []string{BaseContext},
		types:        []string{BaseType},
		issuanceDate: time.Now(),
	}

	for _, f := range opts {
		f(&o)
	}

	cred := DataModel[T]{
		Context:           o.contexts,
		Type:              o.types,
		ID:                o.id,
		IssuanceDate:      o.issuanceDate.UTC().Format(time.RFC3339),
		CredentialSubject: claims,
		Evidence:          o.evidence,
	}

	if len(o.schemas) > 0 {
		cred.CredentialSchema = o.schemas
	}

	if (o.expirationDate != time.Time{}) {
		cred.ExpirationDate = o.expirationDate.UTC().Format(time.RFC3339)
	}

	return cred
}

// Sign returns a signed JWT conformant with the [vc-jwt] format. sets the provided vc as value of
// the "vc" claim in the jwt. It returns the signed jwt and an error if the signing fails.
//
// [vc-jwt]: https://www.w3.org/TR/vc-data-model/#json-web-token
func (vc DataModel[T]) Sign(bearerDID did.BearerDID, opts ...jwt.SignOpt) (string, error) {
	vc.Issuer = bearerDID.URI
	jwtClaims := jwt.Claims{
		Issuer:  vc.Issuer,
		JTI:     vc.ID,
		Subject: vc.CredentialSubject.GetID(),
	}

	t, err := time.Parse(time.RFC3339, vc.IssuanceDate)
	if err != nil {
		return "", fmt.Errorf("failed to parse issuance date: %w", err)
	}

	jwtClaims.NotBefore = t.Unix()

	if vc.ExpirationDate != "" {
		t, err := time.Parse(time.RFC3339, vc.ExpirationDate)
		if err != nil {
			return "", fmt.Errorf("failed to parse expiration date: %w", err)
		}

		jwtClaims.Expiration = t.Unix()
	}

	jwtClaims.Misc = make(map[string]any)
	jwtClaims.Misc["vc"] = vc

	// typ must be set to "JWT" as per the spec
	opts = append(opts, jwt.Type("JWT"))
	return jwt.Sign(jwtClaims, bearerDID, opts...)
}
