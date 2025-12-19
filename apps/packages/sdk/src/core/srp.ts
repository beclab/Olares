/**
 * This module implements version 6a of the [Secure Remote
 * Password](http://srp.stanford.edu/design.html) protocol.
 *
 * The [[Client]] and [[Server]] classes are the high-level interfaces to be
 * used on the client and server side, respectively. The SRP key negotiation
 * flow usually happens as follows:
 *
 * #### First signup
 *
 * ```ts
 * // CLIENT
 *
 * const u = [username/email];
 * const p = [user password];
 * const i = [iteration count];
 * const s = [random salt];
 *
 * const x =  PBKDF2(p, s, i);
 *
 * const client = new Client();
 * await client.initialize(x);
 * const v = client.v;
 *
 * => Send u, v, s, i to server
 * ```
 *
 * #### Login / create session
 *
 * ```ts
 * // CLIENT => request login for `u`
 *
 * // SERVER
 *
 * const { v, s, i } = getAuthInfo(u);
 *
 * const server = new Server();
 * await server.initialize(v);
 * const B = server.B;
 *
 * // => Send s, i, B to client
 *
 * // CLIENT
 *
 * const x =  PBKDF2(p, s, i);
 *
 * const client = new Client();
 * await client.initialize(x);
 * await client.setB(B);
 *
 * const A = client.A;
 * const M1 = client.M1;
 *
 * // Common session key
 * const K = client.K;
 *
 * // => Send A, M1 to server
 *
 * // SERVER
 *
 * await server.setA(A);
 * if (server.M1 !== M1) {
 *     throw "Invalid credentials!";
 * }
 *
 * // Common session key
 * const K = server.K;
 *
 * // => Success!
 *
 * // [OPTIONAL] (This step is usually not required since even without
 * // verifying M2, the client will notice something is wrong as soon as they try
 * // to verify an authenticated response from the server.)
 *
 * M2 = server.M2;
 * // => Send M2 to client
 *
 * // CLIENT
 * if (client.M2 !== M2) {
 *     throw "Something is fishy!";
 * }
 * ```
 *
 * #### Overview:
 *
 * ```
 *                     ┌──────────┐     ┌──────────┐
 *                     │Client (C)│     │Server (S)│
 *                     └─────┬────┘     └────┬─────┘
 *     ┌───────────────────┐ │               │
 *     │u = [email address]│ │     u, A      │
 *     │a, A = [random*]   │ │──────────────▶│
 *     └───────────────────┘ │               │ ┌────────────────┐
 *                           │               │ │b, B = [random*]│
 * ┌───────────────────────┐ │    s, i, B    │ └────────────────┘
 * │p = [master password]  │ │◁ ─ ─ ─ ─ ─ ─ ─│
 * │x = PBKDF2(p,s,i)      │ │               │
 * │K = K_client(x, a, B)* │ │               │
 * │M = M(A, B, K)*        │ │       M       │ ┌───────────────────────┐
 * └───────────────────────┘ │──────────────▶│ │K' = K_server(v, b, A)*│
 *                           │               │ │M' = M(A, B, K')       │
 *                           │               │ │=> verify M == M'      │
 *         ┌───────────────┐ │      sid      │ │S = [session id]       │
 *         │=> store sid, K│ │◀ ─ ─ ─ ─ ─ ─ ─│ │=> store sid, K        │
 *         └───────────────┘ │               │ └───────────────────────┘
 *                           │               │
 *                           ▼               ▼
 * ```
 */
import { BigInteger } from '../vendor/jsbn';
import {
	bytesToHex,
	hexToBytes,
	concatBytes,
	Serializable,
	AsBigInteger,
	AsDate
} from './encoding';
import { HashParams } from './crypto';
import { getCryptoProvider as getProvider } from './platform';
import { uuid } from './util';

async function digest(
	hash: 'SHA-1' | 'SHA-256' = 'SHA-256',
	...input: Uint8Array[]
): Promise<Uint8Array> {
	return getProvider().hash(
		concatBytes(input),
		new HashParams({ algorithm: hash })
	);
}

function i2b(i: BigInteger): Uint8Array {
	let hex = i.toString(16);
	if (hex.length % 2) {
		hex = '0' + hex;
	}
	return hexToBytes(hex);
}

function b2i(b: Uint8Array): BigInteger {
	return new BigInteger(bytesToHex(b), 16);
}

