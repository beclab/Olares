import { base64ToString, stringToBase64 } from '@didvault/sdk/src/core';
import { useSSIStore } from '../../stores/ssi';
import { ClientSchema } from '../../globals';
import { VCCardInfo, getSubmitApplicationJWS } from '../../utils/vc';
import { GeneralJwsSigner } from '../../jose/jws/general/signer';
import { walletService } from '../../wallet';
import {
	PrivateJwk,
	GetResponseResponse,
	CredentialManifest
} from '@bytetrade/core';
import { PrivateKey } from '@trustwallet/wallet-core/dist/src/wallet-core';

import { i18n } from '../../boot/i18n';
import { TermipassConfig } from '../../utils/config';

function getFullyQualifiedVerificationMethodID(did: string): string {
	const [, , id] = did.split(':');

	return did + '#' + id;
}

export async function signJWS(
	issuer: string,
	data: any,
	privateKey: PrivateJwk
) {
	const t = { ...data };

	const issuanceDate = new Date(); //date.toISOString();
	const expirationDate = new Date(
		issuanceDate.getTime() + 1000 * 60 * 60 * 24 * 365
	);

	const unixTime = expirationDate.getTime();
	if (!unixTime) {
		throw Error(
			i18n.global.t('errors.could_not_convert_expiration_date_to_unix_time')
		);
	}
	t.exp = Math.floor(unixTime / 1000);

	t.iss = getFullyQualifiedVerificationMethodID(issuer);
	t.nbf =
		Math.floor(issuanceDate.getTime() / 1000) -
		TermipassConfig.jwt_payload_nbf_sub_second;

	const signer = await GeneralJwsSigner.create(
		new TextEncoder().encode(JSON.stringify(t)),
		[
			{
				privateJwk: privateKey,
				protectedHeader: {
					alg: 'EdDSA',
					kid: getFullyQualifiedVerificationMethodID(issuer)
				}
			}
		]
	);

	return signer.getJws();
}

export async function signStatement(
	signDomain: any,
	types: any,
	signData: any,
	primaryType: string,
	privateKey: any
) {
	const { EthereumMessageSigner } = walletService.walletCore;

	const outputData = EthereumMessageSigner.signTypedMessage(
		privateKey,
		JSON.stringify({
			types: types,
			domain: signDomain,
			message: signData,
			primaryType: primaryType //'AuthAddressReq'
		})
	);
	return `0x${outputData}`;
}

export async function requestVC(
	did: string,
	vc_type: string,
	vc_request_path: string,
	vc_sign_data: any,
	privateJWK: PrivateJwk
): Promise<VCCardInfo> {
	const ssiStore = useSSIStore();
	const schema: ClientSchema | undefined =
		await ssiStore.get_application_schema(vc_type);
	if (!schema) {
		throw Error(i18n.global.t('errors.get_schema_failure'));
	}
	const manifest = stringToBase64(JSON.stringify(schema?.manifest));

	const jws = await getSubmitApplicationJWS(
		did,
		privateJWK,
		schema!.manifest,
		schema!.application_verifiable_credential.id,
		vc_sign_data
	);

	const response: any = await ssiStore.vcInstance().post(vc_request_path, {
		jws: jws
	});

	if (
		(response.status != 200 && response.status != 201) ||
		response.data.code != 0
	) {
		throw Error(i18n.global.t('errors.get_google_result_failure'));
	}
	const google_result: GetResponseResponse = response.data.data;
	const verifiable_credential: string = google_result.verifiableCredentials![0];

	return { type: vc_type, manifest, verifiable_credential };
}

export function defaultDriverPath(index: number) {
	const bip44chainID = 60;

	const { DerivationPath, Purpose } = walletService.walletCore;
	const derivationPath = DerivationPath.create(
		Purpose.bip44,
		bip44chainID,
		0,
		0,
		index
	);
	return derivationPath.description();
}

export function mnemonicToAddress(mnemonic: string, path: string): string {
	const { AnyAddress, HDWallet, CoinType } = walletService.walletCore;
	const hdWallet = HDWallet.createWithMnemonic(mnemonic, '');
	const key = hdWallet.getKey(CoinType.ethereum, path);
	const publicKey = key.getPublicKeySecp256k1(false);
	const address = AnyAddress.createWithPublicKey(
		publicKey,
		CoinType.ethereum
	).description();
	return address;
}

export function mnemonicToKey(mnemonic: string, path: string): PrivateKey {
	const { HDWallet, CoinType } = walletService.walletCore;
	const hdWallet = HDWallet.createWithMnemonic(mnemonic, '');
	const key = hdWallet.getKey(CoinType.ethereum, path);
	return key;
}
