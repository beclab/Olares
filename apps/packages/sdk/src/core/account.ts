import {
	stringToBytes,
	concatBytes,
	Serializable,
	AsBytes,
	AsDate,
	AsSet,
	Exclude,
	AsSerializable
} from './encoding';
import {
	RSAPublicKey,
	RSAPrivateKey,
	RSAKeyParams,
	HMACKey,
	HMACParams,
	HMACKeyParams
} from './crypto';
import { getCryptoProvider as getProvider } from './platform';
import { Err, ErrorCode } from './error';
import { PBES2Container } from './container';
import { Storable } from './storage';
import { VaultID } from './vault';
import { Org, OrgInfo } from './org';
import { TagInfo, VaultItemID } from './item';

/** Unique identifier for [[Account]] objects */
export type AccountID = string;

export class AccountSecrets extends Serializable {
	constructor({
		signingKey,
		privateKey,
		favorites,
		tags
	}: Partial<AccountSecrets> = {}) {
		super();
		Object.assign(this, { signingKey, privateKey, favorites, tags });
	}

	@AsBytes()
	signingKey!: Uint8Array;

	@AsBytes()
	privateKey!: Uint8Array;

	@AsSet()
	favorites = new Set<VaultItemID>();

	tags: TagInfo[] = [];
}

export class SecurityReportSettings extends Serializable {
	weakPasswords = true;
	reusedPasswords = true;
	compromisedPaswords = true;
	expiredItems = true;
}

export class NotificationSettings extends Serializable {
	failedLoginAttempts = true;
	newLogins = true;
}

export class AccountSettings extends Serializable {
	@AsSerializable(SecurityReportSettings)
	securityReport = new SecurityReportSettings();

	@AsSerializable(NotificationSettings)
	notifications = new NotificationSettings();
}

export const ACCOUNT_NAME_MAX_LENGTH = 100;
export const ACCOUNT_EMAIL_MAX_LENGTH = 255;

/**
 * The `Account` object represents an individual Padloc user and holds general
 * account information as well as cryptographic keys necessary for accessing
 * [[Vaults]] and signing/verifying [[Org]]anization details.
 *
 * The [[privateKey]] and [[signingKey]] properties are considered secret and
 * therefore need to be encrypted at rest. For this, the [[Account]] object
 * serves as a [[PBESContainer]] which is unlocked by the users **master
 * password**.
 */
export class Account extends PBES2Container implements Storable {
	get kind(): string {
		return 'account';
	}
	/** Unique account ID */
	id: AccountID = '';

	/** The users email address */
	did = '';

	/** The users display name */
	name = '';

	local = false;

	/** When the account was created */
	@AsDate()
	created = new Date();

	/** when the account was last updated */
	@AsDate()
	updated = new Date();

	/** The accounts public key */
	@AsBytes()
	publicKey!: RSAPublicKey;

	/**
	 * The accounts private key
	 *
	 * @secret
	 * **IMPORTANT**: This property is considered **secret**
	 * and should never stored or transmitted in plain text
	 */
	@Exclude()
	privateKey?: RSAPrivateKey;

	/**
	 * HMAC key used for signing and verifying organization details
	 *
	 * **IMPORTANT**: This property is considered **secret**
	 * and should never stored or transmitted in plain text
	 *
	 * @secret
	 */
	@Exclude()
	signingKey?: HMACKey;

	/** ID of the accounts main or "private" [[Vault]]. */
	mainVault: {
		id: VaultID;
		name?: string;
		revision?: string;
	} = { id: '' };

	/** All organizations this account is a member of */
	orgs: OrgInfo[] = [];

	/**
	 * Revision id used for ensuring continuity when synchronizing the account
	 * object between client and server
	 */
	revision?: string;

	kid = '';

	@Exclude()
	favorites = new Set<VaultItemID>();

	@Exclude()
	tags: TagInfo[] = [];

	/** Application Settings */
	@AsSerializable(AccountSettings)
	settings = new AccountSettings();

	/**
	 * Whether or not this Account object is current "locked" or, in other words,
	 * whether the `privateKey` and `signingKey` properties have been decrypted.
	 */
	get locked(): boolean {
		return !this.privateKey;
	}

	get masterKey() {
		return this._key;
	}

	set masterKey(key: Uint8Array | undefined) {
		this._key = key;
	}

