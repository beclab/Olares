import * as varint from 'varint';

import { base58btc } from 'multiformats/bases/base58';
import { base64url } from 'multiformats/bases/base64';
import { walletService } from '../wallet';
import * as bip39 from 'bip39';

// multicodec code for Ed25519 keys
const ED25519_CODEC_ID = varint.encode(parseInt('0xed', 16));

function getID(mnemonic) {
	const { HDWallet, Curve } = walletService.walletCore;
	const w = HDWallet.createWithMnemonic(mnemonic, '');
	const privatekey = w.getMasterKey(Curve.ed25519);
	const publicKey = privatekey.getPublicKeyEd25519();

	const idBytes = new Uint8Array(
		publicKey.data().length + ED25519_CODEC_ID.length
	);
	idBytes.set(ED25519_CODEC_ID, 0);
	idBytes.set(publicKey.data(), ED25519_CODEC_ID.length);

	const id = base58btc.encode(idBytes);
	return id;
}

export function getPrivateKeyBase58(mnemonic) {
	const { HDWallet, Curve } = walletService.walletCore;
	const w = HDWallet.createWithMnemonic(mnemonic, '');
	const privatekey = w.getMasterKey(Curve.ed25519);
	const id = base58btc.encode(privatekey.data());
	return id;
}

export async function getEthereumAddress(mnemonic) {
	await walletService.loaded;
	const { HDWallet, CoinType } = walletService.walletCore;
	const w = HDWallet.createWithMnemonic(mnemonic, '');

	return w.getAddressForCoin(CoinType.ethereum);
}

export async function getDID(mnemonic) {
	await walletService.loaded;
	const id = getID(mnemonic);
	const did = `did:key:${id}`;

	return did;
}

export async function getPublicJWK(mnemonic) {
	await walletService.loaded;
	const { HDWallet, Curve } = walletService.walletCore;
	const id = getID(mnemonic);
	const did = `did:key:${id}`;
	const keyId = `${did}#${id}`;
	const w = HDWallet.createWithMnemonic(mnemonic, '');
	const privatekey = w.getMasterKey(Curve.ed25519);
	const publicKey = privatekey.getPublicKeyEd25519();

	// TODO: `publicJWK` code is duplicated in resolve. move JWK creation into a separate method (Moe - 08/01/2022)
	const publicJWK = {
		alg: 'EdDSA',
		crv: 'Ed25519',
		kid: keyId,
		kty: 'OKP',
		use: 'sig',
		x: base64url.baseEncode(publicKey.data())
	};

	return publicJWK;
}

export async function getPrivateJWK(mnemonic) {
	await walletService.loaded;
	const { HDWallet, Curve } = walletService.walletCore;
	const id = getID(mnemonic);
	const did = `did:key:${id}`;
	const keyId = `${did}#${id}`;
	const w = HDWallet.createWithMnemonic(mnemonic, '');
	const privatekey = w.getMasterKey(Curve.ed25519);
	const publicKey = privatekey.getPublicKeyEd25519();

	const privateJWK = {
		alg: 'EdDSA',
		crv: 'Ed25519',
		kid: keyId,
		kty: 'OKP',
		use: 'sig',
		x: base64url.baseEncode(publicKey.data()),
		d: base64url.baseEncode(privatekey.data())
	};
	return privateJWK;
}

/**
 * @typedef {Object} GenerateDIDResult
 * @property {string} did - the generated DID
 * @property {PublicJWK} publicJWK - the public key of the private key that was used to generate the DID
 * @property {PrivateJWK} privateJWK - the private key used to generate the DID
 */

/**
 * * [DID Spec Doc](https://www.w3.org/TR/did-core/)
 * * [DID-Key Spec Draft](https://w3c-ccg.github.io/did-method-key/)
 */
export class DIDKey {
	static async generate() {
		try {
			const mnemonic = bip39.generateMnemonic();
			return mnemonic;
		} catch (e) {
			console.error(e);
		}
	}

	static resolve(did) {
		const [scheme, method, id] = did.split(':');

		if (scheme !== 'did') {
			throw new Error('malformed scheme');
		}

		if (method !== 'key') {
			throw new Error('did method MUST be "key"');
		}

		const idBytes = base58btc.decode(id);
		const publicKeyBytes = idBytes.slice(ED25519_CODEC_ID.length);

		const publicJwk = {
			alg: 'EdDSA',
			crv: 'Ed25519',
			kty: 'OKP',
			use: 'sig',
			x: base64url.baseEncode(publicKeyBytes)
		};

		const keyId = `${did}#${id}`;

		const didDocument = {
			'@context': [
				'https://www.w3.org/ns/did/v1',
				'https://w3id.org/security/suites/ed25519-2020/v1',
				'https://w3id.org/security/suites/x25519-2020/v1'
			],
			id: did,
			verificationMethod: [
				{
					id: keyId,
					type: 'JsonWebKey2020',
					controller: did,
					publicKeyJwk: publicJwk
				}
			],
			authentication: [keyId],
			assertionMethod: [keyId],
			capabilityDelegation: [keyId],
			capabilityInvocation: [keyId]
		};

		return didDocument;
	}
}
