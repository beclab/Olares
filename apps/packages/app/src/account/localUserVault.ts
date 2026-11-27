import { Serializable, Exclude, AsDate, AsSerializable } from './serialization';
import { TerminusDefaultDomain, compareOlaresVersion } from '@bytetrade/core';
import { PBES2Container } from './passwordStorage';

export type UserItemID = string;
export type LocalUserVaultID = string;

export enum SSHPassworResetType {
	RESETED = 0,
	NEED_RESET = 1,
	SKIP = 2
}

export class MnemonicItem extends Serializable {
	constructor(vals: Partial<MnemonicItem> = {}) {
		super();
		Object.assign(this, vals);
	}

	id: UserItemID = ''; //did

	mnemonic = '';

	@AsDate()
	updated: Date = new Date();

	get kind(): string {
		return 'mnemonicitem';
	}
}

export class MnemonicItemCollection
	extends Serializable
	implements Iterable<MnemonicItem>
{
	/** Number of items in this VaultItemCollection */
	get size() {
		return this._items.size;
	}

	get kind(): string {
		return 'mnemonicitemcollection';
	}

	private _items: Map<string, MnemonicItem>;

	constructor(items: MnemonicItem[] = []) {
		super();
		this._items = new Map(
			items.map((item) => [item.id, item] as [string, MnemonicItem])
		);
	}

	/** Get an item with a given `id` */
	get(id: string) {
		return this._items.get(id) || null;
	}

	/**
	 * Updates one or more items based on their id. If no item with the same id
	 * exists, the item will be added to the collection
	 */
	update(...items: MnemonicItem[]) {
		for (const item of items) {
			item.updated = new Date();
			this._items.set(item.id, item);
		}
	}

	/**
	 * Removes one or more items based on their id.
	 */
	remove(...items: MnemonicItem[]) {
		for (const item of items) {
			this._items.delete(item.id);
		}
	}

	protected _toRaw(version: string) {
		return {
			items: Array.from(this).map((item) => item.toRaw(version))
		};
	}

	protected _fromRaw({ items }: any) {
		this._items = new Map(
			items.map(
				(item: any) =>
					[item.id, new MnemonicItem().fromRaw(item)] as [string, MnemonicItem]
			)
		);
	}

	[Symbol.iterator]() {
		return this._items.values();
	}
}

export class UserItem extends Serializable {
	get kind() {
		return 'useritem';
	}

	constructor(vals: Partial<UserItem> = {}) {
		super();
		Object.assign(this, vals);
	}

	id: UserItemID = ''; //did

	/** icon to be displayed for this item */
	name = ''; // terminus name

	setup_finished = false;

	terminus_activate_status = 'wait_activate_vault';

	wizard = '';

	localMachine = '';

	offline_mode = false;

	tailscale_activated = false;

	isLocal = false;

	url = '';

	ip = '';

	binding = false;

	os_version? = '';

	/** Date and time of last update */
	@AsDate()
	updated: Date = new Date();

	access_token = '';
	refresh_token = '';
	passed_fa2 = false;
	session_id = '';

	cloud_id = '';
	cloud_token = '';
	cloud_expired = 0;
	terminus_id = '';
	olares_id = '';

	owner_ssh_is_default = SSHPassworResetType.RESETED;

	get olares_device_id() {
		if (!!this.olares_id) {
			return this.olares_id;
		}
		return this.terminus_id;
	}

	get local_name() {
		const array: string[] = this.name.split('@');
		return array[0];
	}

	get domain_name() {
		const array: string[] = this.name.split('@');
		if (array.length == 2) {
			return array[1];
		} else {
			return TerminusDefaultDomain;
		}
	}

	get terminus_url() {
		return this.get_module_url('', 'https', '');
	}

	get local_terminus_url() {
		const array: string[] = this.name.split('@');
		if (this.isLargeVersion12_4) {
			return 'http://' + array[0] + '.' + 'olares.local';
		} else {
			if (array.length == 1) {
				return 'https://' + 'local.' + array[0] + '.' + TerminusDefaultDomain;
			}
			if (array.length == 2) {
				return 'https://' + 'local.' + array[0] + '.' + array[1];
			}
		}
		return this.terminus_url;
	}

	get vault_url() {
		return this.get_module_url('vault', 'https', '/server');
	}

	get auth_url() {
		return this.get_module_url(
			'auth',
			'https',
			'/',
			this.terminus_activate_status != 'wait_reset_password'
		);
	}

	get desktop_target_url() {
		return this.get_module_url(
			'desktop',
			'https',
			'/',
			this.terminus_activate_status != 'wait_reset_password'
		);
	}

	get vault_target_url() {
		return this.get_module_url(
			'vault',
			'https',
			'/server',
			this.terminus_activate_status != 'wait_reset_password'
		);
	}

	get isLargeVersion12() {
		if (this.os_version) {
			return compareOlaresVersion(this.os_version, '1.12.0-0').compare >= 0;
		}
		return false;
	}

	get isLargeVersion12_1() {
		if (this.os_version) {
			return compareOlaresVersion(this.os_version, '1.12.1-0').compare >= 0;
		}
		return false;
	}

