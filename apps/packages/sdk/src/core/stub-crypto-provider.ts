import {
	CryptoProvider,
	PBKDF2Params,
	AESKey,
	RSAPublicKey,
	RSAPrivateKey,
	HMACKey,
	SymmetricKey,
	AESKeyParams,
	RSAKeyParams,
	HMACParams,
	HMACKeyParams,
	AESEncryptionParams,
	RSAEncryptionParams,
	HashParams,
	RSASigningParams
} from './crypto';
import { concatBytes as concat, equalBytes as equal } from './encoding';
import { Err, ErrorCode } from './error';

/**
 * StubCryptoProvider is a stub implementation of the [[CryptoProvider]]
 * interface mainly used for testing. All methods merely emulate the behavior
 * of an actual implementation in a way that makes it compatible for use
 * with the rest of the **@padloc/core** package. Needless to say, this
 * class is **NOT SECURE AND SHOULD NEVER BE USED IN A PRODUCTION ENVIRONMENT**.
 */
export class StubCryptoProvider implements CryptoProvider {
	async randomBytes(n: number): Promise<Uint8Array> {
		const bytes = new Uint8Array(n);
		for (let i = 0; i < n; i++) {
			bytes[i] = Math.random() * 255;
		}
		return bytes;
	}

	async hash(input: Uint8Array, _params: HashParams): Promise<Uint8Array> {
		return input.slice(0, 32);
	}

	generateKey(params: AESKeyParams): Promise<AESKey>;
	generateKey(
		params: RSAKeyParams
	): Promise<{ privateKey: RSAPrivateKey; publicKey: RSAPublicKey }>;
	generateKey(params: HMACKeyParams): Promise<HMACKey>;
	async generateKey(params: AESKeyParams | RSAKeyParams | HMACKeyParams) {
		switch (params.algorithm) {
			case 'AES':
			case 'HMAC':
				return this.randomBytes(params.keySize / 8);
			case 'RSA':
				const key = await this.randomBytes(32);
				return {
					publicKey: key,
					privateKey: key
				};
		}
	}

	async deriveKey(
		password: Uint8Array,
		params: PBKDF2Params
	): Promise<SymmetricKey> {
		const bytes = new Uint8Array(params.keySize);
		bytes.set(concat([password, params.salt]));
		return bytes.slice(0, 32);
	}

	encrypt(
		key: AESKey,
		data: Uint8Array,
		params: AESEncryptionParams
	): Promise<Uint8Array>;
	encrypt(
		publicKey: RSAPublicKey,
		data: Uint8Array,
		params: RSAEncryptionParams
	): Promise<Uint8Array>;
	async encrypt(
		key: AESKey | RSAPublicKey,
		data: Uint8Array,
		params: AESEncryptionParams | RSAEncryptionParams
	): Promise<Uint8Array> {
		switch (params.algorithm) {
			case 'AES-GCM':
				return concat([key, params.iv, params.additionalData, data]);

			case 'RSA-OAEP':
				return concat([key, data]);
			default:
				throw new Err(ErrorCode.NOT_SUPPORTED);
		}
	}

	decrypt(
		key: AESKey,
		data: Uint8Array,
		params: AESEncryptionParams
	): Promise<Uint8Array>;
	decrypt(
		publicKey: RSAPublicKey,
		data: Uint8Array,
		params: RSAEncryptionParams
	): Promise<Uint8Array>;
	async decrypt(
		key: AESKey | RSAPublicKey,
		data: Uint8Array,
		params: AESEncryptionParams | RSAEncryptionParams
	): Promise<Uint8Array> {
		if (params.algorithm.startsWith('AES')) {
			params = params as AESEncryptionParams;
			const keyLength = key.length;
			const ivLength = params.iv.length;
			const adataLength = params.additionalData.length;
			const extractedKey = data.slice(0, keyLength);
			const iv = data.slice(keyLength, keyLength + ivLength);
			const adata = data.slice(
				keyLength + ivLength,
				keyLength + ivLength + adataLength
			);

			if (
				!equal(key, extractedKey) ||
				!equal(iv, params.iv) ||
				!equal(adata, params.additionalData)
			) {
				throw new Err(ErrorCode.DECRYPTION_FAILED);
			}
			return data.slice(keyLength + adataLength + ivLength);
		} else {
			const keyLength = key.length;
			const extractedKey = data.slice(0, keyLength);

			if (!equal(key, extractedKey)) {
				throw new Err(ErrorCode.DECRYPTION_FAILED);
			}
			return data.slice(keyLength);
		}
	}

	async fingerprint(key: RSAPublicKey): Promise<Uint8Array> {
		return key;
	}

	async sign(
		key: HMACKey,
		data: Uint8Array,
		params: HMACParams
	): Promise<Uint8Array>;
	async sign(
		key: RSAPrivateKey,
		data: Uint8Array,
		params: RSASigningParams
	): Promise<Uint8Array>;
	async sign(
		key: HMACKey | RSAPrivateKey,
		data: Uint8Array,
		_params: HMACParams | RSASigningParams
	): Promise<Uint8Array> {
		return concat([key, data]);
	}

	async verify(
		key: HMACKey,
		signature: Uint8Array,
		data: Uint8Array,
		params: HMACParams
	): Promise<boolean>;
	async verify(
		key: RSAPrivateKey,
		signature: Uint8Array,
		data: Uint8Array,
		params: RSASigningParams
	): Promise<boolean>;
	async verify(
		key: HMACKey | RSAPrivateKey,
		signature: Uint8Array,
		data: Uint8Array,
		_params: HMACParams | RSASigningParams
	): Promise<boolean> {
		const keyLength = key.length;
		const extractedKey = signature.slice(0, keyLength);
		const extractedData = signature.slice(keyLength);
		return equal(key, extractedKey) && equal(data, extractedData);
	}

	async timingSafeEqual(a: Uint8Array, b: Uint8Array): Promise<boolean> {
		let match = true;

		for (let i = 0; i < a.length; i++) {
			match = match && a[i] === b[i];
		}

		return a.length === b.length && match;
	}
}
