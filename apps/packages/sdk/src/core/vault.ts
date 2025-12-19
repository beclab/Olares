import { SharedContainer } from './container';
import { Storable } from './storage';
import { VaultItemCollection } from './collection';
import { AccountID, UnlockedAccount } from './account';
import { OrgInfo } from './org';
import { Exclude, AsDate } from './encoding';
import { Err } from './error';

/** Unique identifier for [[Vault]] objects */
export type VaultID = string;

/**
 * Container for securely storing a collection of [[VaultItem]]s. Vaults can be owned by a single
 * user ("private" vaults) or shared between multiple users ("shared" vaults). Shared vaults are
 * provisioned and managed through [[Org]]s.
 */
export class Vault extends SharedContainer implements Storable {
	get kind(): string {
		return 'vault';
	}
	/** unique identifier */
	id: VaultID = '';

	/** The [[Org]] this vault belongs to (if a shared vault) */
	org?: OrgInfo = undefined;

	/** Vault name */
	name = '';

	/** The vault owner (the [[Account]] that created this vault) */
	owner: AccountID = '';

	/** Time of creation */
	@AsDate()
	created = new Date(0);

	/** Time of last update */
	@AsDate()
	updated = new Date(0);

	/**
	 * Revision id used for ensuring continuity when synchronizing the vault
	 * object between client and server
	 */
	revision = '';

	/**
	 * A collection [[VaultItem]]s representing the senstive data store in this vault
	 *
	 * @secret
	 *
	 * **IMPORTANT**: This property is considered **secret**
	 * and should never stored or transmitted in plain text
	 */
	@Exclude()
	items = new VaultItemCollection();

	@Exclude()
	error?: Err;

	/**
	 * Convenience getter for getting a display label truncated to a certain maximum length
	 */
	get label() {
		return this.org ? `${this.org.name} / ${this.name}` : this.name;
	}

	/**
	 * Unlocks the vault with the given `account`, decrypting the data stored in the vault
	 * and populating the [[items]] property. For this to be successful, the `account` object
	 * needs to be unlocked and the account must have access to this vault.
	 */
	async unlock(account: UnlockedAccount) {
		if (!this.accessors.length) {
			await this.updateAccessors([account]);
			await this.commit();
		} else {
			await super.unlock(account);
			this.items.fromBytes(await this.getData());
		}
	}

	async lock() {
		await super.lock();
		this.items = new VaultItemCollection();
	}

	/**
	 * Commit changes to `items` by reencrypting the data.
	 */
	async commit() {
		await this.setData(this.items.toBytes());
	}

	/**
	 * Merges in changes from another `vault`. This requires both vaults to be unlocked.
	 */
	merge(vault: Vault) {
		this.items.merge(vault.items);
		this.name = vault.name;
		this.revision = vault.revision;
		this.org = vault.org;
		this.accessors = vault.accessors;
		this._key = vault._key;
		this.encryptedData = vault.encryptedData;
		this.updated = vault.updated;
	}

	toString() {
		return this.org ? `${this.org.name} / ${this.name}` : this.name;
	}

	clone() {
		const clone = super.clone();
		clone.items = this.items.clone();
		return clone;
	}
}
