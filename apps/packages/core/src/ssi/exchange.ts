type SignatureType = string;

export enum JWTFormat {
	JWT = 'jwt',
	JWTVC = 'jwt_vc',
	JWTVP = 'jwt_vp'
}

export enum LinkedDataFormat {
	LDP = 'ldp',
	LDPVC = 'ldp_vc',
	LDPVP = 'ldp_vp'
}

export enum Selection {
	All = 'all',
	Pick = 'pick'
}

export interface JWTType {
	alg: SignatureAlgorithm[] | string[];
}

export enum SignatureAlgorithm {
	// EdDSA uses an ed25519 key
	EdDSA = 'EdDSA',
	// ES256K uses a secp256k1 key
	ES256K = 'ES256K',
	// ES256 uses a p-256 curve key
	ES256 = 'ES256',
	// ES384 uses a p-384 curve key
	ES384 = 'ES384',
	// PS256 uses a 2048-bit RSA key
	PS256 = 'PS256'
}

// Used for limiting disclosure, predicates, and relational constraints
export enum Preference {
	Required = 'required',
	Preferred = 'preferred',
	Allowed = 'allowed',
	Disallowed = 'disallowed'
}

export interface LDPType {
	proof_type: SignatureType[];
}

// ClaimFormat https://identity.foundation/presentation-exchange/#claim-format-designations
// At most one field can have non-nil
export interface ClaimFormat {
	jwt?: JWTType;
	jwt_vc?: JWTType;
	jwt_vp?: JWTType;

	ldp?: LDPType;
	ldp_vc?: LDPType;
	ldp_vp?: LDPType;
}

// SubmissionRequirement https://identity.foundation/presentation-exchange/#presentation-definition-extensions
// type SubmissionRequirement struct {
// 	Rule Selection `json:"rule" validate:"required"`
// 	// Either an array of SubmissionRequirement OR a string value
// 	FromOption `validate:"required"`

// 	Name    string `json:"name,omitempty"`
// 	Purpose string `json:"purpose,omitempty"`
// 	Count   int    `json:"count,omitempty" validate:"omitempty,min=1"`
// 	Minimum int    `json:"min,omitempty"`
// 	Maximum int    `json:"max,omitempty"`
// }
export interface SubmissionRequirement {
	rule: Selection;
	from: string;
	from_nested: SubmissionRequirement[];
	name: string;
	purpose: string;
	count: number;
	minimum: number;
	maximum: number;
}

// type InputDescriptor struct {
// 	// Must be unique within the Presentation Definition
// 	ID   string `json:"id" validate:"required"`
// 	Name string `json:"name,omitempty"`
// 	// Purpose for which claim's data is being requested
// 	Purpose     string       `json:"purpose,omitempty"`
// 	Format      *ClaimFormat `json:"format,omitempty" validate:"omitempty,dive"`
// 	Constraints *Constraints `json:"constraints" validate:"required"`
// 	// Must match a grouping strings listed in the `from` values of a submission requirement rule
// 	Group []string `json:"group,omitempty"`
// }
export interface InputDescriptor {
	id: string;
	name?: string;
	purpose?: string;
	format?: ClaimFormat;
	constraints: Constraints;
	group?: string[];
}

// type Field struct {
// 	ID             string   `json:"id,omitempty"`
// 	Name           string   `json:"name,omitempty"`
// 	Path           []string `json:"path,omitempty" validate:"required"`
// 	Purpose        string   `json:"purpose,omitempty"`
// 	Optional       bool     `json:"optional,omitempty"`
// 	IntentToRetain bool     `json:"intent_to_retain,omitempty"`
// 	// If a predicate property is present, filter must be too
// 	// https://identity.foundation/presentation-exchange/#predicate-feature
// 	Predicate *Preference `json:"predicate,omitempty"`
// 	Filter    *Filter     `json:"filter,omitempty"`
// }
export interface ExchangeField {
	id?: string;
	name?: string;
	path: string[];
	purpose?: string;
	optional?: boolean;
	intent_to_retain?: boolean;
	predicate?: Preference;
	filter?: Filter;
}

// type RelationalConstraint struct {
// 	FieldID   string      `json:"field_id" validate:"required"`
// 	Directive *Preference `json:"directive" validate:"required"`
// }
export interface RelationalConstraint {
	fieldId: string;
	directive: Preference;
}

