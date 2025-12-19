import { PBKDF2Params, HMACParams, HMACKey } from './crypto';
import { getCryptoProvider as getProvider } from './platform';
import { SimpleContainer } from './container';
import {
	stringToBytes,
	bytesToString,
	bytesToHex,
	marshal,
	unmarshal,
	concatBytes,
	Serializable,
	AsSerializable,
	AsDate,
	AsBytes
} from './encoding';
import { Account, AccountID } from './account';
import { Org, OrgID } from './org';
import { uuid } from './util';
import { Err, ErrorCode } from './error';

export type InvitePurpose = 'join_org' | 'confirm_membership';

/**
 * Unique identifier for [[Invite]]s.
 */
export type InviteID = string;

class OrgInfo extends Serializable {
	id: OrgID = '';
	name = '';

	@AsBytes()
	publicKey!: Uint8Array;

	/**
	 * Signature created using the HMAC key derived from [[secret]]
	 * Used by invitee to verify organization details.
	 */
	@AsBytes()
	signature!: Uint8Array;

	constructor(vals: Partial<OrgInfo>) {
		super();
		Object.assign(this, vals);
	}
}

class InviteeInfo extends Serializable {
	accountId: AccountID = '';
	name = '';
	did = '';

	@AsBytes()
	publicKey!: Uint8Array;
	/**
	 * Signature created using the HMAC key derived from [[Invite.secret]]
	 * Used by organization owner to verify invitee details.
	 */
	@AsBytes()
	signature!: Uint8Array;
	/**
	 * Signature of organization details created using the invitee accounts
	 * own secret signing key. Will be stored on the [[Member]] object to
	 * allow the member to verify the organization details at a later time.
	 */
	@AsBytes()
	orgSignature!: Uint8Array;

	constructor(vals: Partial<InviteeInfo>) {
		super();
		Object.assign(this, vals);
	}
}

/**
 * The `Invite` class encapsules most of the logic and information necessary to
 * perform a key exchange between an [[Org]] and [[Account]] before adding the
 * [[Account]] as a member. A secret HMAC key is used to sign and verify the public keys
 * of both invitee and organization. This key is derived from a [[secret]], which
 * needs to be communicated between the organization owner and invitee directly.
 *
 * The invite flow generally works as follows:
 *
 * ```ts
 * // ORG OWNER
 *
 * const invite = new Invite("bob@example.com", "add_member");
 *
 * // Generates random secret and signs organization details
 * await invite.intialize(org, orgOwnerAccount);
 *
 * console.log("invite secret: ", invite.secret);
 *
 * // => Invite object is send to server, which sends an email to the invitee
 *
 * // INVITEE
 * // => Invitee fetches `invite` object from server, asks org owner for `secret` (in person)
 *
 * // Verifies organization info and signs own public key
 *
 * const success = await invite.accept(inviteeAccount, secret);
 *
 * if (!success) {
 *     throw "Verification failed! Incorrect secret?";
 * }
 *
 * // => Sends updated invite object to server
 *
 * // ORG OWNER
 *
 * // => Fetches updated invite object
 *
 * // Verify invitee details.
 * if (!(await invite.verifyInvitee())) {
 *     throw "Failed to verify invitee details!";
 * }
 *
 * // DONE!
 * await org.addOrUpdateMember(invite.invitee);
 * ```
 */
export class Invite extends SimpleContainer {
	/** Unique identfier */
	id: InviteID = '';

	/** Time of creation */
	@AsDate()
	created = new Date();

	/**
	 * Expiration time used to limit invite procedure to a certain time
	 * window. This property is also stored in [[encryptedData]] along
	 * with the invite secret to prevent tempering.
	 */
	@AsDate()
	expires = new Date();

	/**
	 * Organization info, including HMAC signature used for verification.
	 * Set during initialization
	 */
	@AsSerializable(OrgInfo)
	org!: OrgInfo;

	/**
	 * Invitee info, including HMAC signature used for verification
	 * Set when the invitee successfully accepts the invite
	 */
	@AsSerializable(InviteeInfo)
	invitee!: InviteeInfo;

	/** Info about who created the invite. */
	invitedBy?: {
		accountId: AccountID;
		name: string;
		did: string;
	} = undefined;

	/**
	 * Random secret used for deriving the HMAC key that is used to sign and
	 * verify organization and invitee details. It is encrypted at rest with an
	 * AES key only available to organization admins. The invitee does not have
	 * access to this property directly but needs to request it from the
	 * organization owner directly.
	 *
	 * @secret
	 * **IMPORTANT**: This property is considered **secret**
	 * and should never stored or transmitted in plain text
	 */
	set secret(s: string | undefined) {
		this._secret = s;
		this._signingKey = null;
	}
	get secret() {
		return this._secret;
	}

	/** Whether this invite has expired */
	get expired(): boolean {
		return new Date() > new Date(this.expires);
	}

	/** Whether this invite has been accepted by the invitee */
	get accepted(): boolean {
		return !!this.invitee;
	}

	private _secret?: string = '';
	private _signingKey?: HMACKey | null = null;