	get isLargeVersion12_2() {
		if (this.os_version) {
			return compareOlaresVersion(this.os_version, '1.12.2-0').compare >= 0;
		}
		return false;
	}

	get isLargeVersion12_3() {
		if (this.os_version) {
			return compareOlaresVersion(this.os_version, '1.12.3-0').compare >= 0;
		}
		return false;
	}

	get isLargeVersion12_4() {
		if (this.os_version) {
			return compareOlaresVersion(this.os_version, '1.12.4-0').compare >= 0;
		}
		return false;
	}

	get isLargeVersion12_6() {
		if (this.os_version) {
			return compareOlaresVersion(this.os_version, '1.12.6-0').compare >= 0;
		}
		return false;
	}

	get local_url() {
		if (this.isLargeVersion12) {
			return '';
		}

		if (this.isLocal) {
			return 'local.';
		}
		return '';
	}

	get_module_url(
		module: string,
		protocol = 'https',
		suffix = '/',
		use_local = true
	) {
		const array: string[] = this.name.split('@');
		const module_prefix = module ? module + '.' : '';

		if (protocol.endsWith(':')) {
			protocol = protocol.slice(0, -1);
		}

		if (use_local && this.isLargeVersion12_4 && this.isLocal) {
			const target = protocol.startsWith('https')
				? 'http'
				: protocol.startsWith('wss')
				? 'ws'
				: protocol;
			return (
				target + '://' + module_prefix + array[0] + '.olares.local' + suffix
			);
		}

		if (array.length == 2) {
			return (
				protocol +
				'://' +
				module_prefix +
				(use_local ? this.local_url : '') +
				array[0] +
				'.' +
				array[1] +
				suffix
			);
		} else {
			return (
				protocol +
				'://' +
				module_prefix +
				(use_local ? this.local_url : '') +
				array[0] +
				'.' +
				TerminusDefaultDomain +
				suffix
			);
		}
	}

	get local_default_host() {
		return this.local_name + '.olares.local';
	}
}

/**
 * A collection of vault items items, used for consolidating changes made independently
 * across multiple instances through "merging".
 */
export class UserItemCollection
	extends Serializable
	implements Iterable<UserItem>
{
	get kind() {
		return 'useritemcollection';
	}

	/** Number of items in this VaultItemCollection */
	get size() {
		return this._items.size;
	}

	private _items: Map<string, UserItem>;

	constructor(items: UserItem[] = []) {
		super();
		this._items = new Map(
			items.map((item) => [item.id, item] as [string, UserItem])
		);
	}

	/** Get an item with a given `id` */
	get(id: string) {
		return this._items.get(id) || null;
	}

	/**
	 * Updates one or more items based on their id. If no item with the same id
	 * exists, the item will be added to the collection
	 */
	update(...items: UserItem[]) {
		for (const item of items) {
			item.updated = new Date();
			this._items.set(item.id, item);
		}
	}

	/**
	 * Removes one or more items based on their id.
	 */
	remove(...items: UserItem[]) {
		for (const item of items) {
			this._items.delete(item.id);
		}
	}

	protected _toRaw(version: string) {
		return {
			items: Array.from(this).map((item) => item.toRaw(version))
		};
	}

	protected _fromRaw({ items }: any) {
		this._items = new Map(
			items.map(
				(item: any) =>
					[item.id, new UserItem().fromRaw(item)] as [string, UserItem]
			)
		);
	}

	[Symbol.iterator]() {
		return this._items.values();
	}
}

/**
 * Local account metadata and protected recovery phrases for this browser.
 */
export class LocalUserVault extends PBES2Container {
	get kind() {
		return 'localuservault';
	}

	/** unique identifier */
	id: LocalUserVaultID = '';

	/** Vault name */
	name = '';

	/** Time of creation */
	@AsDate()
	created = new Date(0);

	// /** Time of last update */
	@AsDate()
	updated = new Date(0);

	@Exclude()
	mnemonics = new MnemonicItemCollection();

	/**
	 * A collection [[UserItem]]s representing the senstive data store in this vault
	 *
	 * @secret
	 *
	 * **IMPORTANT**: This property is considered **secret**
	 * and should never stored or transmitted in plain text
	 */
	@AsSerializable(UserItemCollection)
	items = new UserItemCollection();

	async setPassword(password: string) {
		if (this.encryptedData && !this._key) {
			throw 'Account has to be unlocked first.';
		}
		await this._deriveAndSetKey(password);
		await this.setData(this.mnemonics.toBytes());
		this.updated = new Date();
	}

	async unlock(password: string) {
		const data = await this.unlockData(password);
		try {
			if (data) {
				const mnemonics = new MnemonicItemCollection().fromBytes(data);
				if (mnemonics.size > 0 && this.items.size === 0) {
					this.items.fromBytes(data);
				}
				this.mnemonics = mnemonics;
			}
		} catch (error) {
			await this.lock();
			throw error;
		}
	}

	async lock() {
		await super.lock();
		this.mnemonics = new MnemonicItemCollection();
	}

	get locked() {
		return !this._key;
	}

	async commit() {
		await this.setData(this.mnemonics.toBytes());
	}
}
