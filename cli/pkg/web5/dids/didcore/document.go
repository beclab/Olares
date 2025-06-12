package didcore

import (
	"errors"
	"fmt"

	"github.com/beclab/Olares/cli/pkg/web5/jwk"
)

const (
	PurposeAssertion            Purpose = "assertionMethod"
	PurposeAuthentication       Purpose = "authentication"
	PurposeCapabilityDelegation Purpose = "capabilityDelegation"
	PurposeCapabilityInvocation Purpose = "capabilityInvocation"
	PurposeKeyAgreement         Purpose = "keyAgreement"
)

// Document represents a set of data describing the DID subject including mechanisms such as:
//   - cryptographic public keys - used to authenticate itself and prove
//     association with the DID
//   - services - means of communicating or interacting with the DID subject or
//     associated entities via one or more service endpoints.
//     Examples include discovery services, agent services,
//     social networking services, file storage services,
//     and verifiable credential repository services.
//
// A DID Document can be retrieved by resolving a DID URI.
type Document struct {
	// Context is a URI that defines the schema version used in the document.
	Context []string `json:"@context,omitempty"`

	// Id is the DID URI for a particular DID subject, expressed using the id property in the DID document.
	ID string `json:"id"`

	// AlsoKnownAs can contain multiple identifiers for different purposes, or at different times for the same DID subject.
	// The assertion that two or more DIDs (or other types of URI) refer to the same DID subject can be made using the alsoKnownAs property.
	AlsoKnownAs []string `json:"alsoKnownAs,omitempty"`

	// Controller defines an entity that is authorized to make changes to a DID document.
	// The process of authorizing a DID controller is defined by the DID method.
	// It can be a string or a list of strings.
	Controller []string `json:"controller,omitempty"`

	// VerificationMethod is a list of cryptographic public keys, which can be used to authenticate or authorize
	// interactions with the DID subject or associated parties.
	VerificationMethod []VerificationMethod `json:"verificationMethod,omitempty"`

	// Service expresses ways of communicating with the DID subject or associated entities.
	// A service can be any type of service the DID subject wants to advertise.
	// spec reference: https://www.w3.org/TR/did-core/#verification-methods
	Service []Service `json:"service,omitempty"`

	// AssertionMethod is used to specify how the DID subject is expected to express claims,
	// such as for the purposes of issuing a Verifiable Credential.
	AssertionMethod []string `json:"assertionMethod,omitempty"`

	// Authentication specifies how the DID subject is expected to be authenticated,
	// for purposes such as logging into a website or engaging in any sort of challenge-response protocol.
	Authentication []string `json:"authentication,omitempty"`

	// KeyAgreement specifies how an entity can generate encryption material to transmit confidential
	// information intended for the DID subject, such as for establishing a secure communication channel.
	KeyAgreement []string `json:"keyAgreement,omitempty"`

	// CapabilityDelegation specifies a mechanism used by the DID subject to delegate a
	// cryptographic capability to another party, such as delegating the authority to access a specific HTTP API.
	CapabilityDelegation []string `json:"capabilityDelegation,omitempty"`

	// CapabilityInvocation specifies a verification method used by the DID subject to invoke a
	// cryptographic capability, such as the authorization to update the DID Document.
	CapabilityInvocation []string `json:"capabilityInvocation,omitempty"`
}

type addVMOptions struct {
	purposes []Purpose
}

// AddVMOption is a type returned by all AddVerificationMethod options for variadic parameter support
type AddVMOption func(o *addVMOptions)

// Purposes can be used to select a verification method with a specific purpose.
func Purposes(p ...Purpose) AddVMOption {
	return func(o *addVMOptions) {
		o.purposes = p
	}
}

