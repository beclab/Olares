export interface DropboxAuthPluginInterface {
	initialize(): Promise<void>;
	signIn(): Promise<DropboxAuthResult>;
}

export interface DropboxAuthResult {
	accessToken: string;

	/// The associated user id.
	uid: string;

	/// The refresh token if accessToken is short-lived.
	refreshToken?: string;

	/// The expiration time of the (short-lived) accessToken.
	tokenExpirationTimestamp?: number;
}
