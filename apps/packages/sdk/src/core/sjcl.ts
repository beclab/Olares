import { PBKDF2Params, AESKey, AESEncryptionParams } from './crypto';
import { Err, ErrorCode } from './error';
import { sjcl } from '../vendor/sjcl';

// Shorthands for codec functions
const bytesToBits = sjcl.codec.bytes.toBits;
const bitsToBytes = sjcl.codec.bytes.fromBits;

/**
 * [[CrypoProvider]] implemenation using the [SJCL](https://github.com/bitwiseshiftleft/sjcl)
 * library. This is used to decrypt legacy (<3.x) Padlock containers that use AES in CCM mode
 * which is unfortunately not supported by the WebCrypto standard. Only
 * supports encryption/decryption using AES-CCM.
 */
export const SJCLProvider = {
	isAvailable() {
		return true;
	},

	randomBytes(_bytes: number): Uint8Array {
		throw new Err(ErrorCode.NOT_SUPPORTED);
	},

	async deriveKey(_password: string, _params: PBKDF2Params): Promise<AESKey> {
		throw new Err(ErrorCode.NOT_SUPPORTED);
	},

	async randomKey(_n = 256) {
		throw new Err(ErrorCode.NOT_SUPPORTED);
	},

	async decrypt(
		key: AESKey,
		ct: Uint8Array,
		params: AESEncryptionParams
	): Promise<Uint8Array> {
		if (params.algorithm !== 'AES-CCM') {
			throw new Err(ErrorCode.INVALID_ENCRYPTION_PARAMS);
		}

		// Only AES and CCM are supported
		const algorithm = 'aes';
		const mode = 'ccm';

		try {
			const cipher = new sjcl.cipher[algorithm](bytesToBits(key));
			const pt = sjcl.mode[mode].decrypt(
				cipher,
				bytesToBits(ct),
				bytesToBits(params.iv!),
				bytesToBits(params.additionalData!),
				params.tagSize
			);
			return new Uint8Array(bitsToBytes(pt));
		} catch (e) {
			throw new Err(ErrorCode.DECRYPTION_FAILED);
		}
	},

	async encrypt(
		key: AESKey,
		pt: Uint8Array,
		params: AESEncryptionParams
	): Promise<Uint8Array> {
		if (params.algorithm !== 'AES-CCM') {
			throw new Err(ErrorCode.INVALID_ENCRYPTION_PARAMS);
		}

		// Only AES and CCM are supported
		const algorithm = 'aes';
		const mode = 'ccm';

		try {
			const cipher = new sjcl.cipher[algorithm](bytesToBits(key));
			const ct = sjcl.mode[mode].encrypt(
				cipher,
				bytesToBits(pt),
				bytesToBits(params.iv!),
				bytesToBits(params.additionalData!),
				params.tagSize
			);
			return new Uint8Array(bitsToBytes(ct));
		} catch (e) {
			throw new Err(ErrorCode.ENCRYPTION_FAILED);
		}
	}
};
