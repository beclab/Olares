import { Request, Response, RequestAuthentication } from './transport';
import {
	marshal,
	Serializable,
	stringToBytes,
	AsDate,
	AsSerializable,
	AsBytes
} from './encoding';
import { HMACParams } from './crypto';
import { Storable } from './storage';
import { DeviceInfo } from './platform';
import { AccountID } from './account';
import { getCryptoProvider as getProvider } from './platform';

/** Unique identifier for [[Session]]s */
export type SessionID = string;

/** Public session info (used for display purposes) */
export class SessionInfo extends Serializable {
	id = '';
	account: AccountID = '';

	@AsDate()
	created: Date = new Date(0);

	@AsDate()
	updated: Date = new Date(0);

	@AsDate()
	lastUsed: Date = new Date(0);

	lastLocation?: {
		city?: string;
		country?: string;
	} = undefined;

	@AsDate()
	expires?: Date;

	@AsSerializable(DeviceInfo)
	device?: DeviceInfo;

	asAdmin = false;
}

/**
 * A session represents a trusted connection between a [[Server]] and [[Client]]
 * which can be used to authenticate requests, allowing both parties to verify
 * the other parties identity and check the request/response bodies integrity.
 * The authentication flow usually works as follows:
 *
 * ```ts
 * // CLIENT
 *
 * const request = createRequest();
 * await this.session.authenticate(request);
 *
 * // SERVER
 *
 * if (!(await context.session.verify(request))) {
 *     throw "Failed to verify request!";
 * }
 *
 * const response = processRequest(request);
 * await context.session.authenticate(response);
 *
 * // CLIENT
 *
 * if (!(await context.session.verify(response))) {
 *     throw "Failed to verify response!";
 * }
 *
 * processResponse(response);
 * ```
 *
 * ```
 *                        ┌──────────┐     ┌──────────┐
 *                        │Client (C)│     │Server (S)│
 *                        └─────┬────┘     └────┬─────┘
 * ┌──────────────────────────┐ │               │
 * │req = [request body]      │ │   req, sid,   │
 * │t1 = [timestamp]          │ │   t1, sig1    │ ┌──────────────────────────┐
 * │sig1 = HMAC(K, sid|t1|req)│ │──────────────▶│ │=> verify sig1            │
 * └──────────────────────────┘ │               │ │res = [response body]     │
 *                              │               │ │t2 = [timestamp]          │
 *             ┌──────────────┐ │ res, t2, sig2 │ │sig2 = HMAC(K, sid|t2|res)│
 *             │=> verify sig2│ │◁ ─ ─ ─ ─ ─ ─ ─│ └──────────────────────────┘
 *             └──────────────┘ │               │
 *                              │               │
 *                              ▼               ▼
 * ```
 */
export class Session extends Serializable implements Storable {
	get kind(): string {
		return 'session';
	}

	/** Unique identifier */
	id = '';

	/** Associated [[Account]] */
	account: AccountID = '';

	/** Time of creation */
	@AsDate()
	created = new Date(0);

	/** Time of last update */
	@AsDate()
	updated = new Date(0);

	/** When this session was last used to authenticate a request */
	@AsDate()
	lastUsed = new Date(0);

	/** Expiration time */
	@AsDate()
	expires?: Date;

	/** Session key used to sign/verify requests and responses */
	@AsBytes()
	key?: Uint8Array;

	/** Info about the device the client is running on */
	@AsSerializable(DeviceInfo)
	device?: DeviceInfo;

	asAdmin = false;

	lastLocation?: {
		city?: string;
		country?: string;
	} = undefined;

	/**
	 * Public session info
	 */
	get info(): SessionInfo {
		return new SessionInfo().fromRaw({
			id: this.id,
			account: this.account,
			created: this.created,
			updated: this.updated,
			lastUsed: this.lastUsed,
			expires: this.expires,
			device: this.device && this.device.toRaw(),
			lastLocation: this.lastLocation,
			asAdmin: this.asAdmin
		});
	}

	/**
	 * Authenticates a [[Request]] or [[Response]] by signing the session id,
	 * timestamp and request/response body using the session [[key]].
	 */
	async authenticate(r: Request | Response): Promise<void> {
		const data = (<Request>r).params || (<Response>r).result;
		r.auth = await this._sign(data);
	}

	/**
	 * Verifies session id, timestamp and request/response body of a given
	 * [[Request]] or [[Response]] using the session [[key]].
	 */
	async verify(r: Request | Response): Promise<boolean> {
		if (!r.auth) {
			return false;
		}

		const data = (<Request>r).params || (<Response>r).result;

		return this._verify(r.auth, data);
	}

	private async _sign(data: any): Promise<RequestAuthentication> {
		const time = new Date();
		const session = this.id;
		const message = `${session}_${time.toISOString()}_${marshal(data)}`;
		const signature = await getProvider().sign(
			this.key!,
			stringToBytes(message),
			new HMACParams()
		);
		return new RequestAuthentication({
			session,
			time,
			signature
		});
	}

	private async _verify(
		auth: RequestAuthentication,
		data: any
	): Promise<boolean> {
		const { signature, time } = auth;
		const message = `${this.id}_${time.toISOString()}_${marshal(data)}`;
		return await getProvider().verify(
			this.key!,
			signature,
			stringToBytes(message),
			new HMACParams()
		);
	}
}
