import {
	stringToBytes,
	Serializable,
	concatBytes,
	AsSerializable,
	AsBytes,
	AsDate,
	Exclude
} from './encoding';
import {
	RSAPrivateKey,
	RSAPublicKey,
	AESKey,
	RSAKeyParams,
	AESKeyParams,
	RSASigningParams
} from './crypto';
import { getCryptoProvider as getProvider } from './platform';
import { SharedContainer } from './container';
import { Err, ErrorCode } from './error';
import { Storable } from './storage';
import { Vault, VaultID } from './vault';
import { Account, AccountID, UnlockedAccount } from './account';
import { Invite, InviteID } from './invite';

/** Role of a member within an organization, each associated with certain priviliges */
export enum OrgRole {
	/**
	 * Organization owner. Can manage members, groups and vaults.  Owners have
	 * access to the secret [[Org.privateKey]] and [[Org.invitesKey]]
	 * properties.
	 */
	Owner,
	/**
	 * Organization admin. Can manage groups and vaults.
	 */
	Admin,
	/**
	 * Basic organization member. Can read public organization data and read/write
	 * certain [[Vault]]s they have been assigned to directly or via [[Group]]s.
	 */
	Member,
	/**
	 * Suspended members can read public organization data and access [[Vaults]] they
	 * have been assigned to, but are excluded from any updates to those vaults.
	 * Member information (like public key and email address) of suspended members
	 * are considered unverified, and need to be updated and verified via a
	 * membership confirmation [[Invite]].
	 * @deprecated Use `OrgMemberStatus.Suspended` instead
	 */
	Suspended
}

export enum OrgMemberStatus {
	Provisioned = 'provisioned',
	Active = 'active',
	Suspended = 'suspended'
}

/**
 * Represents an [[Account]]s membership to an [[Org]]
 */
export class OrgMember extends Serializable {
	get kind(): string {
		return 'orgmember';
	}

	get id() {
		return this.accountId;
	}

	/** id of the corresponding [[Account]] */
	accountId?: AccountID = undefined;

	/** name of the corresponding [[Account]] */
	name = '';

	/** email address of the corresponding [[Account]] */
	did = '';

	/** public key of the corresponding [[Account]] */
	@AsBytes()
	publicKey?: RSAPublicKey;

	/** signature used by other members to verify [[id]], [[email]] and [[publicKey]] */
	@AsBytes()
	signature?: Uint8Array;

	/** signature used by the member to verify [[Org.id]] and [[Org.publickey]] of the organization */
	@AsBytes()
	orgSignature?: Uint8Array;

	/** vaults assigned to this member */
	vaults: {
		id: VaultID;
		readonly: boolean;
	}[] = [];

	/** the members organization role */
	role: OrgRole = OrgRole.Member;

	status: OrgMemberStatus = OrgMemberStatus.Active;

	/** time the member was last updated */
	@AsDate()
	updated = new Date(0);

	constructor({
		accountId,
		name,
		did,
		publicKey,
		signature,
		orgSignature,
		role,
		updated,
		status
	}: Partial<OrgMember> = {}) {
		super();

		Object.assign(this, {
			accountId,
			name,
			did,
			publicKey,
			signature,
			orgSignature,
			updated,
			status
		});
		this.role =
			typeof role !== 'undefined' && role in OrgRole
				? role
				: OrgRole.Member;
	}
}

export interface ActiveOrgMember extends OrgMember {
	id: string;
	status: OrgMemberStatus.Active;
	accountId: AccountID;
	publicKey: Uint8Array;
	signature: Uint8Array;
	orgSignature: Uint8Array;
}

/**
 * A group of members, used to manage [[Vault]] access for multiple members at once.
 */
export class Group extends Serializable {
	constructor(vals: Partial<Group> = {}) {
		super();
		Object.assign(this, vals);
	}

	/** display name */
	name = '';
	/** members assigned to this group */
	members: { did: string; accountId?: AccountID }[] = [];
	/** [[Vault]]s assigned to this group */
	vaults: {
		id: VaultID;
		readonly: boolean;
	}[] = [];
}

/** Unique identifier for [[Org]]s */
export type OrgID = string;

