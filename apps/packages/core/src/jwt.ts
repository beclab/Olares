import {
	VerifiableCredential,
	VerifiablePresentation,
	PresentationSubmission
} from './ssi';
//import { uuid } from './util';
import { v4 as uuidv4 } from 'uuid';

export interface JwtPayload {
	[key: string]: any;
	iss?: string | undefined;
	sub?: string | undefined;
	aud?: string | string[] | undefined;
	exp?: number | undefined;
	nbf?: number | undefined;
	iat?: number | undefined;
	jti?: string | undefined;
}

export async function createVerifiableCredential(
	issuer: string,
	schemaID: string,
	credentialSubject: any
): Promise<VerifiableCredential> {
	//RFC 3339 format
	const issuanceDate = new Date(); //date.toISOString();
	const expirationDate = new Date(
		issuanceDate.getTime() + 1000 * 60 * 60 * 24 * 365
	);

	const subject = credentialSubject;
	subject.id = issuer;

	const vc: VerifiableCredential = {
		'@context': ['https://www.w3.org/2018/credentials/v1'],
		//id: await uuid(),
		id: uuidv4(),
		type: ['VerifiableCredential'],
		issuer: issuer,
		issuanceDate: issuanceDate.toISOString(),
		expirationDate: expirationDate.toISOString(),
		credentialSubject: subject,
		credentialSchema: {
			id: schemaID,
			//type: 'https://w3c-ccg.github.io/vc-json-schemas/schema/2.0/schema.json'
			type: 'JsonSchema2023'
		}
	};
	//console.log(vc);
	return vc;
}

export async function createVerifiablePresentation(
	holder: string,
	definition_id: string,
	vcData: string,
	descriptor_map: any
): Promise<VerifiablePresentation> {
	const ps: PresentationSubmission = {
		//id: await uuid(),
		id: uuidv4(),
		definition_id: definition_id,
		descriptor_map: descriptor_map
	};

	const vp: VerifiablePresentation = {
		'@context': ['https://www.w3.org/2018/credentials/v1'],
		//id: await uuid(),
		id: uuidv4(),
		type: ['VerifiablePresentation'],
		holder: holder,
		presentation_submission: ps,
		verifiableCredential: [vcData]
	};
	//console.log(vp);
	return vp;
}
