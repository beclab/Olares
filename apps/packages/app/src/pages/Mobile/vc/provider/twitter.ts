import { GolbalHost, PrivateJwk } from '@bytetrade/core';
import { stringToBase64 } from '@didvault/sdk/src/core';
import { useSSIStore } from '../../../../stores/ssi';
import { ClientSchema } from '../../../../globals';
import { i18n } from '../../../../boot/i18n';
import { VCCardInfo, getSubmitApplicationJWS } from 'src/utils/vc';
import { TwitterLogin } from 'src/platform/interface/capacitor/plugins/twitter';

export async function twitterLogin(
	did: string,
	privateJWK: PrivateJwk
): Promise<VCCardInfo> {
	const ssiStore = useSSIStore();
	const schema: ClientSchema | undefined =
		await ssiStore.get_application_schema('Twitter');
	if (!schema) {
		throw Error(i18n.global.t('errors.get_schema_failure'));
	}
	const manifest = stringToBase64(JSON.stringify(schema?.manifest));

	const jws = await getSubmitApplicationJWS(
		did!,
		privateJWK!,
		schema!.manifest,
		schema!.application_verifiable_credential.id,
		{ did: did }
	);
	const oauthUrl = GolbalHost.VC.en + '/request_twitter_auth?jws=' + jws;
	const result = await TwitterLogin.login({ oauthUrl });

	if (!result.status) {
		throw Error(i18n.global.t('errors.get_twitter_result_failure'));
	}

	const twitterresult = await ssiStore.get_application_response(result.message);

	if (!twitterresult) {
		throw Error(i18n.global.t('errors.get_twitter_result_failure'));
	}
	const verifiable_credential: string = twitterresult.verifiableCredentials![0];

	return { type: 'Twitter', manifest, verifiable_credential };
}