export class OrgSecrets extends Serializable {
	constructor({ invitesKey, privateKey }: Partial<OrgSecrets> = {}) {
		super();
		Object.assign(this, { invitesKey, privateKey });
	}

	@AsBytes()
	invitesKey!: Uint8Array;

	@AsBytes()
	privateKey!: Uint8Array;
}

export interface OrgInfo {
	id: OrgID;
	name: string;
	owner?: {
		did: string;
		accountId?: AccountID;
	};
	revision: string;
}

export class ScimSettings extends Serializable {
	@AsBytes()
	secret!: Uint8Array;

	secretToken = '';
	url = '';
}

export class OrgDirectorySettings extends Serializable {
	syncProvider: 'scim' | 'none' = 'none';
	syncGroups = true;
	syncMembers = true;

	@AsSerializable(ScimSettings)
	scim?: ScimSettings;
}

export const ORG_NAME_MAX_LENGTH = 100;

/**
 * Organizations are the central component of Padlocs secure data sharing architecture.
 *
 * All shared [[Vault]]s are provisioned and managed in the context of an organization,
 * while the [[Org]] class itself is responsible for managing, signing and verifying
 * public keys, identities and priviliges for all of it's members.
 *
 * Vaults can be assigned to members direcly or indirectly through [[Group]]s. In both
 * cases, this access can be declared *readonly*.
 *
 * Before being added to an organization, members need to go throug a key exchange
 * procedure designed to allow verification of organization and member details
 * by both parties. See [[Invite]] class for details.
 *
 * The [[privateKey]] and [[invitesKey]] properties are considered secret and are only
 * accessible to members with the [[OrgRole.Owner]] role. To protect this information
 * from unauthorized access, [[Org]] extends the [[SharedContainer]] class, encrypting
 * this data at rest.
 *
 * #### Organization Structure
 * ```
 * ┌──────────────┐            ┌──────────────┐            ┌──────────────┐
 * │              │           ╱│              │╲           │              │
 * │   Account    │┼─────────○─│  Membership  │──┼────────┼│ Organization │
 * │              │           ╲│              │╱           │              │
 * └──────────────┘            └───┬──────┬───┘            └──────────────┘
 *                                ╲│╱    ╲│╱                       ┼
 *                                 ○      ○                        ○
 *                                 │      │                       ╱│╲
 *                                 │      │                ┌──────────────┐
 *                                 │      │               ╱│              │
 *                                 │      └──────────────○─│    Group     │
 *                                 │                      ╲│              │
 *                                 ○                       └──────────────┘
 *                                ╱│╲                             ╲│╱
 *                         ┌──────────────┐                        ○
 *                         │              │╲                       │
 *                         │ Shared Vault │─○──────────────────────┘
 *                         │              │╱
 *                         └──────────────┘
 * ```
 */
export class Org extends SharedContainer implements Storable {
	get kind(): string {
		return 'org';
	}
	/** Unique identier */
	id: OrgID = '';

	/** Organization name */
	name = '';

	/** Creation date */
	@AsDate()
	created: Date = new Date();

	/** Last updated */
	@AsDate()
	updated: Date = new Date();

	/** Public key used for verifying member signatures */
	@AsBytes()
	publicKey?: RSAPublicKey;

	/**
	 * Private key used for signing member details
	 *
	 * @secret
	 * **IMPORTANT**: This property is considered **secret**
	 * and should never stored or transmitted in plain text
	 */
	@Exclude()
	privateKey?: RSAPrivateKey;

	/**
	 * AES key used as encryption key for [[Invite]]s
	 *
	 * @secret
	 * **IMPORTANT**: This property is considered **secret**
	 * and should never stored or transmitted in plain text
	 */
	@Exclude()
	invitesKey?: AESKey;

	/**
	 * Minimum accepted update time for organization members.
	 * Any members with a [[OrgMember.updated]] value lower than
	 * this should be considered invalid.
	 *
	 * In order to prevent an attacker from rolling back this value, all
	 * clients should verify that updated organization object always have a
	 * [[Org.minMemberUpdated]] value equal to or higher than the previous one.
	 */
	@AsDate()
	minMemberUpdated: Date = new Date();

	/** Parameters for creating member signatures */
	@AsSerializable(RSASigningParams)
	signingParams = new RSASigningParams();