// type Filter struct {
// 	Type                 string        `json:"type,omitempty"`
// 	Format               string        `json:"format,omitempty"`
// 	Properties           interface{}   `json:"properties,omitempty"`
// 	Required             []string      `json:"required,omitempty"`
// 	AdditionalProperties bool          `json:"additionalProperties,omitempty"`
// 	Pattern              string        `json:"pattern,omitempty"`
// 	Minimum              interface{}   `json:"minimum,omitempty"`
// 	Maximum              interface{}   `json:"maximum,omitempty"`
// 	MinLength            int           `json:"minLength,omitempty"`
// 	MaxLength            int           `json:"maxLength,omitempty"`
// 	ExclusiveMinimum     interface{}   `json:"exclusiveMinimum,omitempty"`
// 	ExclusiveMaximum     interface{}   `json:"exclusiveMaximum,omitempty"`
// 	Const                interface{}   `json:"const,omitempty"`
// 	Enum                 []interface{} `json:"enum,omitempty"`
// 	Not                  interface{}   `json:"not,omitempty"`
// 	AllOf                interface{}   `json:"allOf,omitempty"`
// 	OneOf                interface{}   `json:"oneOf,omitempty"`
// }
export interface Filter {
	type: string;
	format: string;
	properties: any;
	required: string[];
	additionalProperties: boolean;
	pattern: string;
	minimum: any;
	maximum: any;
	minLength: number;
	maxLength: number;
	exclusiveMaximum: any;
	const: any;
	enum: any[];
	not: any;
	allOf: any;
	oneOf: any;
}

// type Constraints struct {
// 	Fields          []Field     `json:"fields,omitempty" validate:"omitempty,dive"`
// 	LimitDisclosure *Preference `json:"limit_disclosure,omitempty"`

// 	// https://identity.foundation/presentation-exchange/#relational-constraint-feature
// 	SubjectIsIssuer *Preference           `json:"subject_is_issuer,omitempty"`
// 	IsHolder        *RelationalConstraint `json:"is_holder,omitempty" validate:"omitempty,dive"`
// 	SameSubject     *RelationalConstraint `json:"same_subject,omitempty"`

// 	// https://identity.foundation/presentation-exchange/#credential-status-constraint-feature
// 	Statuses *CredentialStatus `json:"statuses,omitempty"`
// }
export interface Constraints {
	fields?: ExchangeField[];
	limit_disclosure?: Preference | string;
	subject_is_issuer?: Preference | string;
	is_holder?: RelationalConstraint;
	same_subject?: RelationalConstraint;
	statuses?: CredentialStatus;
}

// CredentialStatus https://identity.foundation/presentation-exchange/#credential-status-constraint-feature
// type CredentialStatus struct {
// 	Active *struct {
// 		Directive Preference `json:"directive,omitempty"`
// 	} `json:"active,omitempty"`

// 	Suspended *struct {
// 		Directive Preference `json:"directive,omitempty"`
// 	} `json:"suspended,omitempty"`

// 	Revoked *struct {
// 		Directive Preference `json:"directive,omitempty"`
// 	} `json:"revoked,omitempty"`
// }

export interface CredentialStatus {
	active: {
		directive: Preference;
	};
	suspended: {
		directive: Preference;
	};
	revoked: {
		directive: Preference;
	};
}

export interface SubmissionDescriptor {
	id: string;
	format: string;
	path: string;
	path_nested?: SubmissionDescriptor;
}

export interface PresentationSubmission {
	id: string;
	definition_id: string;
	descriptor_map: SubmissionDescriptor[];
}

// // PresentationDefinition https://identity.foundation/presentation-exchange/#presentation-definition
// type PresentationDefinition struct {
// 	ID                     string                  `json:"id,omitempty" validate:"required"`
// 	Name                   string                  `json:"name,omitempty"`
// 	Purpose                string                  `json:"purpose,omitempty"`
// 	Format                 *ClaimFormat            `json:"format,omitempty" validate:"omitempty,dive"`
// 	InputDescriptors       []InputDescriptor       `json:"input_descriptors" validate:"required,dive"`
// 	SubmissionRequirements []SubmissionRequirement `json:"submission_requirements,omitempty" validate:"omitempty,dive"`

// 	// https://identity.foundation/presentation-exchange/#json-ld-framing-feature
// 	Frame interface{} `json:"frame,omitempty"`
// }
export interface PresentationDefinition {
	id: string;
	name?: string;
	purpose?: string;
	format?: ClaimFormat;
	input_descriptors: InputDescriptor[];
	submission_requirements?: SubmissionRequirement[];
	frame?: any;
}
