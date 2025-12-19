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
} from './core';
import { Err, ErrorCode } from './core';
import { SJCLProvider } from './core';

const webCrypto = require('crypto');
const NodeRSA = require('node-rsa');
//const forge = require('node-forge');
//const rsa = forge.pki.rsa;

function hashToNode(name: string) {
	return name.replace('-', '').toLowerCase();
}

//const crypto2 = require('public-encrypt/browser')

export class WebCryptoProviderLocal implements CryptoProvider {
	async randomBytes(n: number): Promise<Uint8Array> {
		const bytes = require('randombytes')(n);
		return bytes;
	}

	async hash(input: Uint8Array, params: HashParams): Promise<Uint8Array> {
		let bytes;
		switch (params.algorithm) {
			case 'SHA-1':
				bytes = webCrypto.createHash('sha1').update(input).digest('hex');
				break;
			case 'SHA-256':
				bytes = webCrypto.createHash('sha256').update(input).digest('hex');
				break;
			default:
				bytes = webCrypto.createHash('sha256').update(input).digest('hex');
				break;
		}

		return Uint8Array.from(Buffer.from(bytes, 'hex'));
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
				// eslint-disable-next-line no-case-declarations
				const key = new NodeRSA(
					{ b: params.modulusLength, e: 65537 },
					{ encryptionScheme: 'pkcs1' }
				);

				return {
					privateKey: new Uint8Array(key.exportKey('pkcs8-private-der')),
					publicKey: new Uint8Array(key.exportKey('public-der'))
				};
		}
	}

	async deriveKey(
		password: Uint8Array,
		params: PBKDF2Params
	): Promise<SymmetricKey> {
		return webCrypto.pbkdf2Sync(
			password,
			Buffer.from(params.salt),
			params.iterations,
			params.keySize / 8,
			hashToNode(params.hash)
		);
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
			case 'AES-CCM':
				return this._encryptAES(key, data, params);
			case 'RSA-OAEP':
				return this._encryptRSA(key, data, params);
			default:
				throw new Err(ErrorCode.INVALID_ENCRYPTION_PARAMS);
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
		switch (params.algorithm) {
			case 'AES-GCM':
			case 'AES-CCM':
				return this._decryptAES(key, data, params);
			case 'RSA-OAEP':
				return this._decryptRSA(key, data, params);
			default:
				throw new Err(ErrorCode.INVALID_ENCRYPTION_PARAMS);
		}
	}

	async timingSafeEqual(a: Uint8Array, b: Uint8Array): Promise<boolean> {
		const compareKey = await this.generateKey(new HMACKeyParams());
		const hmacA = await this.sign(compareKey, a, new HMACParams());
		const hmacB = await this.sign(compareKey, b, new HMACParams());

		let match = true;

		for (let i = 0; i < hmacA.length; i++) {
			match = match && hmacA[i] === hmacB[i];
		}

		return hmacA.length === hmacB.length && match;
	}

	Uint8ArrayEqual(a: Uint8Array, b: Uint8Array): boolean {
		if (!a && !b) {
			return true;
		}

		if (!a || !b || a.length != b.length) {
			return false;
		}

		for (let i = 0; i < a.length; i++) {
			if (a[i] !== b[i]) {
				return false;
			}
		}

		return true;
	}

	private async _encryptAES(
		key: AESKey,
		data: Uint8Array,
		params: AESEncryptionParams
	): Promise<Uint8Array> {
		if (params.algorithm === 'AES-CCM') {
			return SJCLProvider.encrypt(key, data, params);
		}

		// var crypter = crypto.createCipheriv("aes", key)
		const [alg, mode] = params.algorithm.toLowerCase().split('-');
		const authTagLength = params.tagSize / 8;
		const cipher = webCrypto.createCipheriv(
			`${alg}-${params.keySize}-${mode}` as 'aes-256-gcm',
			key,
			params.iv,
			{
				authTagLength
			} as any
		);
		cipher.setAAD(params.additionalData as Buffer);
		try {
			return new Uint8Array(
				Buffer.concat([
					cipher.update(data),
					cipher.final(),
					cipher.getAuthTag()
				])
			);
		} catch (e) {
			throw new Err(ErrorCode.ENCRYPTION_FAILED);
		}
	}

	private async _decryptAES(
		key: AESKey,
		data: Uint8Array,
		params: AESEncryptionParams
	): Promise<Uint8Array> {
		if (params.algorithm === 'AES-CCM') {
			return SJCLProvider.decrypt(key, data, params);
		}
		const [alg, mode] = params.algorithm.toLowerCase().split('-');
		const authTagLength = params.tagSize / 8;
		const tagPos = data.length - authTagLength;
		const enc = data.slice(0, tagPos);
		const tag = data.slice(tagPos);

		const cipher = webCrypto.createDecipheriv(
			`${alg}-${params.keySize}-${mode}` as 'aes-256-gcm',
			key,
			params.iv,
			{
				authTagLength
			} as any
		);
		cipher.setAAD(params.additionalData as Buffer);
		cipher.setAuthTag(tag);
		try {
			return new Uint8Array(
				Buffer.concat([cipher.update(enc), cipher.final()])
			);
		} catch (e) {
			console.error(e);
			throw new Err(ErrorCode.DECRYPTION_FAILED);
		}
	}

	async _encryptRSA(
		publicKey: RSAPublicKey,
		key: AESKey,
		params: RSAEncryptionParams
	) {
		const key2 = new NodeRSA();
		key2.setOptions({
			encryptionScheme: {
				hash: hashToNode(params.hash),
				label: null
			}
		});

		key2.importKey(Buffer.from(publicKey), 'public-der');
		return key2.encrypt(Buffer.from(key), 'buffer', 'buffer');
	}

	async _decryptRSA(
		privateKey: RSAPrivateKey,
		key: AESKey,
		params: RSAEncryptionParams
	) {
		const key2 = new NodeRSA();

		key2.setOptions({
			encryptionScheme: {
				hash: hashToNode(params.hash),
				label: null
			}
		});

		key2.importKey(Buffer.from(privateKey), 'pkcs8-private-der');

		return key2.decrypt(Buffer.from(key), 'buffer');
	}

	async fingerprint(key: RSAPublicKey): Promise<Uint8Array> {
		const bytes = webCrypto.createHash('sha256').update(key).digest('hex');
		return Uint8Array.from(Buffer.from(bytes, 'hex'));
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
		params: HMACParams | RSASigningParams
	): Promise<Uint8Array> {
		switch (params.algorithm) {
			case 'HMAC':
				return this._signHMAC(key, data, params);
			case 'RSA-PSS':
				return this._signRSA(key, data, params);
			default:
				throw new Err(ErrorCode.NOT_SUPPORTED);
		}
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
		params: HMACParams | RSASigningParams
	): Promise<boolean> {
		switch (params.algorithm) {
			case 'HMAC':
				return this._verifyHMAC(key, signature, data, params);
			case 'RSA-PSS':
				return this._verifyRSA(key, signature, data, params);
			default:
				throw new Err(ErrorCode.NOT_SUPPORTED);
		}
	}

	private async _signHMAC(
		key: HMACKey,
		data: Uint8Array,
		params: HMACParams
	): Promise<Uint8Array> {
		const hmac = webCrypto.createHmac(
			hashToNode(params.hash),
			Buffer.from(key)
		);
		hmac.update(Buffer.from(data));
		hmac.end();
		return hmac.read();
	}

	private async _verifyHMAC(
		key: HMACKey,
		signature: Uint8Array,
		data: Uint8Array,
		params: HMACParams
	): Promise<boolean> {
		// const p = Object.assign({}, params, {
		//   name: params.algorithm,
		//   length: params.keySize,
		// });
		// const k = await webCrypto.importKey("raw", key, p, false, ["verify"]);
		// return await webCrypto.verify(p, k, signature, data);

		const hmac = webCrypto.createHmac(
			hashToNode(params.hash),
			Buffer.from(key)
		);
		hmac.update(Buffer.from(data));
		hmac.end();
		return this.Uint8ArrayEqual(signature, hmac.read());
	}

	private async _signRSA(
		key: RSAPrivateKey,
		data: Uint8Array,
		params: RSASigningParams
	): Promise<Uint8Array> {
		const key2 = new NodeRSA();
		key2.setOptions({
			signingScheme: {
				scheme: 'pss',
				hash: hashToNode(params.hash),
				saltLength: 32
			}
		});
		console.log(key2.$options);

		key2.importKey(Buffer.from(key), 'pkcs8-private-der');

		return key2.sign(Buffer.from(data), 'buffer', 'buffer');
	}

	private async _verifyRSA(
		key: RSAPublicKey,
		signature: Uint8Array,
		data: Uint8Array,
		params: RSASigningParams
	): Promise<boolean> {
		const key2 = new NodeRSA();
		key2.setOptions({
			signingScheme: {
				scheme: 'pss',
				hash: hashToNode(params.hash),
				saltLength: 32
			}
		});

		key2.importKey(Buffer.from(key), 'public-der');
		return key2.verify(
			Buffer.from(data),
			Buffer.from(signature),
			'buffer',
			'buffer'
		);
	}
}

export default WebCryptoProviderLocal;
