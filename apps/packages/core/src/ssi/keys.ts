export type JWT = string;

export enum KeyType {
	Ed25519 = 'Ed25519',
	X25519 = 'X25519',
	SECP256k1 = 'secp256k1',
	P224 = 'P-224',
	P256 = 'P-256',
	P384 = 'P-384',
	P521 = 'P-521',
	RSA = 'RSA'
}

export const W3CSecurityContext = 'https://w3id.org/security/v1';
export const JWS2020LinkedDataContext =
	'https://w3id.org/security/suites/jws-2020/v1';

export const RSAKeySize = 2048;

export enum HashType {
	SHA256 = 'SHA256'
}

export enum ProofPurpose {
	AssertionMethod = 'assertionMethod',
	Authentication = 'authentication'
}

export enum PayloadFormat {
	JWTFormat = 'jwt',
	LDPFormat = 'ldp'
}

export interface StoreKeyRequest {
	id: string;
	type: KeyType;
	controller: string;
	base58PrivateKey: string;
}
