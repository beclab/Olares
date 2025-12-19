/**
 * Contains a public-private key pair and the associated key ID.
 */
export type KeyMaterial = {
	keyId: string;
	keyPair: { publicJwk: PublicJwk; privateJwk: PrivateJwk };
};

export type Jwk = {
	/** The "alg" (algorithm) parameter identifies the algorithm intended for use with the key. */
	alg?: string;
	/** The "alg" (algorithm) parameter identifies the algorithm intended for use with the key. */
	kid?: string;
	/** identifies the cryptographic algorithm family used with the key, such "EC". */
	kty: string;
};

export type PublicJwk = Jwk & {
	/** The "crv" (curve) parameter identifies the cryptographic curve used with the key.
	 * MUST be present for all EC public keys
	 */
	crv: string;
	/**
	 * the x coordinate for the Elliptic Curve point.
	 * Represented as the base64url encoding of the octet string representation of the coordinate.
	 * MUST be present for all EC public keys
	 */
	x: string;
	/**
	 * the y coordinate for the Elliptic Curve point.
	 * Represented as the base64url encoding of the octet string representation of the coordinate.
	 */
	y?: string;
};

export type PrivateJwk = PublicJwk & {
	/**
	 * the Elliptic Curve private key value.
	 * It is represented as the base64url encoding of the octet string representation of the private key value
	 * MUST be present to represent Elliptic Curve private keys.
	 */
	d: string;
};

export interface Signer {
	/**
	 * signs the provided payload using the provided JWK
	 * @param content - the content to sign
	 * @param privateJwk - the key to sign with
	 * @returns the signed content (aka signature)
	 */
	sign(content: Uint8Array, privateJwk: PrivateJwk): Promise<Uint8Array>;
	/**
	 * Verifies a signature against the provided payload hash and public key.
	 * @param content - the content to verify with
	 * @param signature - the signature to verify against
	 * @param publicJwk - the key to verify with
	 * @returns a boolean indicating whether the signature matches
	 */
	verify(
		content: Uint8Array,
		signature: Uint8Array,
		publicJwk: PublicJwk
	): Promise<boolean>;

	/**
	 * generates a random key pair
	 * @returns the public and private keys as JWKs
	 */
	generateKeyPair(): Promise<{
		publicJwk: PublicJwk;
		privateJwk: PrivateJwk;
	}>;

	/**
	 * converts public key in bytes into a JWK
	 * @param publicKeyBytes - the public key to convert into JWK
	 * @returns the public key in JWK format
	 */
	publicKeyToJwk(publicKeyBytes: Uint8Array): Promise<PublicJwk>;
}