	/** Key derivation paramaters used for deriving the HMAC signing key from [[secret]]. */
	@AsSerializable(PBKDF2Params)
	signingKeyParams = new PBKDF2Params({
		iterations: 1e6
	});

	/**
	 * Parameters used for signing organization and initee details.
	 */
	@AsSerializable(HMACParams)
	signingParams = new HMACParams();

	constructor(
		/** invitee email */
		public did = '',
		/** purpose of the invite */
		public purpose: InvitePurpose = 'join_org'
	) {
		super();
	}

	/**
	 * Initializes the invite by generating a random [[secret]] and [[id]] and
	 * signing and storing the organization details.
	 *
	 * @param org The organization this invite is for
	 * @param invitor Account creating the invite
	 * @param duration Number of hours until this invite expires
	 */
	async initialize(org: Org, invitor: Account, duration = 12) {
		this.id = await uuid();
		this.invitedBy = {
			accountId: invitor.id,
			did: invitor.did,
			name: invitor.name
		};

		// Generate secret
		this.secret = bytesToHex(await getProvider().randomBytes(4));
		// Set expiration time (12 hours from now)
		this.expires = new Date(Date.now() + 1000 * 60 * 60 * duration);

		// Encrypt secret and expiration date (the expiration time is also stored/transmitted
		// in plain text, encrypting it will allow verifying it wasn't tempered with later)
		this._key = org.invitesKey;
		await this.setData(
			stringToBytes(
				marshal({ secret: this.secret, expires: this.expires })
			)
		);

		// Initialize signing params
		this.signingKeyParams.salt = await getProvider().randomBytes(16);

		// Create org signature using key derived from secret (see `_getSigningKey`)
		this.org = new OrgInfo({
			id: org.id,
			name: org.name,
			publicKey: org.publicKey,
			signature: await this._sign(
				concatBytes([stringToBytes(org.id), org.publicKey!], 0x00)
			)
		});
	}

	/**
	 * "Unlocks" the invite with the dedicated key (owned by the respective [[Org]]).
	 * This grants access to the [[secret]] property and verfies that [[expires]] has
	 * not been tempered with.
	 */
	async unlock(key: Uint8Array) {
		await super.unlock(key);
		const { secret, expires } = unmarshal(
			bytesToString(await this.getData())
		);
		this.secret = secret;

		// Verify that expiration time has not been tempered with
		if (this.expires.getTime() !== new Date(expires).getTime()) {
			throw new Err(ErrorCode.VERIFICATION_ERROR);
		}
	}

	lock() {
		super.lock();
		delete this._secret;
		delete this._signingKey;
	}

	/**
	 * Accepts the invite by verifying the organization details and, if successful,
	 * signing and storing the invitees own information. Throws if verification
	 * is unsuccessful.
	 */
	async accept(account: Account, secret: string): Promise<boolean> {
		this.secret = secret;

		// Verify org signature
		if (!(await this.verifyOrg())) {
			this.secret = '';
			return false;
		}

		this.invitee = new InviteeInfo({
			accountId: account.id,
			name: account.name,
			did: account.did,
			publicKey: account.publicKey,
			// this is used by the organization owner to verify the invitees public key
			signature: await this._sign(
				concatBytes(
					[
						stringToBytes(account.id),
						stringToBytes(account.did),
						account.publicKey
					],
					0x00
				)
			),
			// this is used by member later to verify the organization public key
			orgSignature: await account.signOrg(this.org)
		});

		return true;
	}

	/** Verifies the organization information. */
	async verifyOrg(): Promise<boolean> {
		if (!this.org) {
			throw 'Invite needs to be initialized first!';
		}

		return (
			this.expires > new Date() &&
			this._verify(
				this.org.signature,
				concatBytes(
					[stringToBytes(this.org.id), this.org.publicKey],
					0x00
				)
			)
		);
	}

	/** Verifies the invitee information. */
	async verifyInvitee(): Promise<boolean> {
		if (!this.invitee) {
			throw 'Invite needs to be accepted first!';
		}

		return (
			this.expires > new Date() &&
			this._verify(
				this.invitee.signature,
				concatBytes(
					[
						stringToBytes(this.invitee.accountId),
						stringToBytes(this.invitee.did),
						this.invitee.publicKey
					],
					0x00
				)
			)
		);
	}

	private async _getSigningKey() {
		if (!this._signingKey) {
			if (!this.secret) {
				throw 'Secret not available! Was the invite unlocked first?';
			}
			this._signingKey = (await getProvider().deriveKey(
				stringToBytes(this.secret),
				this.signingKeyParams
			)) as HMACKey;
		}
		return this._signingKey;
	}

	private async _sign(val: Uint8Array): Promise<Uint8Array> {
		return getProvider().sign(
			await this._getSigningKey(),
			val,
			this.signingParams
		);
	}

	private async _verify(sig: Uint8Array, val: Uint8Array): Promise<boolean> {
		return await getProvider().verify(
			await this._getSigningKey(),
			sig,
			val,
			this.signingParams
		);
	}
}
