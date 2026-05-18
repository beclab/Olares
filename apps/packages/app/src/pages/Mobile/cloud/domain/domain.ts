import { PrivateJwk, GetResponseResponse } from '@bytetrade/core';
import { stringToBase64, VaultItem } from '@didvault/sdk/src/core';
import { useSSIStore } from '../../../../stores/ssi';
import { useCloudStore } from '../../../../stores/cloud';
import { ClientSchema } from '../../../../globals';
import { i18n } from '../../../../boot/i18n';
import {
	VCCardInfo,
	convertVault2CardItem,
	getSubmitApplicationJWS
} from 'src/utils/vc';
import { app } from '../../../../globals';

export async function getDomainVC(
	owner: string,
	did: string,
	domain: string,
	cloudid: string,
	privateJWK: PrivateJwk
): Promise<VCCardInfo> {
	const cloudStore = useCloudStore();

	const { manifest, schema } = await getDomainSchema();

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

export const getDomainSchema = async () => {
	const ssiStore = useSSIStore();
	const schema: ClientSchema | undefined =
		await ssiStore.get_application_schema('Domain');
	if (!schema) {
		throw Error(i18n.global.t('errors.get_schema_failure'));
	}
	const manifest = stringToBase64(JSON.stringify(schema?.manifest));
	return {
		manifest: manifest,
		schema: schema
	};
};

export const getDomainVaultItems = (domain: string) => {
	const vcList: {
		card: VCCardInfo;
		item: VaultItem;
	}[] = [];

	if (domain) {
		for (const vault of app.vaults) {
			for (const item of vault.items) {
				const card = convertVault2CardItem(item);

				if (card && card?.type === domain) {
					vcList.push({
						card: card,
						item: item
					});
				}
			}
		}
		return vcList;
	}

	return [];
};