	/** Array of organization members */
	@AsSerializable(OrgMember)
	members: OrgMember[] = [];

	/** This organization's [[Group]]s. */
	@AsSerializable(Group)
	groups: Group[] = [];

	/** Shared [[Vault]]s owned by this organization */
	vaults: { id: VaultID; name: string; revision?: string }[] = [];

	/** Pending [[Invite]]s */
	@AsSerializable(Invite)
	invites: Invite[] = [];

	@AsSerializable(OrgDirectorySettings)
	directory: OrgDirectorySettings = new OrgDirectorySettings();

	/**
	 * Revision id used for ensuring continuity when synchronizing the account
	 * object between client and server
	 */
	revision = '';

	/** [[Account]] which created this organization */
	get owner() {
		const owner = this.members.find(
			(member) => member.role === OrgRole.Owner
		);
		return owner
			? {
					accountId: owner.accountId,
					did: owner.did
			  }
			: undefined;
	}

	get info(): OrgInfo {
		return {
			id: this.id,
			name: this.name,
			owner: this.owner,
			revision: this.revision
		};
	}

	/** Whether the given [[Account]] is an [[OrgRole.Owner]] */
	isOwner({ did }: { did: string }) {
		return this.owner?.did === did;
	}

	/** Whether the given [[Account]] is an [[OrgRole.Admin]] */
	isAdmin(m: { did: string }) {
		const member = this.getMember(m);
		return !!member && member.role <= OrgRole.Admin;
	}

	/** Whether the given [[Account]] is currently suspended */
	isSuspended(m: { did: string }) {
		const member = this.getMember(m);
		return !!member && member.status === OrgMemberStatus.Suspended;
	}

	/** Get the [[OrgMember]] object for this [[Account]] */
	getMember({
		did
	}:
		| { did: string; accountId?: AccountID }
		| { accountId: AccountID; did?: string }) {
		return this.members.find((m) => m.did === did);
	}

	/** Whether the given [[Account]] is an organization member */
	isMember(acc: { did: string }) {
		return !!this.getMember(acc);
	}

	/** Get group with the given `name` */
	getGroup(name: string) {
		return [...this.groups].find((g) => g.name === name);
	}

	/** Get all members of a given `group` */
	getMembersForGroup(group: Group): OrgMember[] {
		return (
			group.members
				.map((m) => this.getMember(m))
				// Filter out undefined members
				.filter((m) => !!m) as OrgMember[]
		);
	}

	/** Get all [[Group]]s the given [[Account]] is a member of */
	getGroupsForMember({ did }: { did: string }) {
		return this.groups.filter((g) => g.members.some((m) => m.did === did));
	}

	/** Get all groups assigned to a given [[Vault]] */
	getGroupsForVault({ id }: { id: VaultID }): Group[] {
		return this.groups.filter((group) =>
			group.vaults.some((v) => v.id === id)
		);
	}

	/** Get all members directly assigned to a given [[Vault]] */
	getMembersForVault({ id }: { id: VaultID }): ActiveOrgMember[] {
		return this.members.filter(
			(member) =>
				member.status === OrgMemberStatus.Active &&
				member.accountId &&
				member.publicKey &&
				member.vaults.some((v) => v.id === id)
		) as ActiveOrgMember[];
	}

	/** Get all membes that have acess to a given `vault`, either directly or through a [[Group]] */
	getAccessors(vault: Vault): ActiveOrgMember[] {
		const results = new Set<ActiveOrgMember>(
			this.getMembersForVault(vault)
		);

		for (const group of this.getGroupsForVault(vault)) {
			for (const m of group.members) {
				const member = this.getMember(m);
				if (
					member &&
					member.status === OrgMemberStatus.Active &&
					member.publicKey &&
					member.accountId
				) {
					results.add(member as ActiveOrgMember);
				}
			}
		}

		return [...results];
	}

	/** Get all vaults the given member has access to */
	getVaultsForMember(acc: { did: string }) {
		const member = this.getMember(acc);

		if (!member) {
			return [];
		}

		const results = new Set<VaultID>(member.vaults.map((v) => v.id));

		for (const group of this.getGroupsForMember(member)) {
			for (const vault of group.vaults) {
				results.add(vault.id);
			}
		}

		return this.vaults.filter((v) => results.has(v.id));
	}

