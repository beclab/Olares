import { Buffer } from 'buffer';
import {
	createCipheriv,
	createDecipheriv,
	pbkdf2,
	randomBytes
} from 'crypto-browserify';
import { Err, ErrorCode } from './error';
import {
	AsBytes,
	AsSerializable,
	Exclude,
	Serializable
} from './serialization';

class AESEncryptionParams extends Serializable {
	algorithm: 'AES-GCM' | 'AES-CCM' = 'AES-GCM';
	tagSize = 128;
	keySize = 256;

	@AsBytes()
	iv = new Uint8Array();

	@AsBytes()
	additionalData = new Uint8Array();

	get kind() {
		return 'aesencryptionparams';
	}

	validate() {
		return (
			['AES-GCM', 'AES-CCM'].includes(this.algorithm) &&
			[64, 96, 128].includes(this.tagSize) &&
			this.keySize === 256
		);
	}
}

class PBKDF2Params extends Serializable {
	algorithm = 'PBKDF2';
	hash = 'SHA-256';
	keySize = 256;
	iterations = 100000;

	@AsBytes()
	salt = new Uint8Array();

	get kind() {
		return 'pbkdf2params';
	}

	validate() {
		return (
			this.algorithm === 'PBKDF2' &&
			this.hash === 'SHA-256' &&
			this.keySize === 256 &&
			Number.isInteger(this.iterations) &&
			this.iterations > 0 &&
			this.iterations < 10000000
		);
	}
}

function deriveKey(
	password: string,
	params: PBKDF2Params
): Promise<Uint8Array> {
	return new Promise((resolve, reject) => {
		pbkdf2(
			Buffer.from(password, 'utf8'),
			Buffer.from(params.salt),
			params.iterations,
			params.keySize / 8,
			'sha256',
			(error: Error | null, key: Uint8Array) => {
				if (error) {
					reject(error);
				} else {
					resolve(key);
				}
			}
		);
	});
}

async function decrypt(
	key: Uint8Array,
	ciphertext: Uint8Array,
	params: AESEncryptionParams
) {
	if (params.algorithm === 'AES-CCM') {
		const { sjcl } = await import('./sjcl');
		const { toBits, fromBits } = sjcl.codec.bytes;
		return new Uint8Array(
			fromBits(
				sjcl.mode.ccm.decrypt(
					new sjcl.cipher.aes(toBits(key)),
					toBits(ciphertext),
					toBits(params.iv),
					toBits(params.additionalData),
					params.tagSize
				)
			)
		);
	}

	if (params.tagSize !== 128 && globalThis.crypto?.subtle) {
		const subtle = globalThis.crypto.subtle;
		const aesKey = await subtle.importKey(
			'raw',
			Uint8Array.from(key),
			'AES-GCM',
			false,
			['decrypt']
		);
		const data = await subtle.decrypt(
			{
				name: 'AES-GCM',
				iv: Uint8Array.from(params.iv),
				additionalData: Uint8Array.from(params.additionalData),
				tagLength: params.tagSize
			},
			aesKey,
			Uint8Array.from(ciphertext)
		);
		return new Uint8Array(data);
	}

	const tagOffset = ciphertext.length - params.tagSize / 8;
	if (tagOffset < 0) {
		throw new Err(ErrorCode.DECRYPTION_FAILED);
	}
	const decipher = createDecipheriv(
		'aes-256-gcm',
		Buffer.from(key),
		Buffer.from(params.iv)
	);
	decipher.setAAD(Buffer.from(params.additionalData));
	decipher.setAuthTag(Buffer.from(ciphertext.subarray(tagOffset)));
	return Buffer.concat([
		decipher.update(Buffer.from(ciphertext.subarray(0, tagOffset))),
		decipher.final()
	]);
}

/** Password-protected data stored alongside local account metadata. */
export abstract class PBES2Container extends Serializable {
	@AsSerializable(AESEncryptionParams)
	encryptionParams = new AESEncryptionParams();

	@AsBytes()
	encryptedData?: Uint8Array = undefined;

	@AsSerializable(PBKDF2Params)
	keyParams = new PBKDF2Params();

	@Exclude()
	protected _key?: Uint8Array = undefined;

	protected async _deriveAndSetKey(password: string) {
		if (!this.keyParams.salt.length) {
			this.keyParams.salt = randomBytes(16);
		}
		this._key = await deriveKey(password, this.keyParams);
	}

	protected async setData(data: Uint8Array) {
		const key = this._key;
		if (!key) {
			throw new Err(ErrorCode.ENCRYPTION_FAILED, 'No encryption key provided!');
		}

		const params = new AESEncryptionParams();
		params.iv = randomBytes(16);
		params.additionalData = randomBytes(16);
		const cipher = createCipheriv(
			'aes-256-gcm',
			Buffer.from(key),
			Buffer.from(params.iv)
		);
		cipher.setAAD(Buffer.from(params.additionalData));
		const encrypted = Buffer.concat([
			cipher.update(Buffer.from(data)),
			cipher.final(),
			cipher.getAuthTag()
		]);
		this.encryptionParams = params;
		this.encryptedData = encrypted;
	}

	protected async unlockData(password: string) {
		this._key = undefined;
		if (!this.encryptedData) {
			await this._deriveAndSetKey(password);
			return undefined;
		}

		const key = await deriveKey(password, this.keyParams);
		const data = await this.readData(password, key);
		this._key = key;
		return data;
	}

	lock() {
		this._key = undefined;
	}

	async verifyPassword(password: string) {
		if (!this.encryptedData) {
			return true;
		}
		if (!this.keyParams.salt.length) {
			return false;
		}
		try {
			const key = await deriveKey(password, this.keyParams);
			await this.readData(password, key);
			return true;
		} catch {
			return false;
		}
	}

	private async readData(password: string, key: Uint8Array) {
		const data = this.encryptedData;
		if (!data) {
			throw new Err(ErrorCode.DECRYPTION_FAILED);
		}
		try {
			return await decrypt(key, data, this.encryptionParams);
		} catch {
			return this.readOldBrowserData(password, data);
		}
	}

	// Compatibility with records saved by the old browser placeholder. Only
	// the read path accepts these records. The next save writes AES-GCM data.
	private readOldBrowserData(password: string, data: Uint8Array) {
		const params = this.encryptionParams;
		const prefix = Buffer.concat([
			Buffer.concat(
				[Buffer.from(password, 'utf8'), Buffer.from(this.keyParams.salt)],
				32
			),
			Buffer.from(params.iv),
			Buffer.from(params.additionalData)
		]);
		if (
			params.algorithm !== 'AES-GCM' ||
			params.tagSize !== 128 ||
			params.iv.length !== 16 ||
			params.additionalData.length !== 16 ||
			this.keyParams.salt.length !== 16 ||
			data.length <= prefix.length ||
			!prefix.equals(Buffer.from(data.subarray(0, prefix.length)))
		) {
			throw new Err(ErrorCode.DECRYPTION_FAILED);
		}
		const contents = data.subarray(prefix.length);
		try {
			const record = JSON.parse(new TextDecoder().decode(contents));
			if (Array.isArray(record.items)) {
				return contents;
			}
		} catch {
			// The old record must contain the expected serialized collection.
		}
		throw new Err(ErrorCode.DECRYPTION_FAILED);
	}
}