export class SRPSession extends Serializable {
	id = '';

	@AsDate()
	created = new Date();

	async init() {
		this.id = await uuid();
	}

	failedAttempts = 0;

	asAdmin = false;

	@AsBigInteger()
	x?: BigInteger;
	@AsBigInteger()
	v?: BigInteger;
	@AsBigInteger()
	a?: BigInteger;
	@AsBigInteger()
	A?: BigInteger;
	@AsBigInteger()
	b?: BigInteger;
	@AsBigInteger()
	B?: BigInteger;
	@AsBigInteger()
	K?: BigInteger;
	@AsBigInteger()
	M1?: BigInteger;
	@AsBigInteger()
	M2?: BigInteger;
}

/**
 * High-level interface for the client side
 */
export class Client {
	/** Verifier value, available after calling [[initialize]] */
	get v() {
		return this._session.v ? i2b(this._session.v) : null;
	}

	/** Client-side random initializer, available after calling [[initialize]] */
	get A() {
		return this._session.A ? i2b(this._session.A) : null;
	}

	/** Common session key, available after calling [[initialize]] and [[setB]] */
	get K() {
		return this._session.K ? i2b(this._session.K) : null;
	}

	/**
	 * First value used to verify a successful key exchange, available after
	 * calling [[initialize]] and [[setB]]
	 */
	get M1() {
		return this._session.M1 ? i2b(this._session.M1) : null;
	}

	/**
	 * Second value used to verify a successful key exchange, available after
	 * calling [[initialize]] and [[setB]]
	 */
	get M2() {
		return this._session.M2 ? i2b(this._session.M2) : null;
	}

	private _srp: Core;
	private _session: SRPSession;

	constructor(state = new SRPSession(), length: SRPGroupLength = 4096) {
		this._srp = new Core(length);
		this._session = state;
	}

	/** Initialize client using the given secret `x`, generating [[v]] and [[A]] */
	async initialize(x: Uint8Array) {
		this._session.x = b2i(x);
		this._session.v = this._srp.v(this._session.x);
		this._session.a = b2i(await getProvider().randomBytes(32));
		this._session.A = this._srp.A(this._session.a);
	}

	/**
	 * Apply [[Server.B]] value, calculating [[K]], [[M1]] and [[M2]].
	 * Should only be called after [[initialize]] has been called.
	 */
	async setB(B: Uint8Array) {
		if (!this._session.x || !this._session.a || !this._session.A) {
			throw 'not initialized';
		}
		this._session.B = b2i(B);
		this._session.K = await this._getKey();
		this._session.M1 = await this._srp.M1(
			this._session.A,
			this._session.B,
			this._session.K
		);
		this._session.M2 = await this._srp.M2(
			this._session.A,
			this._session.M1,
			this._session.K
		);
	}

	private async _getKey(): Promise<BigInteger> {
		if (
			!this._session.x ||
			!this._session.a ||
			!this._session.A ||
			!this._session.B
		) {
			throw 'not initialized';
		}

		if (this._srp.isZeroWhenModN(this._session.B)) {
			throw 'Invalid B value';
		}

		const u = await this._srp.u(this._session.A, this._session.B);
		const S = await this._srp.clientS(
			this._session.B,
			this._session.x,
			this._session.a,
			u
		);
		const K = await this._srp.K(S);
		return K;
	}
}

export class Server {
	/** Server-side random initializer, available after calling [[initialize]] */
	get B() {
		return this._session.B ? i2b(this._session.B) : null;
	}

	/** Common session key, available after calling [[initialize]] and [[setA]] */
	get K() {
		return this._session.K ? i2b(this._session.K) : null;
	}

	/**
	 * First value used to verify a successful key exchange, available after
	 * calling [[initialize]] and [[setA]]
	 */
	get M1() {
		return this._session.M1 ? i2b(this._session.M1) : null;
	}

	/**
	 * Second value used to verify a successful key exchange, available after
	 * calling [[initialize]] and [[setA]]
	 */
	get M2() {
		return this._session.M2 ? i2b(this._session.M2) : null;
	}

	private _srp: Core;
	private _session: SRPSession;

	constructor(state: SRPSession, length: SRPGroupLength = 4096) {
		this._srp = new Core(length);
		this._session = state;
	}

