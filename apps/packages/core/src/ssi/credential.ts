import { JWT } from './keys';
type Proof = any;
type CredentialSubject = Record<string, any>;

export interface CredentialSchema {
	id: string;
	type: string;
}

export interface RefreshService {
	id: string;
	type: string;
}

export interface Prohibition {
	assigner: string;
	assignee: string;
	target: string;
	action: string[];
}

export interface TermsOfUse {
	id?: string;
	type?: string;
	profile?: string;
	prohibition?: Prohibition[];
}

// VerifiableCredential is the verifiable credential model outlined in the
// vc-data-model spec https://www.w3.org/TR/2021/REC-vc-data-model-20211109/#basic-concepts
// type VerifiableCredential struct {
// 	// Either a string or set of strings
// 	Context interface{} `json:"@context" validate:"required"`
// 	ID      string      `json:"id,omitempty"`
// 	// Either a string or a set of strings https://www.w3.org/TR/2021/REC-vc-data-model-20211109/#types
// 	Type interface{} `json:"type" validate:"required"`
// 	// either a URI or an object containing an `id` property.
// 	Issuer interface{} `json:"issuer" validate:"required"`
// 	// https://www.w3.org/TR/xmlschema11-2/#dateTimes
// 	IssuanceDate     string      `json:"issuanceDate" validate:"required"`
// 	ExpirationDate   string      `json:"expirationDate,omitempty"`
// 	CredentialStatus interface{} `json:"credentialStatus,omitempty" validate:"omitempty,dive"`
// 	// This is where the subject's ID *may* be present
// 	CredentialSubject CredentialSubject `json:"credentialSubject" validate:"required"`
// 	CredentialSchema  *CredentialSchema `json:"credentialSchema,omitempty" validate:"omitempty,dive"`
// 	RefreshService    *RefreshService   `json:"refreshService,omitempty" validate:"omitempty,dive"`
// 	TermsOfUse        []TermsOfUse      `json:"termsOfUse,omitempty" validate:"omitempty,dive"`
// 	Evidence          []interface{}     `json:"evidence,omitempty" validate:"omitempty,dive"`
// 	// For embedded proof support
// 	// Proof is a digital signature over a credential https://www.w3.org/TR/2021/REC-vc-data-model-20211109/#proofs-signatures
// 	Proof *crypto.Proof `json:"proof,omitempty"`
// }
export interface VerifiableCredential {
	'@context': any;
	id: string;
	type: any;
	issuer: any;
	issuanceDate: string;
	expirationDate?: string;
	credentialStatus?: any;
	credentialSubject: CredentialSubject;
	credentialSchema?: CredentialSchema;
	refreshService?: RefreshService;
	termsOfUse?: TermsOfUse[];
	evidence?: any[];
	proof?: Proof;
}

// Container acts as an abstraction over both possible credential representations
// JWT representations are parsed upon container creation, while the original JWT is maintained
// type Container struct {
// 	// Credential ID
// 	ID            string
// 	Credential    *credential.VerifiableCredential
// 	CredentialJWT *keyaccess.JWT
// 	Revoked       bool
// 	Suspended     bool
// }
export interface Container {
	id: string;
	FullyQualifiedVerificationMethodID: string;
	credential: VerifiableCredential;
	credentialJwt: JWT;
	revoked: boolean;
	suspended: boolean;
}

export interface ListCredentialsResponse {
	credentials: Container[];
}

// VerifiablePresentation https://www.w3.org/TR/2021/REC-vc-data-model-20211109/#presentations-0
// type VerifiablePresentation struct {
// 	// Either a string or set of strings
// 	Context interface{} `json:"@context,omitempty"`
// 	ID      string      `json:"id,omitempty"`
// 	Holder  string      `json:"holder,omitempty"`
// 	Type    interface{} `json:"type" validate:"required"`
// 	// an optional field as a part of https://identity.foundation/presentation-exchange/#embed-targets
// 	PresentationSubmission interface{} `json:"presentation_submission,omitempty"`
// 	// Verifiable credential could be our object model, a JWT, or any other valid credential representation
// 	VerifiableCredential []interface{} `json:"verifiableCredential,omitempty"`
// 	Proof                *crypto.Proof `json:"proof,omitempty"`
// }
export interface VerifiablePresentation {
	'@context'?: any;
	id?: string;
	holder?: string;
	type: any;
	presentation_submission: any;
	verifiableCredential?: any[];
	proof?: Proof;
}
