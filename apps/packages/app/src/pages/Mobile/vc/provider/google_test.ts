import {
	PrivateJwk,
	GetResponseResponse,
	CredentialManifest
} from '@bytetrade/core';
import { stringToBase64 } from '@didvault/sdk/src/core';
import { useSSIStore } from '../../../../stores/ssi';
import { ClientSchema } from '../../../../globals';
import { i18n } from '../../../../boot/i18n';
import { VCCardInfo, getSubmitApplicationJWS } from 'src/utils/vc';

export interface aa {
	manifest: CredentialManifest;
}

export async function googleTestLogin(
	did: string,
	privateJWK: PrivateJwk,
	domain: string | null
): Promise<VCCardInfo> {
	const ssiStore = useSSIStore();
	const schema: ClientSchema | undefined =
		await ssiStore.get_application_schema('Google');
	if (!schema) {
		throw Error(i18n.global.t('errors.get_schema_failure'));
	}
	const manifest = stringToBase64(JSON.stringify(schema?.manifest));

	const jws = await getSubmitApplicationJWS(
		did,
		privateJWK,
		schema!.manifest,
		schema!.application_verifiable_credential.id,
		{ token: '1234' }
	);
	let obj: any = {
		jws: jws
	};
	if (domain) {
		obj = { jws, domain };
	}
	const response: any = await ssiStore
		.vcInstance()
		.post('/get_google_test_info/', obj);
	if (
		(response.status != 200 && response.status != 201) ||
		response.data.code != 0
	) {
		throw Error(
			response.data.message
				? response.data.message
				: i18n.global.t('errors.get_google_result_failure')
		);
	}
	const google_result: GetResponseResponse = response.data.data;
	const verifiable_credential: string = google_result.verifiableCredentials![0];

	return { type: 'Google', manifest, verifiable_credential };
}
