import { DataDisplay, EntityStyleDescriptor } from './render';
import {
	ClaimFormat,
	PresentationSubmission,
	PresentationDefinition,
	SubmissionDescriptor
} from './exchange';
import { JWT } from './keys';
import { Container } from './credential';
// OutputDescriptor https://identity.foundation/credential-manifest/#output-descriptor
export interface OutputDescriptor {
	// Must be unique within a manifest
	id: string;
	schema: string;
	name?: string;
	description?: string;
	// both below: an object or URI as defined by the DIF Entity Styles specification
	display?: DataDisplay;
	styles?: EntityStyleDescriptor;
}

export interface Issuer {
	id: string;
	name: string;
	styles: EntityStyleDescriptor;
}

// CreateManifestRequest is the request body for creating a manifest, which populates all remaining fields
// and builds a well-formed manifest object.
// type CreateManifestRequest struct {
// 	Name                   *string                          `json:"name,omitempty"`
// 	Description            *string                          `json:"description,omitempty"`
// 	IssuerDID              string                           `json:"issuerDid" validate:"required"`
// 	IssuerName             *string                          `json:"issuerName,omitempty"`
// 	ClaimFormat            *exchange.ClaimFormat            `json:"format" validate:"required,dive"`
// 	OutputDescriptors      []manifestsdk.OutputDescriptor   `json:"outputDescriptors" validate:"required,dive"`
// 	PresentationDefinition *exchange.PresentationDefinition `json:"presentationDefinition,omitempty" validate:"omitempty,dive"`
// }
export interface CreateManifestRequest {
	name: string;
	description: string;
	issuerDid: string;
	issuerName: string;
	verificationMethodId: string;
	format: ClaimFormat;
	outputDescriptors: OutputDescriptor[];
	presentationDefinitionId: string;
	presentationDefinition?: PresentationDefinition;
}

// CredentialManifest https://identity.foundation/credential-manifest/#general-composition
// type CredentialManifest struct {
// 	ID                     string                           `json:"id" validate:"required"`
// 	SpecVersion            string                           `json:"spec_version" validate:"required"`
// 	Name                   string                           `json:"name,omitempty"`
// 	Description            string                           `json:"description,omitempty"`
// 	Issuer                 Issuer                           `json:"issuer" validate:"required,dive"`
// 	OutputDescriptors      []OutputDescriptor               `json:"output_descriptors" validate:"required,dive"`
// 	Format                 *exchange.ClaimFormat            `json:"format,omitempty" validate:"omitempty,dive"`
// 	PresentationDefinition *exchange.PresentationDefinition `json:"presentation_definition,omitempty" validate:"omitempty,dive"`
// }

export interface CredentialManifest {
	id: string;
	specVersion: string;
	name: string;
	description: string;
	issuer: Issuer;
	output_descriptors: OutputDescriptor[];
	format: ClaimFormat;
	presentation_definition: PresentationDefinition;
}

// type CreateManifestResponse struct {
// 	Manifest    manifestsdk.CredentialManifest `json:"credential_manifest"`
// 	ManifestJWT keyaccess.JWT                  `json:"manifestJwt"`
// }
export interface CreateManifestResponse {
	credential_manifest: CredentialManifest;
	manifestJwt: JWT;
}

// type GetManifestResponse struct {
// 	ID          string                         `json:"id"`
// 	Manifest    manifestsdk.CredentialManifest `json:"credential_manifest"`
// 	ManifestJWT keyaccess.JWT                  `json:"manifestJwt"`
// }
export interface GetManifestResponse {
	id: string;
	credential_manifest: CredentialManifest;
	manifestJwt: JWT;
}

// type GetManifestsResponse struct {
// 	Manifests []GetManifestResponse `json:"manifests,omitempty"`
// }
export interface GetManifestsResponse {
	manifests: GetManifestResponse[];
}