// AddVerificationMethod adds a verification method to the document. if Purposes are provided,
// the verification method's ID will be added to the corresponding list of purposes.
func (d *Document) AddVerificationMethod(method VerificationMethod, opts ...AddVMOption) {
	o := &addVMOptions{purposes: []Purpose{}}
	for _, opt := range opts {
		opt(o)
	}

	d.VerificationMethod = append(d.VerificationMethod, method)

	for _, p := range o.purposes {
		switch p {
		case PurposeAssertion:
			d.AssertionMethod = append(d.AssertionMethod, method.ID)
		case PurposeAuthentication:
			d.Authentication = append(d.Authentication, method.ID)
		case PurposeKeyAgreement:
			d.KeyAgreement = append(d.KeyAgreement, method.ID)
		case PurposeCapabilityDelegation:
			d.CapabilityDelegation = append(d.CapabilityDelegation, method.ID)
		case PurposeCapabilityInvocation:
			d.CapabilityInvocation = append(d.CapabilityInvocation, method.ID)
		}
	}
}

// VMSelector is an interface that can be implemented to provide a means to select
// a specific verification method from a DID Document.
type VMSelector interface {
	selector()
}

// Purpose can be used to select a verification method with a specific purpose.
type Purpose string

func (p Purpose) selector() {}

// ID can be used to select a verification method by its ID.
type ID string

func (i ID) selector() {}

// SelectVerificationMethod takes a selector that can be used to select a specific verification
// method from the DID Document. If a nil selector is provided, the first verification method
// is returned
//
// The selector can either be an ID, Purpose, or nil. If a Purpose is provided, the first verification
// method in the DID Document that has the provided purpose will be returned.
func (d *Document) SelectVerificationMethod(selector VMSelector) (VerificationMethod, error) {
	if len(d.VerificationMethod) == 0 {
		return VerificationMethod{}, errors.New("no verification methods found")
	}

	if selector == nil {
		return d.VerificationMethod[0], nil
	}

	var vmID string
	switch s := selector.(type) {
	case Purpose:
		switch s {
		case PurposeAssertion:
			if len(d.AssertionMethod) == 0 {
				return VerificationMethod{}, fmt.Errorf("no verification method found for purpose: %s", s)
			}

			vmID = d.AssertionMethod[0]
		case PurposeAuthentication:
			if len(d.Authentication) == 0 {
				return VerificationMethod{}, fmt.Errorf("no %s verification method found", s)
			}

			vmID = d.Authentication[0]
		case PurposeCapabilityDelegation:
			if len(d.CapabilityDelegation) == 0 {
				return VerificationMethod{}, fmt.Errorf("no %s verification method found", s)
			}

			vmID = d.CapabilityDelegation[0]
		case PurposeCapabilityInvocation:
			if len(d.CapabilityInvocation) == 0 {
				return VerificationMethod{}, fmt.Errorf("no %s verification method found", s)
			}

			vmID = d.CapabilityInvocation[0]
		case PurposeKeyAgreement:
			if len(d.KeyAgreement) == 0 {
				return VerificationMethod{}, fmt.Errorf("no %s verification method found", s)
			}

			vmID = d.KeyAgreement[0]
		default:
			return VerificationMethod{}, fmt.Errorf("unsupported purpose: %s", s)
		}
	case ID:
		vmID = string(s)
	}

	for _, vm := range d.VerificationMethod {
		if vm.ID == vmID {
			return vm, nil
		}
	}

	return VerificationMethod{}, fmt.Errorf("no verification method found for id: %s", vmID)
}

// AddService will append the given Service to the Document.Services array
func (d *Document) AddService(service Service) {
	d.Service = append(d.Service, service)
}

// GetAbsoluteResourceID returns a fully qualified ID for a document resource (e.g. service, verification method)
// Document Resource IDs are allowed to be relative DID URLs as a means to reduce storage size of DID Documents.
// More info here: https://www.w3.org/TR/did-core/#relative-did-urls
func (d *Document) GetAbsoluteResourceID(id string) string {
	if id[0] == '#' {
		return d.ID + id
	}

	return id
}

