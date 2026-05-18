export enum LDKeyType {
	X25519KeyAgreementKey2020 = 'X25519KeyAgreementKey2020',
	Ed25519VerificationKey2020 = 'Ed25519VerificationKey2020',
	X25519KeyAgreementKey2019 = 'X25519KeyAgreementKey2019',
	Ed25519VerificationKey2018 = 'Ed25519VerificationKey2018',
	EcdsaSecp256k1VerificationKey2019 = 'EcdsaSecp256k1VerificationKey2019'
}

/**
 * Contains a public-private key pair and the associated key ID.
 */
export type KeyMaterial = {
	keyId: string;
	keyPair: { publicJwk: PublicJwk; privateJwk: PrivateJwk };
};

export type Jwk = {
	/** The "alg" (algorithm) parameter identifies the algorithm intended for use with the key. */
	alg?: string;
	/** The "alg" (algorithm) parameter identifies the algorithm intended for use with the key. */
	kid?: string;
	/** identifies the cryptographic algorithm family used with the key, such "EC". */
	kty: string;

	use?: string;
};

export type PublicJwk = Jwk & {
	/** The "crv" (curve) parameter identifies the cryptographic curve used with the key.
	 * MUST be present for all EC public keys
	 */
	crv: string;
	/**
	 * the x coordinate for the Elliptic Curve point.
	 * Represented as the base64url encoding of the octet string representation of the coordinate.
	 * MUST be present for all EC public keys
	 */
	x: string;
	/**
	 * the y coordinate for the Elliptic Curve point.
	 * Represented as the base64url encoding of the octet string representation of the coordinate.
	 */
	y?: string;
};

export type PrivateJwk = PublicJwk & {
	/**
	 * the Elliptic Curve private key value.
	 * It is represented as the base64url encoding of the octet string representation of the private key value
	 * MUST be present to represent Elliptic Curve private keys.
	 */
	d: string;
};

// PublicKeyJWK complies with RFC7517 https://datatracker.ietf.org/doc/html/rfc7517
// type PublicKeyJWK struct {
// 	KTY    string `json:"kty" validate:"required"`
// 	CRV    string `json:"crv,omitempty"`
// 	X      string `json:"x,omitempty"`
// 	Y      string `json:"y,omitempty"`
// 	N      string `json:"n,omitempty"`
// 	E      string `json:"e,omitempty"`
// 	Use    string `json:"use,omitempty"`
// 	KeyOps string `json:"key_ops,omitempty"`
// 	Alg    string `json:"alg,omitempty"`
// 	KID    string `json:"kid,omitempty"`
// }
// export interface PublicKeyJWK {
//   kty: string;
//   crv?: string;
//   x?: string;
//   y?: string;
//   n?: string;
//   e?: string;
//   use?: string;
//   key_ops?: string;
//   alg?: string;
//   kid?: string;
// }

// DIDDocumentMetadata https://www.w3.org/TR/did-core/#did-document-metadata
// type DIDDocumentMetadata struct {
// 	Created       string `json:"created,omitempty" validate:"datetime"`
// 	Updated       string `json:"updated,omitempty" validate:"datetime"`
// 	Deactivated   bool   `json:"deactivated,omitempty"`
// 	NextUpdate    string `json:"nextUpdate,omitempty"`
// 	VersionID     string `json:"versionId,omitempty"`
// 	NextVersionID string `json:"nextVersionId,omitempty"`
// 	EquivalentID  string `json:"equivalentId,omitempty"`
// 	CanonicalID   string `json:"canonicalId,omitempty"`
// }
export interface DIDDocumentMetadata {
	created?: string;
	updated?: string;
	deactivated?: boolean;
	nextUpdate?: string;
	versionId?: string;
	nextVersionId?: string;
	equivalentId?: string;
	canonicalId?: string;
}

// type VerificationMethod struct {
// 	ID              string                `json:"id" validate:"required"`
// 	Type            cryptosuite.LDKeyType `json:"type" validate:"required"`
// 	Controller      string                `json:"controller" validate:"required"`
// 	PublicKeyBase58 string                `json:"publicKeyBase58,omitempty"`
// 	// must conform to https://datatracker.ietf.org/doc/html/rfc7517
// 	PublicKeyJWK *crypto.PublicKeyJWK `json:"publicKeyJwk,omitempty" validate:"omitempty,dive"`
// 	// https://datatracker.ietf.org/doc/html/draft-multiformats-multibase-03
// 	PublicKeyMultibase string `json:"publicKeyMultibase,omitempty"`
// 	// for PKH DIDs - https://github.com/w3c-ccg/did-pkh/blob/90b28ad3c18d63822a8aab3c752302aa64fc9382/did-pkh-method-draft.md
// 	BlockchainAccountID string `json:"blockchainAccountId,omitempty"`
// }
export interface VerificationMethod {
	id: string;
	type: LDKeyType;
	controller: string;
	publicKeyBase58?: string;
	publicKeyJwk?: PublicJwk;
	publicKeyMultibase?: string;
	blockchainAccountId?: string;
}

export type VerificationMethodSet = any;

// Service is a property compliant with the did-core spec https://www.w3.org/TR/did-core/#services
// type Service struct {
// 	ID   string `json:"id" validate:"required"`
// 	Type string `json:"type" validate:"required"`
// 	// A string, map, or set composed of one or more strings and/or maps
// 	// All string values must be valid URIs
// 	ServiceEndpoint interface{} `json:"serviceEndpoint" validate:"required"`
// 	RoutingKeys     []string    `json:"routingKeys,omitempty"`
// 	Accept          []string    `json:"accept,omitempty"`
// }
export interface Service {
	id: string;
	type: string;
	serviceEndpoint: any;
	routingKeys?: string[];
	accept?: string[];
}