	/**
	 * Generates the accounts [[privateKey]], [[publicKey]] and [[signingKey]] and
	 * encrypts [[privateKey]] and [[singingKey]] using the master password.
	 */
	async initialize(password: string) {
		const { publicKey, privateKey } = await getProvider().generateKey(
			new RSAKeyParams()
		);
		this.publicKey = publicKey;
		this.privateKey = privateKey;
		this.signingKey = await getProvider().generateKey(new HMACKeyParams());
		await this.setPassword(password);
	}

	/** Updates the master password by reencrypting the [[privateKey]] and [[signingKey]] properties */
	async setPassword(password: string) {
		if (this.encryptedData && !this._key) {
			throw 'Account has to be unlocked first.';
		}
		await this._deriveAndSetKey(password);
		await this.commitSecrets();
		this.updated = new Date();
	}

	/**
	 * "Unlocks" the account by decrypting and extracting [[privateKey]] and
	 * [[signingKey]] from [[encryptedData]]
	 */
	async unlock(password: string) {
		await super.unlock(password);
		await this._loadSecrets();
	}

	/**
	 * Unlocks the account by providing the encryption key directly rather than
	 * deriving it from the master password
	 */
	async unlockWithMasterKey(key: Uint8Array) {
		this._key = key;
		await this._loadSecrets();
	}

	/**
	 * "Locks" the account by deleting all sensitive data from the object
	 */
	lock() {
		super.lock();
		delete this.privateKey;
		delete this.signingKey;
		this.favorites.clear();
		this.tags = [];
	}

	clone() {
		const clone = super.clone();
		clone.copySecrets(this);
		return clone;
	}

	// toString() {
	//     return this.name || this.email;
	// }
	toString() {
		return this.name || this.id;
	}

	/**
	 * Creates a signature that can be used later to verify an organizations id and public key
	 */
	async signOrg({ id, publicKey }: { id: string; publicKey: Uint8Array }) {
		if (!this.signingKey) {
			throw 'Signing key not available!';
		}
		return getProvider().sign(
			this.signingKey,
			concatBytes([stringToBytes(id), publicKey], 0x00),
			new HMACParams()
		);
	}

	/**
	 * Verifies an organizations id an public key, using the signature stored
	 * in the [[Member]] object associated with the account.
	 */
	async verifyOrg(org: Org): Promise<void> {
		if (!this.signingKey) {
			throw 'Account needs to be unlocked first';
		}

		if (!org.publicKey) {
			throw 'Org has not been initialized yet!';
		}

		const member = org.getMember(this);

		if (!member) {
			throw new Err(
				ErrorCode.VERIFICATION_ERROR,
				'Account is not a member.'
			);
		}

		// const verified =
		// 	member.orgSignature &&
		// 	(await getProvider().verify(
		// 		this.signingKey,
		// 		member.orgSignature,
		// 		concatBytes([stringToBytes(org.id), org.publicKey!], 0x00),
		// 		new HMACParams()
		// 	));

		// if (!verified) {
		// 	throw new Err(
		// 		ErrorCode.VERIFICATION_ERROR,
		// 		`Failed to verify public key of ${org.name}!`
		// 	);
		// }
	}

	async toggleFavorite(id: VaultItemID, favorite: boolean) {
		favorite ? this.favorites.add(id) : this.favorites.delete(id);
		await this.commitSecrets();
	}

	copySecrets(account: Account) {
		this.privateKey = account.privateKey;
		this.signingKey = account.signingKey;
		this.favorites = account.favorites;
		this.tags = account.tags;
		this._key = account._key;
	}

	validate() {
		// if (this.email.length > ACCOUNT_EMAIL_MAX_LENGTH) {
		//     return false;
		// }

		if (this.name.length > ACCOUNT_NAME_MAX_LENGTH) {
			return false;
		}

		return true;
	}

	async commitSecrets() {
		const secrets = new AccountSecrets(this as UnlockedAccount);
		await this.setData(secrets.toBytes());
	}

	private async _loadSecrets() {
		const secrets = new AccountSecrets().fromBytes(await this.getData());
		if (!secrets.favorites) {
			secrets.favorites = new Set<VaultItemID>();
		}
		if (!secrets.tags) {
			secrets.tags = [];
		}
		Object.assign(this, secrets);
	}
}

export interface UnlockedAccount extends Account {
	privateKey: Uint8Array;
	signingKey: Uint8Array;
}