// DocumentMetadata contains metadata about the DID Document
// This metadata typically does not change between invocations of
// the resolve and resolveRepresentation functions unless the DID document
// changes, as it represents metadata about the DID document.
//
// Spec: https://www.w3.org/TR/did-core/#dfn-diddocumentmetadata
type DocumentMetadata struct {
	// timestamp of the Create operation. The value of the property MUST be a
	// string formatted as an XML Datetime normalized to UTC 00:00:00 and
	// without sub-second decimal precision. For example: 2020-12-20T19:17:47Z.
	Created string `json:"created,omitempty"`
	// timestamp of the last Update operation for the document version which was
	// resolved. The value of the property MUST follow the same formatting rules
	// as the created property. The updated property is omitted if an Update
	// operation has never been performed on the DID document. If an updated
	// property exists, it can be the same value as the created property
	// when the difference between the two timestamps is less than one second.
	Updated string `json:"updated,omitempty"`
	// If a DID has been deactivated, DID document metadata MUST include this
	// property with the boolean value true. If a DID has not been deactivated,
	// this property is OPTIONAL, but if included, MUST have the boolean value
	// false.
	Deactivated bool `json:"deactivated,omitempty"`
	// indicates the version of the last Update operation for the document version
	// which was resolved.
	VersionID string `json:"versionId,omitempty"`
	// indicates the timestamp of the next Update operation. The value of the
	// property MUST follow the same formatting rules as the created property.
	NextUpdate string `json:"nextUpdate,omitempty"`
	// if the resolved document version is not the latest version of the document.
	// It indicates the timestamp of the next Update operation. The value of the
	// property MUST follow the same formatting rules as the created property.
	NextVersionID string `json:"nextVersionId,omitempty"`
	// A DID method can define different forms of a DID that are logically
	// equivalent. An example is when a DID takes one form prior to registration
	// in a verifiable data registry and another form after such registration.
	// In this case, the DID method specification might need to express one or
	// more DIDs that are logically equivalent to the resolved DID as a property
	// of the DID document. This is the purpose of the equivalentId property.
	EquivalentID []string `json:"equivalentId,omitempty"`
	// The canonicalId property is identical to the equivalentId property except:
	//   * it is associated with a single value rather than a set
	//   * the DID is defined to be the canonical ID for the DID subject within
	//     the scope of the containing DID document.
	CanonicalID string `json:"canonicalId,omitempty"`
}

// Service is used in DID documents to express ways of communicating with
// the DID subject or associated entities.
// A service can be any type of service the DID subject wants to advertise.
//
// Specification Reference: https://www.w3.org/TR/did-core/#services
type Service struct {
	// Id is the value of the id property and MUST be a URI conforming to RFC3986.
	// A conforming producer MUST NOT produce multiple service entries with
	// the same id. A conforming consumer MUST produce an error if it detects
	// multiple service entries with the same id.
	ID string `json:"id"`

	// Type is an example of registered types which can be found
	// here: https://www.w3.org/TR/did-spec-registries/#service-types
	Type string `json:"type"`

	// ServiceEndpoint is a network address, such as an HTTP URL, at which services
	// operate on behalf of a DID subject.
	ServiceEndpoint []string `json:"serviceEndpoint"`
}

// VerificationMethod expresses verification methods, such as cryptographic
// public keys, which can be used to authenticate or authorize interactions
// with the DID subject or associated parties. For example,
// a cryptographic public key can be used as a verification method with
// respect to a digital signature; in such usage, it verifies that the
// signer could use the associated cryptographic private key.
//
// Specification Reference: https://www.w3.org/TR/did-core/#verification-methods
type VerificationMethod struct {
	ID string `json:"id"`
	// references exactly one verification method type. In order to maximize global
	// interoperability, the verification method type SHOULD be registered in the
	// DID Specification Registries: https://www.w3.org/TR/did-spec-registries/
	Type string `json:"type"`
	// a value that conforms to the rules in DID Syntax: https://www.w3.org/TR/did-core/#did-syntax
	Controller string `json:"controller"`
	// specification reference: https://www.w3.org/TR/did-core/#dfn-publickeyjwk
	PublicKeyJwk *jwk.JWK `json:"publicKeyJwk,omitempty"`
}