// DIDDocument is a representation of the did core specification https://www.w3.org/TR/did-core
// TODO(gabe) enforce validation of DID syntax https://www.w3.org/TR/did-core/#did-syntax
// type DIDDocument struct {
// 	Context interface{} `json:"@context,omitempty"`
// 	// As per https://www.w3.org/TR/did-core/#did-subject intermediate representations of DID Documents do not
// 	// require an ID property. The provided test vectors demonstrate IRs. As such, the property is optional.
// 	ID                   string                  `json:"id,omitempty"`
// 	Controller           string                  `json:"controller,omitempty"`
// 	AlsoKnownAs          string                  `json:"alsoKnownAs,omitempty"`
// 	VerificationMethod   []VerificationMethod    `json:"verificationMethod,omitempty" validate:"dive"`
// 	Authentication       []VerificationMethodSet `json:"authentication,omitempty" validate:"dive"`
// 	AssertionMethod      []VerificationMethodSet `json:"assertionMethod,omitempty" validate:"dive"`
// 	KeyAgreement         []VerificationMethodSet `json:"keyAgreement,omitempty" validate:"dive"`
// 	CapabilityInvocation []VerificationMethodSet `json:"capabilityInvocation,omitempty" validate:"dive"`
// 	CapabilityDelegation []VerificationMethodSet `json:"capabilityDelegation,omitempty" validate:"dive"`
// 	Services             []Service               `json:"service,omitempty" validate:"dive"`
// }
export interface DIDDocument {
	'@context'?: any;
	id?: string;
	controller?: string;
	alsoKnownAs?: string;
	verificationMethod?: VerificationMethod[];
	authentication?: VerificationMethodSet[];
	assertionMethod?: VerificationMethodSet[];
	keyAgreement?: VerificationMethodSet[];
	capabilityInvocation?: VerificationMethodSet[];
	capabilityDelegation?: VerificationMethodSet[];
	services?: Service[];
}

// ResolutionError https://www.w3.org/TR/did-core/#did-resolution-metadata
// type ResolutionError struct {
// 	Code                       string `json:"code"`
// 	InvalidDID                 bool   `json:"invalidDid"`
// 	NotFound                   bool   `json:"notFound"`
// 	RepresentationNotSupported bool   `json:"representationNotSupported"`
// }
export interface ResolutionError {
	code: string;
	invalidDid?: boolean;
	notFound?: boolean;
	representationNotSupported?: boolean;
}

// DIDResolutionMetadata https://www.w3.org/TR/did-core/#did-resolution-metadata
// type DIDResolutionMetadata struct {
// 	ContentType string
// 	Error       *ResolutionError
// }
export interface DIDResolutionMetadata {
	contentType?: string;
	error?: ResolutionError;
}

// type ResolutionResult struct {
// 	Context                      any `json:"@context"`
// 	didsdk.DIDResolutionMetadata `json:"didResolutionMetadata"`
// 	didsdk.DIDDocument           `json:"didDocument"`
// 	didsdk.DIDDocumentMetadata   `json:"didDocumentMetadata"`
// }
export interface ResolutionResult {
	'@context'?: any;
	didResolutionMetadata: DIDResolutionMetadata;
	didDocument: DIDDocument;
	didDocumentMetadata: DIDDocumentMetadata;
}

// export type DidDocument = {
//   '@context'?: 'https://www.w3.org/ns/did/v1' | string | string[];
//   id: string;
//   alsoKnownAs?: string[];
//   controller?: string | string[];
//   verificationMethod?: VerificationMethod[];
//   service?: ServiceEndpoint[];
//   authentication?: VerificationMethod[] | string[];
//   assertionMethod?: VerificationMethod[] | string[];
//   keyAgreement?: VerificationMethod[] | string[];
//   capabilityInvocation?: VerificationMethod[] | string[];
//   capabilityDelegation?: VerificationMethod[] | string[];
// };

export type DwnServiceEndpoint = {
	nodes: string[];
};

export type ServiceEndpoint = {
	id: string;
	type: string;
	serviceEndpoint: string | DwnServiceEndpoint;
	description?: string;
};

// export type DidResolutionResult = {
//   '@context'?: 'https://w3id.org/did-resolution/v1' | string | string[];
//   didResolutionMetadata: DidResolutionMetadata;
//   didDocument?: DidDocument;
//   didDocumentMetadata: DidDocumentMetadata;
// };

export type DidResolutionMetadata = {
	contentType?: string;
	error?:
		| 'invalidDid'
		| 'notFound'
		| 'representationNotSupported'
		| 'unsupportedDidMethod'
		| string;
};

export type DidDocumentMetadata = {
	// indicates the timestamp of the Create operation. ISO8601 timestamp
	created?: string;
	// indicates the timestamp of the last Update operation for the document version which was
	// resolved. ISO8601 timestamp
	updated?: string;
	// indicates whether the DID has been deactivated
	deactivated?: boolean;
	// indicates the version of the last Update operation for the document version which
	// was resolved
	versionId?: string;
	// indicates the timestamp of the next Update operation if the resolved document version
	// is not the latest version of the document.
	nextUpdate?: string;
	// indicates the version of the next Update operation if the resolved document version
	// is not the latest version of the document.
	nextVersionId?: string;
	// @see https://www.w3.org/TR/did-core/#dfn-equivalentid
	equivalentId?: string;
	// @see https://www.w3.org/TR/did-core/#dfn-canonicalid
	canonicalId?: string;
};
