import { GetResponseResponse, PrivateJwk } from '@bytetrade/core';
import { stringToBase64 } from '@didvault/sdk/src/core';
import { getSubmitApplicationJWS } from './vc';
import { i18n } from '../boot/i18n';
import { VCCardInfo } from 'src/utils/vc';

export async function getChannelCredentialJWS(
	did: string,
	privateJWK: PrivateJwk,
	credentialSubject: any,
	schema: any
): Promise<string> {
	if (!credentialSubject) {
		throw Error(i18n.global.t('errors.get_credentialSubject_failure'));
	}

	if (!schema) {
		throw Error(i18n.global.t('errors.get_schema_failure'));
	}

	const jws = await getSubmitApplicationJWS(
		did!,
		privateJWK!,
		schema!.manifest,
		schema!.application_verifiable_credential.id,
		credentialSubject
	);
	return jws;
}

export async function submitChannelVCInfo(
	response: any,
	schema: any
): Promise<VCCardInfo> {
	if (!response || response.code != 0) {
		throw Error(i18n.global.t('errors.get_response_result_failure'));
	}

	if (!schema) {
		throw Error(i18n.global.t('errors.get_schema_failure'));
	}

	const manifest = stringToBase64(JSON.stringify(schema?.manifest));
	const result: GetResponseResponse = response.data;
	const verifiable_credential: string = result.verifiableCredentials![0];

	return { type: 'Channel', manifest, verifiable_credential };
}
