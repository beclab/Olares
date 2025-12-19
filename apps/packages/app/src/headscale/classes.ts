export class ACL {
	public groups: { [key: string]: [string] } = {};

	public constructor(init?: Partial<Route>) {
		Object.assign(this, init);
	}
}

export class Route {
	// current (hs 18+) method of handling a route
	advertised = true;
	prefix = '';
	enabled = false;
	id = 0;

	public constructor(init?: Partial<Route>) {
		Object.assign(this, init);
	}
}

export class APIKey {
	id = '';
	prefix = '';
	expiration = '';
	createdAt = '';
	lastSeen = '';

	public constructor(init?: Partial<Route>) {
		Object.assign(this, init);
	}
}

export class PreAuthKey {
	public user = '';
	public id = '';
	public key = '';
	public createdAt = '';
	public expiration = '';
	public reusable = false;
	public ephemeral = false;
	public used = false;

	public constructor(init?: Partial<PreAuthKey>) {
		Object.assign(this, init);
	}
}

export class User {
	public id = '';
	public name = '';
	public createdAt = '';
	public constructor(init?: Partial<User>) {
		Object.assign(this, init);
	}
}