	/** Initialize server using the given verfifier `v`. Generates [[B]]. */
	async initialize(v: Uint8Array) {
		this._session.v = b2i(v);
		this._session.b = b2i(await getProvider().randomBytes(32));
		this._session.B = await this._srp.B(this._session.v, this._session.b);
	}

	/**
	 * Apply [[Client.A]] value, calculating [[K]], [[M1]] and [[M2]].
	 * Should only be called after [[initialize]] has been called.
	 */
	async setA(A: Uint8Array) {
		if (!this._session.v || !this._session.b || !this._session.B) {
			throw 'not initialized';
		}
		this._session.A = b2i(A);
		this._session.K = await this._getKey();
		this._session.M1 = await this._srp.M1(
			this._session.A,
			this._session.B,
			this._session.K
		);
		this._session.M2 = await this._srp.M2(
			this._session.A,
			this._session.M1,
			this._session.K
		);
	}

	private async _getKey() {
		if (
			!this._session.v ||
			!this._session.b ||
			!this._session.B ||
			!this._session.A
		) {
			throw 'not initialized';
		}

		if (this._srp.isZeroWhenModN(this._session.A)) {
			throw 'Invalid A value';
		}

		const u = await this._srp.u(this._session.A, this._session.B);
		const S = this._srp.serverS(
			this._session.A,
			this._session.v,
			u,
			this._session.b
		);
		const K = await this._srp.K(S);
		return K;
	}
}

/**
 * Implements various formulas defined in the [SRP
 * specification](http://srp.stanford.edu/design.html), used by [[Client]] and
 * [[Server]] classes.
 */
export class Core {
	private _params: SRPParams;

	constructor(length: SRPGroupLength) {
		this._params = getParams(length);
		if (!this._params) {
			throw 'Invalid group length!';
		}
	}

	/** Creates a hash of all arguments, concatenated */
	async H(...inp: BigInteger[]): Promise<BigInteger> {
		const hash = await digest(this._params.hash, ...inp.map((i) => i2b(i)));
		return b2i(hash);
	}

	/**
	 * Calculates verifier `v` from secret `x` according to the formula
	 * ```
	 * v = g ^ x % N
	 * ```
	 */
	v(x: BigInteger): BigInteger {
		return this._params.g.modPow(x, this._params.N);
	}

	/** Calculates `A` from random value `a`, according to the formula
	 * ```
	 * A = g ^ a % N
	 * ```
	 */
	A(a: BigInteger) {
		return this._params.g.modPow(a, this._params.N);
	}

	/**
	 * Calculates `B` from `v` and the random value `b` according to the formula
	 * ```
	 * B = (k * v + g ^ b % N) % N
	 * ```
	 */
	async B(v: BigInteger, b: BigInteger): Promise<BigInteger> {
		const k = await this.k();
		return k
			.multiply(v)
			.add(this._params.g.modPow(b, this._params.N))
			.mod(this._params.N);
	}

	/**
	 * Calculates `u` from `A` and `B` according to the formula
	 * ```
	 * u = H(A | B)
	 * ```
	 */
	async u(A: BigInteger, B: BigInteger): Promise<BigInteger> {
		return await this.H(A, B);
	}

	/**
	 * Calculates `S` from `B`, `x`, `a` and `u` according to the formula
	 * ```
	 * S = (B - k * (g ^ x % N)) ^ (a + u * x) % N
	 * ```
	 */
	async clientS(
		B: BigInteger,
		x: BigInteger,
		a: BigInteger,
		u: BigInteger
	): Promise<BigInteger> {
		const k = await this.k();
		return B.subtract(
			k.multiply(this._params.g.modPow(x, this._params.N))
		).modPow(a.add(u.multiply(x)), this._params.N);
	}

	/**
	 * Calculates `S` from `A`, `v`, `b` and `u` according to the formula
	 * ```
	 * S = (A * v ^ u % N) ^ b % N
	 * ```
	 */
	serverS(A: BigInteger, v: BigInteger, u: BigInteger, b: BigInteger) {
		return A.multiply(v.modPow(u, this._params.N)).modPow(
			b,
			this._params.N
		);
	}

	/**
	 * Calculates the multiplier `k = H(N | g)` according to the SRP-6a specification
	 */
	async k(): Promise<BigInteger> {
		return this.H(this._params.N, this._params.g);
	}