// Application
// type SubmitApplicationRequest struct {
// 	ApplicantDID    string                            `json:"applicantDid" validate:"required"`
// 	Application     manifestsdk.CredentialApplication `json:"application" validate:"required"`
// 	Credentials     []cred.Container                  `json:"credentials,omitempty"`
// 	ApplicationJWT  keyaccess.JWT                     `json:"applicationJwt,omitempty" validate:"required"`
// 	ApplicationJSON map[string]any                    `json:"applicationJson,omitempty"`
// }
// export interface SubmitApplicationRequest {
//     applicantDid : string;
//     application : CredentialApplication;
//     credentials? : Container[];
//     applicationJwt? : JWT;
//     applicationJson? : any;
// }
export interface SubmitApplicationRequest {
	applicationJwt: JWT;
}

// CredentialApplication https://identity.foundation/credential-manifest/#credential-application
// type CredentialApplication struct {
// 	ID          string                `json:"id" validate:"required"`
// 	SpecVersion string                `json:"spec_version" validate:"required"`
// 	ManifestID  string                `json:"manifest_id" validate:"required"`
// 	Format      *exchange.ClaimFormat `json:"format" validate:"required,dive"`
// 	// Must be present if the corresponding manifest contains a presentation_definition
// 	PresentationSubmission *exchange.PresentationSubmission `json:"presentation_submission,omitempty" validate:"omitempty,dive"`
// }
export interface CredentialApplication {
	id: string;
	applicant: string;
	spec_version: string;
	manifest_id: string;
	format: ClaimFormat;
	presentation_submission?: PresentationSubmission;
}

export interface CredentialApplicationWrapper {
	credential_application: CredentialApplication;
	iss: string;
	verifiableCredentials: any[];
	// vcs : any[];
}

export interface Fulfillment {
	descriptor_map: SubmissionDescriptor[];
}

export interface Denial {
	reason: string;
	input_descriptors?: string[];
}

// CredentialResponse https://identity.foundation/credential-manifest/#credential-response
// type CredentialResponse struct {
// 	ID            string `json:"id" validate:"required"`
// 	SpecVersion   string `json:"spec_version" validate:"required"`
// 	ManifestID    string `json:"manifest_id" validate:"required"`
// 	ApplicationID string `json:"application_id"`
// 	Fulfillment   *struct {
// 		DescriptorMap []exchange.SubmissionDescriptor `json:"descriptor_map" validate:"required"`
// 	} `json:"fulfillment,omitempty" validate:"omitempty,dive"`
// 	Denial *struct {
// 		Reason           string   `json:"reason" validate:"required"`
// 		InputDescriptors []string `json:"input_descriptors,omitempty"`
// 	} `json:"denial,omitempty" validate:"omitempty,dive"`
// }
export interface CredentialResponse {
	id: string;
	spec_version: string;
	manifest_id: string;
	application_id: string;
	fulfillment?: Fulfillment;
	denial?: Denial;
}

// type GetResponseResponse struct {
// 	Response manifestsdk.CredentialResponse `json:"credential_response"`
// 	// this is an interface type to union Data Integrity and JWT style VCs
// 	Credentials any           `json:"verifiableCredentials,omitempty"`
// 	ResponseJWT keyaccess.JWT `json:"responseJwt,omitempty"`
// }

export interface GetResponseResponse {
	credential_response: CredentialResponse;
	verifiableCredentials?: any[];
	responseJwt?: JWT;
}

export interface GetApplicationsResponse {
	applications: CredentialApplication[];
}

export interface GetApplicationResponse {
	id: string;
	application: CredentialApplication;
}

export interface CredentialOverride {
	data: Record<string, any>;
	expiry: string;
	revocable: boolean;
}

export interface ReviewApplicationRequest {
	approved: boolean;
	reason: string;

	// Overrides to apply to the credentials that will be created. Keys are the ID that corresponds to an
	// OutputDescriptor.ID from the manifest.
	credentialOverrides?: Record<string, CredentialOverride>;
	//credential_overrides? : any;
}

export interface SubmitApplicationResponse {
	credential_response: CredentialResponse;
	verifiableCredentials?: any[];
	responseJwt?: JWT;
}
