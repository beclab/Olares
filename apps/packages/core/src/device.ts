export class TermiPassDeviceInfo {
	//termiPassID: string | undefined = '';
	/** Platform/Operating System running on the device */
	platform = '';
	/** OS version running on the device */
	osVersion = '';
	/** Unique device identifier */
	id = '';
	appVersion = '';
	vendorVersion = '';
	/** The user agent of the browser running the application */
	userAgent = '';
	/** The devices locale setting */
	locale = 'en';
	/** The device manufacturer, if available */
	manufacturer = '';
	/** The device mode, if available */
	model = '';
	/** The browser the application was loaded in, if applicable */
	browser = '';
	/** The version of the browser the application was loaded in, if applicable */
	browserVersion = '';
	description = 'Unknown Device';
	runtime = '';
	tailScaled = false;
	tailScale_id = '';
	sso = '';
	srp_id = '';
	createTime = 0;
	lastSeenTime = 0;
	lastIp = '';
	lastIpLocation = '';
	client_type: 'larePass' | 'desktop' | '' = 'desktop';
	firebase_token = '';

	constructor(props?: Partial<TermiPassDeviceInfo>) {
		props && Object.assign(this, props);
	}
}

export class VaultActiveSessionInfo {
	platform = '';
	osVersion = '';
	id = '';
	appVersion = '';
	vendorVersion = '';
	userAgent = '';
	locale = 'en';
	manufacturer = '';
	model = '';
	browser = '';
	browserVersion = '';
	description = 'Unknown Device';
	runtime = '';
	kind = '';
	version = '';
	constructor(props?: Partial<VaultActiveSessionInfo>) {
		props && Object.assign(this, props);
	}
}

export interface SSOToken {
	expireTime: number;
	createTime: number;
	tokenType: string;
	username: string;
	uninitialized: string;
	authLevel: string;
	firstFactorTimestamp: number;
	secondFactorTimestamp: number;
}

export class HeadScaleDevice {
	public id = '';
	public name = '';
	public givenName = '';
	public lastSeen = '';
	public ipAddresses: string[] = [];
	public forcedTags: string[] = [];
	public validTags: string[] = [];
	public invalidTags: string[] = [];
	public expiry = '';
	public user: { id: string; name: string; createdAt: string } = {
		name: '',
		id: '',
		createdAt: ''
	};
	public createdAt = '';
	public registerMethod = '';
	public online = false;

	public constructor(init?: Partial<HeadScaleDevice>) {
		Object.assign(this, init);
	}
}