	/** Check whether the given `account` has read access to a `vault` */
	canRead(vault: { id: VaultID }, account: { did: string }) {
		const member = this.getMember(account);

		return (
			member &&
			[member, ...this.getGroupsForMember(member)].some(({ vaults }) =>
				vaults.some((v) => v.id === vault.id)
			)
		);
	}

	/** Check whether the given `account` has write access to a `vault` */
	canWrite(vault: { id: VaultID }, acc: { did: string }) {
		const member = this.getMember(acc);

		return (
			member &&
			member.status === OrgMemberStatus.Active &&
			[member, ...this.getGroupsForMember(member)].some(({ vaults }) =>
				vaults.some((v) => v.id === vault.id && !v.readonly)
			)
		);
	}

	/** Get the invite with the given `id` */
	getInvite(id: InviteID) {
		return this.invites.find((inv) => inv.id === id);
	}

	/** Remove an invite */
	removeInvite({ id }: Invite) {
		this.invites = this.invites.filter((inv) => inv.id !== id);
	}

	/**
	 * Initializes the organization, generating [[publicKey]], [[privateKey]],
	 * and [[invitesKey]] and adding the given `account` as the organization
	 * owner.
	 */
	async initialize(account: Account) {
		// Update access to keypair
		await this.updateAccessors([account]);

		// Generate cryptographic keys
		await this.generateKeys();

		// Set minimum date for member update times
		this.minMemberUpdated = new Date();

		const orgSignature = await account.signOrg({
			id: this.id,
			publicKey: this.publicKey!
		});

		await this.addOrUpdateMember(
			new OrgMember({
				accountId: account.id,
				name: account.name,
				did: account.did,
				publicKey: account.publicKey,
				orgSignature,
				role: OrgRole.Owner,
				updated: new Date(),
				status: OrgMemberStatus.Active
			})
		);
	}

	/**
	 * Generates a new [[publicKey]], [[privateKey]] and [[invitesKey]] and
	 * encrypts the latter two
	 */
	async generateKeys() {
		this.invitesKey = await getProvider().generateKey(new AESKeyParams());
		const { privateKey, publicKey } = await getProvider().generateKey(
			new RSAKeyParams()
		);
		this.privateKey = privateKey;
		this.publicKey = publicKey;
		await this.setData(new OrgSecrets(this as UnlockedOrg).toBytes());
	}

	/**
	 * Regenerates all cryptographic keys and updates all member signatures
	 */
	async rotateKeys(force = false) {
		if (!force) {
			// Verify members and groups with current public key
			await this.verifyAll();
		}

		// Rotate Org key pair
		await this.generateKeys();

		// Rotate org encryption key
		await this.updateAccessors([
			this.getMember(this.owner!) as ActiveOrgMember
		]);

		// Re-sign all members
		await Promise.all(
			this.members
				.filter((m) => m.status === OrgMemberStatus.Active)
				.map((m) => this.addOrUpdateMember(m))
		);
	}

	/**
	 * "Unlocks" the organization, granting access to the organizations
	 * [[privateKey]] and [[invitesKey]] properties.
	 */
	async unlock(account: UnlockedAccount) {
		await super.unlock(account);
		if (this.encryptedData) {
			const secrets = new OrgSecrets().fromBytes(await this.getData());
			Object.assign(this, secrets);
		}
	}

	lock() {
		super.lock();
		delete this.privateKey;
		delete this.invitesKey;
		this.invites.forEach((invite) => invite.lock);
	}

	/**
	 * Signs the `member`s public key, id, role and email address so they can be verified later
	 */
	async sign(member: OrgMember): Promise<OrgMember> {
		if (!this.privateKey) {
			throw 'Organisation needs to be unlocked first.';
		}

		if (!member.publicKey) {
			throw 'The member needs to be assigned a public key first.';
		}

		member.signature = await getProvider().sign(
			this.privateKey,
			concatBytes(
				[
					stringToBytes(member.accountId || ''),
					stringToBytes(member.did),
					new Uint8Array([member.role]),
					member.publicKey,
					stringToBytes(member.updated.toISOString())
				],
				0x00
			),
			this.signingParams
		);
		return member;
	}

