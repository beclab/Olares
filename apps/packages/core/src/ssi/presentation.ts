import { VerifiablePresentation, Container } from './credential';
import {
	PresentationSubmission,
	SubmissionRequirement,
	InputDescriptor,
	ClaimFormat,
	PresentationDefinition
} from './exchange';
import { JWT } from './keys';

// type CreatePresentationDefinitionRequest struct {
// 	Name                   string                           `json:"name,omitempty"`
// 	Purpose                string                           `json:"purpose,omitempty"`
// 	Format                 *exchange.ClaimFormat            `json:"format,omitempty" validate:"omitempty,dive"`
// 	InputDescriptors       []exchange.InputDescriptor       `json:"inputDescriptors" validate:"required,dive"`
// 	SubmissionRequirements []exchange.SubmissionRequirement `json:"submissionRequirements,omitempty" validate:"omitempty,dive"`

// 	// DID of the author of this presentation definition. The DID must have been previously created with the DID API,
// 	// or the PrivateKey must have been added independently. The privateKey associated to this DID will be used to
// 	// sign an envelope that contains the created presentation definition.
// 	Author string `json:"author" validate:"required"`
// }
export interface CreatePresentationDefinitionRequest {
	name?: string;
	purpose?: string;
	format?: ClaimFormat;
	inputDescriptors: InputDescriptor[];
	submissionRequirements?: SubmissionRequirement[];
	author: string;
}

// type CreatePresentationDefinitionResponse struct {
// 	PresentationDefinition exchange.PresentationDefinition `json:"presentation_definition"`

// 	// Signed envelope that contains the PresentationDefinition created using the privateKey of the author of the
// 	// definition.
// 	PresentationDefinitionJWT keyaccess.JWT `json:"presentationDefinitionJWT"`
// }
export interface CreatePresentationDefinitionResponse {
	presentation_definition: PresentationDefinition;
	presentationDefinitionJwt: JWT;
}

// type CreateSubmissionRequest struct {
// 	Presentation  credsdk.VerifiablePresentation  `json:"presentation" validate:"required"`
// 	SubmissionJWT keyaccess.JWT                   `json:"submissionJwt,omitempty" validate:"required"`
// 	Submission    exchange.PresentationSubmission `json:"submission" validate:"required"`
// 	Credentials   []credential.Container          `json:"credentials,omitempty"`
// }
export interface CreateSubmissionRequest {
	// presentation : VerifiablePresentation;
	// submissionJwt : JWT;
	// submission : PresentationSubmission;
	// credentials? : Container[];
	submissionJwt: JWT;
}

// type ListDefinitionsResponse struct {
// 	Definitions []*exchange.PresentationDefinition `json:"definitions"`
// }
export interface ListDefinitionsResponse {
	definitions: PresentationDefinition[];
}

export interface ReviewSubmissionRequest {
	approved: boolean;
	reason: string;
}

// type Submission struct {
// 	// One of {`pending`, `approved`, `denied`, `cancelled`}.
// 	Status string `json:"status" validate:"required"`
// 	// The reason why the submission was approved or denied.
// 	Reason string `json:"reason"`
// 	// The verifiable presentation containing the presentation_submission along with the credentials presented.
// 	VerifiablePresentation *credsdk.VerifiablePresentation `json:"verifiablePresentation,omitempty"`
// }

export interface Submission {
	status: string;
	reason: string;
	verifiablePresentation?: VerifiablePresentation;
}
