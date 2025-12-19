import { Serializable, AsBytes } from './encoding';

// Default number of pbkdf2 iterations
export const PBKDF2_ITER_DEFAULT = 1e5;
//export const PBKDF2_ITER_DEFAULT = 1e3;
// Maximum number of pbkdf2 iterations
export const PBKDF2_ITER_MAX = 1e7;

export type AESKey = Uint8Array;
export type HMACKey = Uint8Array;
export type SymmetricKey = AESKey | HMACKey;
export type RSAPublicKey = Uint8Array;
export type RSAPrivateKey = Uint8Array;

export class AESEncryptionParams extends Serializable {
	algorithm: 'AES-GCM' | 'AES-CCM' = 'AES-GCM';
	tagSize: 64 | 96 | 128 = 128;
	keySize: 256 = 256;

	@AsBytes()
	iv: Uint8Array = new Uint8Array();

	@AsBytes()
	additionalData = new Uint8Array();

	validate() {
		return (
			['AES-GCM', 'AES-CCM'].includes(this.algorithm) &&
			[64, 96, 128].includes(this.tagSize) &&
			[256].includes(this.keySize)
		);
	}
}

export class AESKeyParams extends Serializable {
	algorithm: 'AES' = 'AES';
	keySize: 256 = 256;

	validate() {
		return this.algorithm === 'AES' && this.keySize === 256;
	}
}

export class HMACKeyParams extends Serializable {
	algorithm: 'HMAC' = 'HMAC';
	keySize: 256 = 256;

	validate() {
		return this.algorithm === 'HMAC' && this.keySize === 256;
	}
}

export class RSAKeyParams extends Serializable {
	algorithm: 'RSA' = 'RSA';
	modulusLength: 2048 = 2048;
	hash: 'SHA-256' = 'SHA-256';

	@AsBytes()
	publicExponent: Uint8Array = new Uint8Array([0x01, 0x00, 0x01]);

	validate() {
		return (
			this.algorithm === 'RSA' &&
			this.modulusLength === 2048 &&
			this.hash === 'SHA-256' &&
			this.publicExponent instanceof Uint8Array
		);
	}
}

export class PBKDF2Params extends Serializable {
	algorithm: 'PBKDF2' = 'PBKDF2';
	hash: 'SHA-256' = 'SHA-256';
	keySize: 256 = 256;
	iterations: number = PBKDF2_ITER_DEFAULT;

	@AsBytes()
	salt: Uint8Array = new Uint8Array();

	constructor(props?: Partial<PBKDF2Params>) {
		super();
		props && Object.assign(this, props);
	}

	validate() {
		return (
			this.algorithm === 'PBKDF2' &&
			this.hash === 'SHA-256' &&
			this.keySize === 256 &&
			typeof this.iterations === 'number' &&
			this.iterations < PBKDF2_ITER_MAX &&
			this.salt instanceof Uint8Array
		);
	}
}

export class RSAEncryptionParams extends Serializable {
	algorithm: 'RSA-OAEP' = 'RSA-OAEP';
	hash: 'SHA-256' = 'SHA-256';

	validate() {
		return this.algorithm === 'RSA-OAEP' && this.hash === 'SHA-256';
	}
}

export class RSASigningParams extends Serializable {
	algorithm: 'RSA-PSS' = 'RSA-PSS';
	hash: 'SHA-256' = 'SHA-256';
	saltLength: 32 = 32;

	validate() {
		return (
			this.algorithm === 'RSA-PSS' &&
			this.hash === 'SHA-256' &&
			this.saltLength === 32
		);
	}
}

export class HMACParams extends Serializable {
	algorithm: 'HMAC' = 'HMAC';
	hash: 'SHA-1' | 'SHA-256' = 'SHA-256';
	keySize: number = 256;

	constructor(props?: Partial<HMACParams>) {
		super();
		props && Object.assign(this, props);
	}

	validate() {
		return (
			this.algorithm === 'HMAC' &&
			this.hash === 'SHA-256' &&
			this.keySize === 256
		);
	}
}

export class HashParams extends Serializable {
	algorithm: 'SHA-1' | 'SHA-256' = 'SHA-256';

	constructor(props?: Partial<HashParams>) {
		super();
		props && Object.assign(this, props);
	}

	validate() {
		return ['SHA-1', 'SHA-256'].includes(this.algorithm);
	}
}

/**
 * CryptoProvider provides a unified interface for cryptographic primitives
 * accross all platforms. This is usually a thin wrapper around the
 * native crypto module provided by the platform.
 */
export interface CryptoProvider {
	/**
	 * Generates an Array of `n` random bytes
	 */
	randomBytes(n: number): Promise<Uint8Array>;

	/**
	 * Creates a digest of the provided **input** using the algorithm specified in **params**
	 * For algorithms that should be supported, see [[HashParams]].
	 */
	hash(input: Uint8Array, params: HashParams): Promise<Uint8Array>;

	/**
	 * Generates a random key or key pair for the algorithm specified in **params**
	 */
	generateKey(params: AESKeyParams): Promise<AESKey>;
	generateKey(params: HMACKeyParams): Promise<HMACKey>;
	generateKey(
		params: RSAKeyParams
	): Promise<{ privateKey: RSAPrivateKey; publicKey: RSAPublicKey }>;

	/**
	 * Derives a key from a given `password` using the provided key derivation **params**
	 * `password` should be the byte array representation of a UTF-8 encoded password.
	 */
	deriveKey(
		password: Uint8Array,
		params: PBKDF2Params
	): Promise<SymmetricKey>;

	/**
	 * Encrypts `data` with `key` using the cipher and parameters specified in `params`
	 */
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

	/**
	 * Decrypts `data` with `key` using the cipher and parameters specified in `params`
	 */
	decrypt(
		key: AESKey,
		data: Uint8Array,
		params: AESEncryptionParams
	): Promise<Uint8Array>;
	decrypt(
		privateKey: RSAPrivateKey,
		data: Uint8Array,
		params: RSAEncryptionParams
	): Promise<Uint8Array>;

	/**
	 * Creates a signature from `data` with `key` using the algorithm and parameters specified in `params`
	 */
	sign(
		key: HMACKey,
		data: Uint8Array,
		params: HMACParams
	): Promise<Uint8Array>;
	sign(
		key: RSAPrivateKey,
		data: Uint8Array,
		params: RSASigningParams
	): Promise<Uint8Array>;

	/**
	 * Verifies `signature` with `data` and `key` using the algorithm and parameters specified in `params`
	 */
	verify(
		key: HMACKey,
		signature: Uint8Array,
		data: Uint8Array,
		params: HMACParams
	): Promise<boolean>;
	verify(
		key: RSAPublicKey,
		signature: Uint8Array,
		data: Uint8Array,
		params: RSASigningParams
	): Promise<boolean>;

	/**
	 * Creates a fingerprint from a given rsa public key
	 */
	fingerprint(key: RSAPublicKey): Promise<Uint8Array>;

	/**
	 * Compares two values without leaking timing information that would allow an attacker to guess one of the values
	 */
	timingSafeEqual(a: Uint8Array, b: Uint8Array): Promise<boolean>;
}