	/**
	 * Verifies the `member`s public key, id, role and email address.
	 * Throws if verification fails.
	 */
	async verify(member: OrgMember): Promise<void> {
		if (!member.signature || !member.publicKey || !this.publicKey) {
			throw new Err(
				ErrorCode.VERIFICATION_ERROR,
				'No signed public key provided!'
			);
		}

		const verified =
			member.updated >= this.minMemberUpdated &&
			(await getProvider().verify(
				this.publicKey,
				member.signature,
				concatBytes(
					[
						stringToBytes(member.accountId || ''),
						stringToBytes(member.did),
						new Uint8Array([member.role]),
						member.publicKey,
						stringToBytes(member.updated.toISOString())
					],
					0x00
				),
				this.signingParams
			));

		if (!verified) {
			throw new Err(
				ErrorCode.VERIFICATION_ERROR,
				`Failed to verify public key of ${member.name}!`
			);
		}
	}

	/**
	 * Verify all provided `members`, throws if verification fails for any of them.
	 */
	async verifyAll(
		members: OrgMember[] = this.members.filter(
			(m) => m.status === OrgMemberStatus.Active
		)
	) {
		// Verify public keys for members and groups
		await Promise.all(members.map(async (obj) => this.verify(obj)));
	}

	/**
	 * Adds a member to the organization, or updates the existing member with the same id.
	 */
	async addOrUpdateMember({
		accountId,
		name,
		did,
		publicKey,
		orgSignature,
		role = OrgRole.Member,
		status = OrgMemberStatus.Active
	}: {
		accountId?: string;
		name: string;
		did: string;
		publicKey?: Uint8Array;
		orgSignature?: Uint8Array;
		role?: OrgRole;
		status?: OrgMemberStatus;
	}) {
		if (!this.privateKey) {
			throw 'Organisation needs to be unlocked first.';
		}

		const existing = this.getMember({ did, accountId });
		const updated = new Date();

		if (existing) {
			Object.assign(existing, {
				name,
				did,
				accountId,
				publicKey,
				orgSignature,
				role,
				status,
				updated
			});
			await this.sign(existing);
		} else {
			this.members.push(
				await this.sign(
					new OrgMember({
						accountId,
						name,
						did,
						publicKey,
						orgSignature,
						role,
						status,
						updated
					})
				)
			);
		}
	}

	/**
	 * Removes a member from the organization
	 */
	async removeMember(member: { did: string }, reSignMembers = true) {
		if (reSignMembers && !this.privateKey) {
			throw 'Organisation needs to be unlocked first.';
		}

		// Remove member from all groups
		for (const group of this.getGroupsForMember(member)) {
			group.members = group.members.filter((m) => m.did !== member.did);
		}

		// Remove member
		this.members = this.members.filter((m) => m.did !== member.did);

		if (reSignMembers) {
			// Verify remaining members (since we're going to re-sign them)
			await this.verifyAll();

			// Bump minimum update date
			this.minMemberUpdated = new Date();

			// Re-sign all members
			await Promise.all(
				this.members
					.filter((m) => m.status === OrgMemberStatus.Active)
					.map((m) => this.addOrUpdateMember(m))
			);
		}
	}

	/**
	 * Transfers organization ownership to a different member
	 */
	async makeOwner(member: { did: string }) {
		if (!this.privateKey) {
			throw 'Organisation needs to be unlocked first.';
		}

		// Verify members and groups with current public key
		await this.verifyAll();

		const newOwner = this.getMember(member);
		const existingOwner = this.getMember(this.owner!)!;

		if (!newOwner || !existingOwner) {
			throw 'New and/or existing owner not found.';
		}

		newOwner.role = OrgRole.Owner;
		existingOwner.role = OrgRole.Admin;

		await this.addOrUpdateMember(newOwner);
		await this.addOrUpdateMember(existingOwner);

		await this.updateAccessors([newOwner as ActiveOrgMember]);
	}

	validate() {
		if (this.name.length > ORG_NAME_MAX_LENGTH) {
			return false;
		}

		return true;
	}

	toString() {
		return this.name;
	}
}

export interface UnlockedOrg extends Org {
	privateKey: Uint8Array;
	invitesKey: Uint8Array;
}
