import { PrivateJwk, GetResponseResponse } from '@bytetrade/core';
import { stringToBase64 } from '@didvault/sdk/src/core';
import { useSSIStore } from '../../../../stores/ssi';
import { useCloudStore } from '../../../../stores/cloud';
import { ClientSchema } from '../../../../globals';
import { i18n } from '../../../../boot/i18n';
import { VCCardInfo, getSubmitApplicationJWS } from 'src/utils/vc';

export async function getDomainVC(
	owner: string,
	did: string,
	domain: string,
	cloudid: string,
	privateJWK: PrivateJwk
): Promise<VCCardInfo> {
	const ssiStore = useSSIStore();
	const cloudStore = useCloudStore();
	const schema: ClientSchema | undefined =
		await ssiStore.get_application_schema('Domain');
	if (!schema) {
		throw Error(i18n.global.t('errors.get_schema_failure'));
	}
	const manifest = stringToBase64(JSON.stringify(schema?.manifest));

	const jws = await getSubmitApplicationJWS(
		did,
		privateJWK,
		schema!.manifest,
		schema!.application_verifiable_credential.id,
		{ owner, did, domain, cloudid }
	);

	const data: any = await cloudStore.requestDomainVC(jws, domain);

	const google_result: GetResponseResponse = data;

	const verifiable_credential: string = google_result.verifiableCredentials![0];

	return { type: 'Domain', manifest, verifiable_credential };
}
