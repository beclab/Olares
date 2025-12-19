import {
	CredentialApplicationWrapper,
	CredentialManifest,
	JwtPayload,
	PresentationDefinition,
	PrivateJwk,
	SignatureAlgorithm,
	VerifiableCredential,
	VerifiablePresentation,
	createVerifiableCredential,
	createVerifiablePresentation
} from '@bytetrade/core';
import { VaultItem, VaultType, uuid } from '@didvault/sdk/src/core';
import { i18n } from 'src/boot/i18n';
import { GeneralJwsSigner } from 'src/jose/jws/general/signer';
import { TermipassConfig } from './config';

export interface VCCardInfo {
	type: string;
	manifest: string;
	verifiable_credential: string;
}

export interface VCCardItem extends VCCardInfo {
	id: string;
}

export function convertVault2CardItem(item: VaultItem): VCCardItem | undefined {
	if (item.type != VaultType.VC) {
		return undefined;
	}
	return {
		id: item.id,
		type: item.fields[0].value,
		manifest: item.fields[1].value,
		verifiable_credential: item.fields[2].value
	};
}

export function getFullyQualifiedVerificationMethodID(did: string): string {
	const [, , id] = did.split(':');

	return did + '#' + id;
}

export async function signVerifiableCredentialJWT(
	cred: VerifiableCredential,
	privateKey: PrivateJwk
): Promise<string> {
	const t: JwtPayload = {};

	if (cred.expirationDate) {
		const unixTime = new Date(cred.expirationDate).getTime();
		if (!unixTime) {
			throw Error(
				i18n.global.t('errors.could_not_convert_expiration_date_to_unix_time')
			);
		}
		t.exp = Math.floor(unixTime / 1000);
	}
	t.iss = cred.issuer;
	t.nbf =
		Math.floor(new Date(cred.issuanceDate).getTime() / 1000) -
		TermipassConfig.jwt_payload_nbf_sub_second;
	t.jti = cred.id;
	t.sub = cred.credentialSubject.id;
	t.vc = cred;

	// const ecPrivateKey = await jose.importJWK(privateKey, 'ES256');

	// const jws = await new jose.CompactSign(
	// 	new TextEncoder().encode(JSON.stringify(t))
	// )
	// 	.setProtectedHeader({ alg: 'EdDSA', kid: cred.issuer, typ: 'JWT' })
	// 	.sign(ecPrivateKey);

	// return jws;
	const signer = await GeneralJwsSigner.create(
		new TextEncoder().encode(JSON.stringify(t)),
		[
			{
				privateJwk: privateKey,
				protectedHeader: {
					alg: 'EdDSA',
					kid: getFullyQualifiedVerificationMethodID(cred.issuer),
					iss: cred.issuer
				}
			}
		]
	);

	return signer.getJws();
}

export async function getSubmitApplicationJWS(
	holder: string,
	holderPrivateKey: PrivateJwk,
	manifest: CredentialManifest,
	schema_id: string,
	credentialSubject: any
): Promise<string> {
	const vc = await createVerifiableCredential(
		holder,
		schema_id,
		credentialSubject
	);
	const vc_jwt = await signVerifiableCredentialJWT(vc, holderPrivateKey);
	const application_id = await uuid();
	const descriptor_map: any[] = [];
	for (const input of manifest.presentation_definition.input_descriptors) {
		descriptor_map.push({
			id: input.id,
			format: 'jwt_vc',
			path: '$.verifiableCredentials[0]'
		});
	}

	const submit: CredentialApplicationWrapper = {
		credential_application: {
			id: application_id,
			spec_version:
				'https://identity.foundation/credential-manifest/spec/v1.0.0/',
			manifest_id: manifest.id,
			format: {
				jwt_vc: {
					alg: [SignatureAlgorithm.EdDSA]
				}
			},
			applicant: holder,
			presentation_submission: {
				id: 'psid',
				definition_id: manifest.presentation_definition.id,
				descriptor_map: descriptor_map
			}
		},
		iss: holder,
		verifiableCredentials: [vc_jwt]
	};
	const signer = await GeneralJwsSigner.create(
		new TextEncoder().encode(JSON.stringify(submit)),
		[
			{
				privateJwk: holderPrivateKey,
				protectedHeader: {
					alg: 'EdDSA',
					kid: getFullyQualifiedVerificationMethodID(holder),
					iss: holder
				}
			}
		]
	);
	return signer.getJws();
}

export async function getPresentationJWS(
	holder: string,
	holderPrivateKey: PrivateJwk,
	definition: PresentationDefinition,
	vc: string
) {
	const descriptor_map: any[] = [];
	for (const input of definition.input_descriptors) {
		descriptor_map.push({
			id: input.id,
			format: 'jwt_vc',
			path: '$.verifiableCredential[0]'
		});
	}

	const vp = await createVerifiablePresentation(
		holder,
		definition.id,
		vc,
		descriptor_map
	);

	return await signVerifiablePresentationJWT(vp, holderPrivateKey);
}

export async function signVerifiablePresentationJWT(
	pres: VerifiablePresentation,
	privateKey: PrivateJwk
): Promise<string> {
	const t: JwtPayload = {};

	t.jti = pres.id;
	t.sub = pres.holder;
	t.iss = getFullyQualifiedVerificationMethodID(pres.holder!);
	t.vp = pres;
	t.nonce = await uuid();

	return await signJwtPayload(t, privateKey);
}

export async function signJwtPayload(
	t: JwtPayload,
	holderPrivateKey: PrivateJwk
) {
	const signer = await GeneralJwsSigner.create(
		new TextEncoder().encode(JSON.stringify(t)),
		[
			{
				privateJwk: holderPrivateKey,
				protectedHeader: {
					alg: 'EdDSA',
					kid: getFullyQualifiedVerificationMethodID(t.sub!)
				}
			}
		]
	);
	return signer.getJws();
}
