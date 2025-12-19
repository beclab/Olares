import type { PrivateJwk } from '../../types';
/**
 * General JWS definition. Payload is returned as an empty
 * string when JWS Unencoded Payload Option
 * [RFC7797](https://www.rfc-editor.org/rfc/rfc7797) is used.
 */
export type GeneralJws = {
	payload: string;
	signatures: SignatureEntry[];
};

/**
 * An entry of the `signatures` array in a general JWS.
 */
export type SignatureEntry = {
	/**
	 * The "protected" member MUST be present and contain the value
	 * BASE64URL(UTF8(JWS Protected Header)) when the JWS Protected
	 * Header value is non-empty; otherwise, it MUST be absent.  These
	 * Header Parameter values are integrity protected.
	 */
	protected: string;

	/**
	 * The "signature" member MUST be present and contain the value
	 * BASE64URL(JWS Signature).
	 */
	signature: string;
};

export type JwsHeaderParameters = {
	/** JWS "alg" (Algorithm) Header Parameter. */
	alg: string;
	/** JWS "kid" (Key ID) Parameter. */
	kid: string;

	iss?: string;
};

export type SignatureInput = {
	protectedHeader: JwsHeaderParameters;
	privateJwk: PrivateJwk;
};
