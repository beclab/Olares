import {
	randomBytes,
	createHash,
	createHmac,
	sign,
	//verify,
	constants,
	generateKeyPair,
	pbkdf2,
	createCipheriv,
	createDecipheriv,
	publicEncrypt,
	privateDecrypt,
	timingSafeEqual
} from 'crypto';
import {
	CryptoProvider,
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
	RSASigningParams,
	PBKDF2Params
} from '../crypto';
import { Err, ErrorCode } from '../error';

// Converts hash algorithm name to a format that node can understand
// E.g.: "SHA-256" => "sha256"
function hashToNode(name: string) {
	return name.replace('-', '').toLowerCase();
}

export class NodeCryptoProvider implements CryptoProvider {
	async randomBytes(n: number) {
		return new Promise<Uint8Array>((resolve, reject) => {
			randomBytes(n, (err, buf) => {
				if (err) {
					reject(err);
				} else {
					resolve(new Uint8Array(buf));
				}
			});
		});
	}

	async hash(input: Uint8Array, params: HashParams) {
		const alg = hashToNode(params.algorithm);
		const hash = createHash(alg);
		hash.update(Buffer.from(input));
		return new Uint8Array(hash.digest());
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
				return new Promise<{
					privateKey: Uint8Array;
					publicKey: Uint8Array;
				}>((resolve, reject) =>
					generateKeyPair(
						'rsa',
						{
							modulusLength: params.modulusLength,
							publicExponent: Buffer.from(
								params.publicExponent
							).readUIntBE(0, params.publicExponent.length),
							publicKeyEncoding: {
								type: 'spki',
								format: 'der'
							},
							privateKeyEncoding: {
								type: 'pkcs8',
								format: 'der'
							}
						} as any,
						(err, publicKey: Buffer, privateKey: Buffer) => {
							if (err) {
								reject(err);
							} else {
								resolve({
									privateKey: new Uint8Array(privateKey),
									publicKey: new Uint8Array(publicKey)
								});
							}
						}
					)
				);
		}
	}

	async deriveKey(
		password: Uint8Array,
		params: PBKDF2Params
	): Promise<SymmetricKey> {
		return new Promise<SymmetricKey>((resolve, reject) =>
			pbkdf2(
				password,
				params.salt,
				params.iterations,
				params.keySize / 8,
				hashToNode(params.hash),
				(err, key) => {
					if (err) {
						reject(err);
					} else {
						resolve(new Uint8Array(key));
					}
				}
			)
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

	async fingerprint(key: RSAPublicKey): Promise<Uint8Array> {
		return await this.hash(key, new HashParams({ algorithm: 'SHA-256' }));
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
			// TODO: Implement, comment by pengpeng 2023.11.06, typescript compile error
			// case 'RSA-PSS':
			// 	return this._verifyRSA(key, signature, data, params);
			default:
				throw new Err(ErrorCode.NOT_SUPPORTED);
		}
	}

	async timingSafeEqual(a: Uint8Array, b: Uint8Array): Promise<boolean> {
		return a.length === b.length && timingSafeEqual(a, b);
	}

	private async _encryptAES(
		key: AESKey,
		data: Uint8Array,
		params: AESEncryptionParams
	): Promise<Uint8Array> {
		const [alg, mode] = params.algorithm.toLowerCase().split('-');
		const authTagLength = params.tagSize / 8;
		const cipher = createCipheriv(
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
		const [alg, mode] = params.algorithm.toLowerCase().split('-');
		const authTagLength = params.tagSize / 8;
		const tagPos = data.length - authTagLength;
		const enc = data.slice(0, tagPos);
		const tag = data.slice(tagPos);

		const cipher = createDecipheriv(
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

	private async _encryptRSA(
		publicKey: RSAPublicKey,
		key: AESKey,
		params: RSAEncryptionParams
	) {
		try {
			const buf = await publicEncrypt(
				{
					key: Buffer.from(publicKey),
					format: 'der',
					type: 'spki',
					oaepHash: hashToNode(params.hash)
				} as any,
				key
			);
			return new Uint8Array(buf);
		} catch (e) {
			throw new Err(ErrorCode.ENCRYPTION_FAILED);
		}
	}

	private async _decryptRSA(
		privateKey: RSAPrivateKey,
		key: AESKey,
		params: RSAEncryptionParams
	) {
		try {
			const buf = await privateDecrypt(
				{
					key: Buffer.from(privateKey),
					format: 'der',
					type: 'pkcs8',
					oaepHash: hashToNode(params.hash)
				} as any,
				key
			);
			return new Uint8Array(buf);
		} catch (e) {
			throw new Err(ErrorCode.DECRYPTION_FAILED);
		}
	}

	private async _signHMAC(
		key: HMACKey,
		data: Uint8Array,
		params: HMACParams
	): Promise<Uint8Array> {
		const hash = hashToNode(params.hash);
		const hmac = createHmac(hash, Buffer.from(key));
		hmac.update(Buffer.from(data));
		return new Uint8Array(hmac.digest());
	}

	private async _verifyHMAC(
		key: HMACKey,
		signature: Uint8Array,
		data: Uint8Array,
		params: HMACParams
	): Promise<boolean> {
		const sig = await this._signHMAC(key, data, params);
		return this.timingSafeEqual(sig, signature);
	}

	private _signRSA(
		key: RSAPrivateKey,
		data: Uint8Array,
		params: RSASigningParams
	) {
		const sig = sign(hashToNode(params.hash), Buffer.from(data), {
			key: Buffer.from(key),
			type: 'pkcs8',
			format: 'der',
			dsaEncoding: 'der',
			saltLength: params.saltLength,
			padding: constants.RSA_PKCS1_PSS_PADDING
		});

		return new Uint8Array(sig);
	}

	// TODO: Implement, comment by pengpeng 2023.11.06, typescript compile error

	// private _verifyRSA(
	// 	key: RSAPublicKey,
	// 	signature: Uint8Array,
	// 	data: Uint8Array,
	// 	params: RSASigningParams
	// ) {
	// 	return verify(
	// 		hashToNode(params.hash),
	// 		Buffer.from(data),
	// 		{
	// 			key: Buffer.from(key),
	// 			type: 'spki',
	// 			format: 'der',
	// 			dsaEncoding: 'der',
	// 			saltLength: params.saltLength,
	// 			padding: constants.RSA_PKCS1_PSS_PADDING
	// 		},
	// 		signature
	// 	);
	// }
}
