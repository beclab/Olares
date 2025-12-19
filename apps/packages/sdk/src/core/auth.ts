import {
	Serializable,
	stringToBytes,
	AsBytes,
	AsSerializable,
	AsDate,
	bytesToBase64
} from './encoding';
import { PBKDF2Params } from './crypto';
import { getCryptoProvider as getProvider } from './platform';
import { DeviceInfo } from './platform';
import { Storable } from './storage';
import { AccountID } from './account';
import { KeyStoreEntryInfo } from './key-store';
import { SessionInfo } from './session';
import { SRPSession } from './srp';
import { getIdFromDID, uuid } from './util';

export enum AuthPurpose {
	Signup = 'signup',
	Login = 'login',
	Recover = 'recover',
	//GetLegacyData = 'get_legacy_data',
	AccessKeyStore = 'access_key_store',
	TestAuthenticator = 'test_authenticator',
	AdminLogin = 'admin_login'
}

export enum AuthType {
	DID = 'did',
	WebAuthnPlatform = 'webauthn_platform',
	WebAuthnPortable = 'webauthn_portable',
	Totp = 'totp',
	PublicKey = 'public_key',
	OpenID = 'openid',
	OsToken = 'os_token',
	OsPassword = 'os_password',
	SSI = 'ssi'
}

export enum AuthenticatorStatus {
	Registering = 'registering',
	Active = 'active',
	Revoked = 'revoked'
}

export class AuthenticatorInfo extends Serializable {
	/** Time of creation */
	@AsDate()
	created = new Date();

	@AsDate()
	lastUsed?: Date;

	id = '';

	description = '';

	type: AuthType = AuthType.SSI;

	purposes: AuthPurpose[] = [];

	status: AuthenticatorStatus = AuthenticatorStatus.Registering;

	@AsSerializable(DeviceInfo)
	device?: DeviceInfo;

	constructor(init: Partial<AuthenticatorInfo> = {}) {
		super();
		Object.assign(this, init);
	}
}

export class Authenticator<T = any> extends Serializable {
	/** Time of creation */
	@AsDate()
	created = new Date();

	@AsDate()
	lastUsed?: Date;

	id = '';

	description = '';

	type: AuthType = AuthType.SSI;

	purposes: AuthPurpose[] = [];

	status: AuthenticatorStatus = AuthenticatorStatus.Registering;

	@AsSerializable(DeviceInfo)
	device?: DeviceInfo;

	state?: T = undefined;

	get info() {
		return new AuthenticatorInfo(this);
	}

	constructor(init: Partial<Authenticator> = {}) {
		super();
		Object.assign(this, init);
	}

	async init() {
		this.id = await uuid();
		this.created = new Date();
	}
}

export enum AuthRequestStatus {
	Started = 'started',
	Verified = 'verified',
	Canceled = 'canceled'
}

export class AuthRequest<T = any> extends Serializable {
	id = '';

	/** Time of creation */
	@AsDate()
	created!: Date;

	@AsDate()
	verified!: Date;

	type: AuthType = AuthType.SSI;

	@AsSerializable(DeviceInfo)
	device?: DeviceInfo = undefined;

	authenticatorId = '';

	purpose: AuthPurpose = AuthPurpose.Login;

	token = '';

	state?: T = undefined;

	tries = 0;

	status: AuthRequestStatus = AuthRequestStatus.Started;

	constructor(init: Partial<AuthRequest> = {}) {
		super();
		Object.assign(this, init);
	}

	async init() {
		this.id = await uuid();
		this.created = new Date();
		// Create random 16-byte verification token
		this.token = bytesToBase64(await getProvider().randomBytes(16));
		this.tries = 0;
	}
}

export interface AuthServer {
	supportsType(type: AuthType): boolean;

	initAuthenticator(
		authenticator: Authenticator,
		auth: Auth,
		params?: any
	): Promise<any>;

	activateAuthenticator(
		authenticator: Authenticator,
		params?: any
	): Promise<void>;

	initAuthRequest(
		authenticator: Authenticator,
		request: AuthRequest,
		params?: any
	): Promise<any>;

	verifyAuthRequest(
		authenticator: Authenticator,
		request: AuthRequest,
		params?: any
	): Promise<void>;
}

export interface AuthClient {
	supportsType(type: AuthType): boolean;
	prepareRegistration(serverData: any): Promise<any>;
	prepareAuthentication(serverData: any): Promise<any>;
}

export enum AccountStatus {
	Unregistered = 'unregistered',
	Active = 'active',
	Blocked = 'blocked',
	Deleted = 'deleted'
}
/**
 * Contains authentication data needed for SRP session negotiation
 */
export class Auth extends Serializable implements Storable {
	get kind(): string {
		return 'auth';
	}

	id = '';

	@AsDate()
	created: Date = new Date();

	/** Id of the [[Account]] the authentication data belongs to */
	account?: AccountID = undefined;

	get accountId() {
		return this.account;
	}

	accountStatus: AccountStatus = AccountStatus.Unregistered;

	/** Verifier used for SRP session negotiation */
	@AsBytes()
	verifier?: Uint8Array;

	/**
	 * Key derivation params used by the client to compute session key from the
	 * users master password
	 * */
	@AsSerializable(PBKDF2Params)
	keyParams = new PBKDF2Params();

	@AsSerializable(DeviceInfo)
	trustedDevices: DeviceInfo[] = [];

	@AsSerializable(Authenticator)
	authenticators: Authenticator[] = [];

	@AsSerializable(AuthRequest)
	authRequests: AuthRequest[] = [];

	@AsSerializable(KeyStoreEntryInfo)
	keyStoreEntries: KeyStoreEntryInfo[] = [];

	@AsSerializable(SessionInfo)
	sessions: SessionInfo[] = [];

	@AsSerializable(SRPSession)
	srpSessions: SRPSession[] = [];

	mfaOrder: string[] = [];

	/** Invites to organizations */
	invites: {
		id: string;
		orgId: string;
		orgName: string;
		expires: string;
	}[] = [];

	// @AsSerializable(PBES2Container)
	// legacyData?: PBES2Container;

	/** Completely disables mfa for a given account. Only use for testing! */
	disableMFA = false;

	metaData?: any = undefined;

	constructor(public did: string = '') {
		super();
	}

	async init() {
		this.id = await getIdFromDID(this.did);
	}

	/**
	 * Generate the session key from the users master `password`
	 */
	async getAuthKey(password: string) {
		// If no salt is set yet (i.e. during initialization),
		// generate a random value
		if (!this.keyParams.salt.length) {
			this.keyParams.salt = await getProvider().randomBytes(16);
		}
		return getProvider().deriveKey(stringToBytes(password), this.keyParams);
	}
}