	/**
	 * Calculates the shared key `K = H(S)`
	 */
	async K(S: BigInteger): Promise<BigInteger> {
		return this.H(S);
	}

	/**
	 * Calculates the first verification value `M1 = H(A | B | K)`
	 */
	async M1(A: BigInteger, B: BigInteger, K: BigInteger): Promise<BigInteger> {
		return await this.H(A, B, K);
	}

	/**
	 * Calculates the second verification value `M2 = H(A | M1 | K)`
	 */
	async M2(
		A: BigInteger,
		M1: BigInteger,
		K: BigInteger
	): Promise<BigInteger> {
		return await this.H(A, M1, K);
	}

	/**
	 * This is used to ensure that values are not zero when mod N.
	 */
	isZeroWhenModN(n: BigInteger) {
		return n.mod(this._params.N).equals(BigInteger.ZERO);
	}
}

type SRPGroupLength = 3072 | 4096 | 6144 | 8192;

interface SRPParams {
	length: SRPGroupLength;
	hash: 'SHA-1' | 'SHA-256';
	g: BigInteger;
	N: BigInteger;
}

function h2i(hex: string): BigInteger {
	return new BigInteger(hex.replace(' ', ''), 16);
}

function getParams(length: SRPGroupLength): SRPParams {
	switch (length) {
		case 3072:
			return {
				length: 3072,
				hash: 'SHA-256',
				g: h2i('05'),
				N: h2i(
					'FFFFFFFF FFFFFFFF C90FDAA2 2168C234 C4C6628B 80DC1CD1 29024E08 8A67CC74 020BBEA6 3B139B22 514A0879 8E3404DD EF9519B3 CD3A431B 302B0A6D F25F1437 4FE1356D 6D51C245 E485B576 625E7EC6 F44C42E9 A637ED6B 0BFF5CB6 F406B7ED EE386BFB 5A899FA5 AE9F2411 7C4B1FE6 49286651 ECE45B3D C2007CB8 A163BF05 98DA4836 1C55D39A 69163FA8 FD24CF5F 83655D23 DCA3AD96 1C62F356 208552BB 9ED52907 7096966D 670C354E 4ABC9804 F1746C08 CA18217C 32905E46 2E36CE3B E39E772C 180E8603 9B2783A2 EC07A28F B5C55DF0 6F4C52C9 DE2BCBF6 95581718 3995497C EA956AE5 15D22618 98FA0510 15728E5A 8AAAC42D AD33170D 04507A33 A85521AB DF1CBA64 ECFB8504 58DBEF0A 8AEA7157 5D060C7D B3970F85 A6E1E4C7 ABF5AE8C DB0933D7 1E8C94E0 4A25619D CEE3D226 1AD2EE6B F12FFA06 D98A0864 D8760273 3EC86A64 521F2B18 177B200C BBE11757 7A615D6C 770988C0 BAD946E2 08E24FA0 74E5AB31 43DB5BFC E0FD108E 4B82D120 A93AD2CA FFFFFFFF FFFFFFFF'
				)
			};
		case 4096:
			return {
				length: 4096,
				hash: 'SHA-256',
				g: h2i('05'),
				N: h2i(
					'FFFFFFFF FFFFFFFF C90FDAA2 2168C234 C4C6628B 80DC1CD1 29024E08 8A67CC74 020BBEA6 3B139B22 514A0879 8E3404DD EF9519B3 CD3A431B 302B0A6D F25F1437 4FE1356D 6D51C245 E485B576 625E7EC6 F44C42E9 A637ED6B 0BFF5CB6 F406B7ED EE386BFB 5A899FA5 AE9F2411 7C4B1FE6 49286651 ECE45B3D C2007CB8 A163BF05 98DA4836 1C55D39A 69163FA8 FD24CF5F 83655D23 DCA3AD96 1C62F356 208552BB 9ED52907 7096966D 670C354E 4ABC9804 F1746C08 CA18217C 32905E46 2E36CE3B E39E772C 180E8603 9B2783A2 EC07A28F B5C55DF0 6F4C52C9 DE2BCBF6 95581718 3995497C EA956AE5 15D22618 98FA0510 15728E5A 8AAAC42D AD33170D 04507A33 A85521AB DF1CBA64 ECFB8504 58DBEF0A 8AEA7157 5D060C7D B3970F85 A6E1E4C7 ABF5AE8C DB0933D7 1E8C94E0 4A25619D CEE3D226 1AD2EE6B F12FFA06 D98A0864 D8760273 3EC86A64 521F2B18 177B200C BBE11757 7A615D6C 770988C0 BAD946E2 08E24FA0 74E5AB31 43DB5BFC E0FD108E 4B82D120 A9210801 1A723C12 A787E6D7 88719A10 BDBA5B26 99C32718 6AF4E23C 1A946834 B6150BDA 2583E9CA 2AD44CE8 DBBBC2DB 04DE8EF9 2E8EFC14 1FBECAA6 287C5947 4E6BC05D 99B2964F A090C3A2 233BA186 515BE7ED 1F612970 CEE2D7AF B81BDD76 2170481C D0069127 D5B05AA9 93B4EA98 8D8FDDC1 86FFB7DC 90A6C08F 4DF435C9 34063199 FFFFFFFF FFFFFFFF'
				)
			};
		case 6144:
			return {
				length: 6144,
				hash: 'SHA-256',
				g: h2i('05'),
				N: h2i(
					'FFFFFFFF FFFFFFFF C90FDAA2 2168C234 C4C6628B 80DC1CD1 29024E08 8A67CC74 020BBEA6 3B139B22 514A0879 8E3404DD EF9519B3 CD3A431B 302B0A6D F25F1437 4FE1356D 6D51C245 E485B576 625E7EC6 F44C42E9 A637ED6B 0BFF5CB6 F406B7ED EE386BFB 5A899FA5 AE9F2411 7C4B1FE6 49286651 ECE45B3D C2007CB8 A163BF05 98DA4836 1C55D39A 69163FA8 FD24CF5F 83655D23 DCA3AD96 1C62F356 208552BB 9ED52907 7096966D 670C354E 4ABC9804 F1746C08 CA18217C 32905E46 2E36CE3B E39E772C 180E8603 9B2783A2 EC07A28F B5C55DF0 6F4C52C9 DE2BCBF6 95581718 3995497C EA956AE5 15D22618 98FA0510 15728E5A 8AAAC42D AD33170D 04507A33 A85521AB DF1CBA64 ECFB8504 58DBEF0A 8AEA7157 5D060C7D B3970F85 A6E1E4C7 ABF5AE8C DB0933D7 1E8C94E0 4A25619D CEE3D226 1AD2EE6B F12FFA06 D98A0864 D8760273 3EC86A64 521F2B18 177B200C BBE11757 7A615D6C 770988C0 BAD946E2 08E24FA0 74E5AB31 43DB5BFC E0FD108E 4B82D120 A9210801 1A723C12 A787E6D7 88719A10 BDBA5B26 99C32718 6AF4E23C 1A946834 B6150BDA 2583E9CA 2AD44CE8 DBBBC2DB 04DE8EF9 2E8EFC14 1FBECAA6 287C5947 4E6BC05D 99B2964F A090C3A2 233BA186 515BE7ED 1F612970 CEE2D7AF B81BDD76 2170481C D0069127 D5B05AA9 93B4EA98 8D8FDDC1 86FFB7DC 90A6C08F 4DF435C9 34028492 36C3FAB4 D27C7026 C1D4DCB2 602646DE C9751E76 3DBA37BD F8FF9406 AD9E530E E5DB382F 413001AE B06A53ED 9027D831 179727B0 865A8918 DA3EDBEB CF9B14ED 44CE6CBA CED4BB1B DB7F1447 E6CC254B 33205151 2BD7AF42 6FB8F401 378CD2BF 5983CA01 C64B92EC F032EA15 D1721D03 F482D7CE 6E74FEF6 D55E702F 46980C82 B5A84031 900B1C9E 59E7C97F BEC7E8F3 23A97A7E 36CC88BE 0F1D45B7 FF585AC5 4BD407B2 2B4154AA CC8F6D7E BF48E1D8 14CC5ED2 0F8037E0 A79715EE F29BE328 06A1D58B B7C5DA76 F550AA3D 8A1FBFF0 EB19CCB1 A313D55C DA56C9EC 2EF29632 387FE8D7 6E3C0468 043E8F66 3F4860EE 12BF2D5B 0B7474D6 E694F91E 6DCC4024 FFFFFFFF FFFFFFFF'
				)
			};
		case 8192:
			return {
				length: 8192,
				hash: 'SHA-256',
				g: h2i('13'),
				N: h2i(
					'FFFFFFFF FFFFFFFF C90FDAA2 2168C234 C4C6628B 80DC1CD1 29024E08 8A67CC74 020BBEA6 3B139B22 514A0879 8E3404DD EF9519B3 CD3A431B 302B0A6D F25F1437 4FE1356D 6D51C245 E485B576 625E7EC6 F44C42E9 A637ED6B 0BFF5CB6 F406B7ED EE386BFB 5A899FA5 AE9F2411 7C4B1FE6 49286651 ECE45B3D C2007CB8 A163BF05 98DA4836 1C55D39A 69163FA8 FD24CF5F 83655D23 DCA3AD96 1C62F356 208552BB 9ED52907 7096966D 670C354E 4ABC9804 F1746C08 CA18217C 32905E46 2E36CE3B E39E772C 180E8603 9B2783A2 EC07A28F B5C55DF0 6F4C52C9 DE2BCBF6 95581718 3995497C EA956AE5 15D22618 98FA0510 15728E5A 8AAAC42D AD33170D 04507A33 A85521AB DF1CBA64 ECFB8504 58DBEF0A 8AEA7157 5D060C7D B3970F85 A6E1E4C7 ABF5AE8C DB0933D7 1E8C94E0 4A25619D CEE3D226 1AD2EE6B F12FFA06 D98A0864 D8760273 3EC86A64 521F2B18 177B200C BBE11757 7A615D6C 770988C0 BAD946E2 08E24FA0 74E5AB31 43DB5BFC E0FD108E 4B82D120 A9210801 1A723C12 A787E6D7 88719A10 BDBA5B26 99C32718 6AF4E23C 1A946834 B6150BDA 2583E9CA 2AD44CE8 DBBBC2DB 04DE8EF9 2E8EFC14 1FBECAA6 287C5947 4E6BC05D 99B2964F A090C3A2 233BA186 515BE7ED 1F612970 CEE2D7AF B81BDD76 2170481C D0069127 D5B05AA9 93B4EA98 8D8FDDC1 86FFB7DC 90A6C08F 4DF435C9 34028492 36C3FAB4 D27C7026 C1D4DCB2 602646DE C9751E76 3DBA37BD F8FF9406 AD9E530E E5DB382F 413001AE B06A53ED 9027D831 179727B0 865A8918 DA3EDBEB CF9B14ED 44CE6CBA CED4BB1B DB7F1447 E6CC254B 33205151 2BD7AF42 6FB8F401 378CD2BF 5983CA01 C64B92EC F032EA15 D1721D03 F482D7CE 6E74FEF6 D55E702F 46980C82 B5A84031 900B1C9E 59E7C97F BEC7E8F3 23A97A7E 36CC88BE 0F1D45B7 FF585AC5 4BD407B2 2B4154AA CC8F6D7E BF48E1D8 14CC5ED2 0F8037E0 A79715EE F29BE328 06A1D58B B7C5DA76 F550AA3D 8A1FBFF0 EB19CCB1 A313D55C DA56C9EC 2EF29632 387FE8D7 6E3C0468 043E8F66 3F4860EE 12BF2D5B 0B7474D6 E694F91E 6DBE1159 74A3926F 12FEE5E4 38777CB6 A932DF8C D8BEC4D0 73B931BA 3BC832B6 8D9DD300 741FA7BF 8AFC47ED 2576F693 6BA42466 3AAB639C 5AE4F568 3423B474 2BF1C978 238F16CB E39D652D E3FDB8BE FC848AD9 22222E04 A4037C07 13EB57A8 1A23F0C7 3473FC64 6CEA306B 4BCBC886 2F8385DD FA9D4B7F A2C087E8 79683303 ED5BDD3A 062B3CF5 B3A278A6 6D2A13F8 3F44F82D DF310EE0 74AB6A36 4597E899 A0255DC1 64F31CC5 0846851D F9AB4819 5DED7EA1 B1D510BD 7EE74D73 FAF36BC3 1ECFA268 359046F4 EB879F92 4009438B 481C6CD7 889A002E D5EE382B C9190DA6 FC026E47 9558E447 5677E9AA 9E3050E2 765694DF C81F56E8 80B96E71 60C980DD 98EDD3DF FFFFFFFF FFFFFFFF'
				)
			};

		default:
			throw 'invalid group length';
	}
}
