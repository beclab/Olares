import {
	Serializable,
	stringToBytes,
	AsBytes,
	AsSerializable
} from './encoding';
import { Err, ErrorCode } from './error';
import {
	PBKDF2Params,
	AESKey,
	AESEncryptionParams,
	AESKeyParams,
	RSAEncryptionParams,
	RSAPrivateKey,
	RSAPublicKey
} from './crypto';
import { getCryptoProvider as getProvider } from './platform';

/**
 * Base class for all **Container** implementations. In general, a **Container** is
 * an object for holding data encrypted using a symmetric cipher. Implementations
 * vary mostly in how the encryption key is generated. Sub classes must implement
 * the [[unlock]] method and may likely also want to augment [[lock]], [[validate]],
 * [[_fromRaw]] and [[_toRaw]].
 */
export abstract class BaseContainer extends Serializable {
	/** Parameters used for encryption of content data */
	@AsSerializable(AESEncryptionParams)
	encryptionParams: AESEncryptionParams = new AESEncryptionParams();

	/** Encrypted data */
	@AsBytes()
	encryptedData?: Uint8Array;

	/**
	 * The key used for encryption. Sub classes must set this property in the [[unlock]] method.
	 */
	protected _key?: AESKey;

	/**
	 * Encrypts the provided `data` and stores it in the container
	 */
	async setData(data: Uint8Array) {
		if (!this._key) {
			throw new Err(ErrorCode.ENCRYPTION_FAILED, 'No encryption key provided!');
		}

		// Generate random initialization vector
		this.encryptionParams.iv = await getProvider().randomBytes(16);

		// Generate additional authenticated data.
		// Note: Without knowing anything about the nature of the encrypted data,
		// we can't really choose a meaningful value for this. In the future,
		// we may want to provide the option to pass this as an argument but for now
		// a random value should be sufficient.
		this.encryptionParams.additionalData = await getProvider().randomBytes(16);

		// Encrypt the data and store it.
		this.encryptedData = await getProvider().encrypt(
			this._key,
			data,
			this.encryptionParams
		);
	}

	/**
	 * Decrypts and extracts the plain text data from the container. This will
	 * usually require unlocking the container first.
	 */
	async getData(): Promise<Uint8Array> {
		if (!this.encryptedData || !this._key) {
			throw new Err(ErrorCode.DECRYPTION_FAILED);
		}
		return await getProvider().decrypt(
			this._key,
			this.encryptedData,
			this.encryptionParams
		);
	}

	/**
	 * Unlocks the container, making it possible to extract the plain text
	 * data via [[getData]]. The type of **secret** provided will differ based
	 * on the encryption scheme used by implemenations.
	 */
	abstract unlock(secret: unknown): Promise<void>;

	/**
	 * Locks the container, removing the possibility to extract the plain text data
	 * via [[getData]] until the container is unlocked again. Subclasses extending
	 * this class must take care to delete any keys or other sensitive data
	 * that may have been stored temporarily after unlocking the container.
	 */
	lock() {
		delete this._key;
	}

	clone() {
		const clone = super.clone();
		clone._key = this._key;
		return clone;
	}
}

/**
 * Most basic **Container** implementation where the encryption key is
 * simply passed explicitly.
 */
export class SimpleContainer extends BaseContainer {
	async unlock(key: AESKey) {
		this._key = key;
	}
}

/**
 * Password-based **Container** that uses the
 * [PBES2](https://tools.ietf.org/html/rfc2898#section-6.2) encryption scheme,
 * deriving the encryption key from a user-provided passphrase.
 */
export class PBES2Container extends BaseContainer {
	/** Parameters used for key derivation */
	@AsSerializable(PBKDF2Params)
	keyParams: PBKDF2Params = new PBKDF2Params();

	protected async _deriveAndSetKey(password: string) {
		if (!this.keyParams.salt.length) {
			this.keyParams.salt = await getProvider().randomBytes(16);
		}
		this._key = await getProvider().deriveKey(
			stringToBytes(password),
			this.keyParams
		);
	}

	/**
	 * Unlocks the container using the given **password**
	 */
	async unlock(password: string) {
		await this._deriveAndSetKey(password);
		// If this container has data already, make sure the derived key properly decrypts it.
		if (this.encryptedData) {
			await this.getData();
		}
	}
}

/**
 * Represents an individual with access to a [[SharedContainer]]. Each accessor is mapped
 * to an entity owning a public/private key pair via their `id`.
 */
export class Accessor extends Serializable {
	/**
	 * Identifier used to map an `Accessor` to the owner of the public key used to encrypt the shared key.
	 */
	id = '';

	/** Shared key encrypted with the public key of the entity associated with the `Accessor` object */
	@AsBytes()
	encryptedKey: Uint8Array = new Uint8Array();

	/**
	 * Public key used to encrypt the shared key
	 */
	@AsBytes()
	publicKey?: Uint8Array;
}

/**
 * The `SharedContainer` is used to securely share data between a number of
 * accessors using a shared-key encryption scheme where the content data is
 * encrypted using a randomly generated shared key that is then encrypted with
 * each accessors public key and stored along the encrypted data. Accessors can
 * then retrieve the shared key by decrypting it using their private key and
 * use it to recover the original data.
 */
export class SharedContainer extends BaseContainer {
	/** Parameters used to wrap the shared encryption key */
	@AsSerializable(RSAEncryptionParams)
	keyParams: RSAEncryptionParams = new RSAEncryptionParams();

	/** The ids and encrypted keys of all accessors */
	@AsSerializable(Accessor)
	accessors: Accessor[] = [];

	/**
	 * Unlocks the container using the id and private key of a given accessor.
	 * The id is used to look up the corresponding encrypted key while the
	 * private key is used to decrypt it.
	 */
	async unlock({ id, privateKey }: { id: string; privateKey: RSAPrivateKey }) {
		if (this._key) {
			// Container is already unlocked, no need to unlock it again
			return;
		}

		// Find accessor object with the same id
		const accessor = this.accessors.find((a) => a.id === id);

		if (!accessor || !accessor.encryptedKey) {
			// No corresponding accessor found.
			throw new Err(ErrorCode.MISSING_ACCESS, 'No appropriate accessor found.');
		}

		// Decrypt shared key using provided private key
		this._key = await getProvider().decrypt(
			privateKey,
			accessor.encryptedKey,
			this.keyParams
		);
	}

	/**
	 * Updates the containers accessors, generating a new shared key and encrypting
	 * it with the public keys of the provided **subjects**. Non-empty containers
	 * need to be unlocked first.
	 */
	async updateAccessors(subjects: { id: string; publicKey: RSAPublicKey }[]) {
		// Get existing data so we can reencrypt it after rotating the key
		let data: Uint8Array | null = null;

		// If the container already contains data, we need to reencrypt it after generating a new key
		if (this.encryptedData) {
			if (!this._key) {
				throw 'Non-empty containers need to be unlocked before accessors can be updated!';
			}
			data = await this.getData();
		}

		// Updating the accessors also requires generating a new shared key
		this._key = await getProvider().generateKey(new AESKeyParams());

		// Reencrypt data with new key
		if (data) {
			await this.setData(data);
		}

		// Encrypt the shared key with the public key of each accessor and store it along with their id
		this.accessors = await Promise.all(
			subjects.map(async ({ id, publicKey }) => {
				const accessor = new Accessor();
				accessor.id = id;
				accessor.publicKey = publicKey;
				accessor.encryptedKey = await getProvider().encrypt(
					publicKey,
					this._key!,
					this.keyParams
				);
				return accessor;
			})
		);
	}
}
