/** @fileOverview Javascript cryptography implementation.
 *
 * Crush to remove comments, shorten variable names and
 * generally reduce transmission size.
 *
 * @author Emily Stark
 * @author Mike Hamburg
 * @author Dan Boneh
 */

/*jslint indent: 2, bitwise: false, nomen: false, plusplus: false, white: false, regexp: false */
/*global document, window, escape, unescape, module, require, Uint32Array */

/**
 * The Stanford Javascript Crypto Library, top-level namespace.
 * @namespace
 */
var sjcl = {
	/**
	 * Symmetric ciphers.
	 * @namespace
	 */
	cipher: {},

	/**
	 * Hash functions.  Right now only SHA256 is implemented.
	 * @namespace
	 */
	hash: {},

	/**
	 * Key exchange functions.  Right now only SRP is implemented.
	 * @namespace
	 */
	keyexchange: {},

	/**
	 * Cipher modes of operation.
	 * @namespace
	 */
	mode: {},

	/**
	 * Miscellaneous.  HMAC and PBKDF2.
	 * @namespace
	 */
	misc: {},

	/**
	 * Bit array encoders and decoders.
	 * @namespace
	 *
	 * @description
	 * The members of this namespace are functions which translate between
	 * SJCL's bitArrays and other objects (usually strings).  Because it
	 * isn't always clear which direction is encoding and which is decoding,
	 * the method names are "fromBits" and "toBits".
	 */
	codec: {},

	/**
	 * Exceptions.
	 * @namespace
	 */
	exception: {
		/**
		 * Ciphertext is corrupt.
		 * @constructor
		 */
		corrupt: function (message) {
			this.toString = function () {
				return 'CORRUPT: ' + this.message;
			};
			this.message = message;
		},

		/**
		 * Invalid parameter.
		 * @constructor
		 */
		invalid: function (message) {
			this.toString = function () {
				return 'INVALID: ' + this.message;
			};
			this.message = message;
		},

		/**
		 * Bug or missing feature in SJCL.
		 * @constructor
		 */
		bug: function (message) {
			this.toString = function () {
				return 'BUG: ' + this.message;
			};
			this.message = message;
		},

		/**
		 * Something isn't ready.
		 * @constructor
		 */
		notReady: function (message) {
			this.toString = function () {
				return 'NOT READY: ' + this.message;
			};
			this.message = message;
		}
	}
};
/** @fileOverview Low-level AES implementation.
 *
 * This file contains a low-level implementation of AES, optimized for
 * size and for efficiency on several browsers.  It is based on
 * OpenSSL's aes_core.c, a public-domain implementation by Vincent
 * Rijmen, Antoon Bosselaers and Paulo Barreto.
 *
 * An older version of this implementation is available in the public
 * domain, but this one is (c) Emily Stark, Mike Hamburg, Dan Boneh,
 * Stanford University 2008-2010 and BSD-licensed for liability
 * reasons.
 *
 * @author Emily Stark
 * @author Mike Hamburg
 * @author Dan Boneh
 */

/**
 * Schedule out an AES key for both encryption and decryption.  This
 * is a low-level class.  Use a cipher mode to do bulk encryption.
 *
 * @constructor
 * @param {Array} key The key as an array of 4, 6 or 8 words.
 */
sjcl.cipher.aes = function (key) {
	if (!this._tables[0][0][0]) {
		this._precompute();
	}

	var i,
		j,
		tmp,
		encKey,
		decKey,
		sbox = this._tables[0][4],
		decTable = this._tables[1],
		keyLen = key.length,
		rcon = 1;

	if (keyLen !== 4 && keyLen !== 6 && keyLen !== 8) {
		throw new sjcl.exception.invalid('invalid aes key size');
	}

	this._key = [(encKey = key.slice(0)), (decKey = [])];

	// schedule encryption keys
	for (i = keyLen; i < 4 * keyLen + 28; i++) {
		tmp = encKey[i - 1];

		// apply sbox
		if (i % keyLen === 0 || (keyLen === 8 && i % keyLen === 4)) {
			tmp =
				(sbox[tmp >>> 24] << 24) ^
				(sbox[(tmp >> 16) & 255] << 16) ^
				(sbox[(tmp >> 8) & 255] << 8) ^
				sbox[tmp & 255];

			// shift rows and add rcon
			if (i % keyLen === 0) {
				tmp = (tmp << 8) ^ (tmp >>> 24) ^ (rcon << 24);
				rcon = (rcon << 1) ^ ((rcon >> 7) * 283);
			}
		}

		encKey[i] = encKey[i - keyLen] ^ tmp;
	}

	// schedule decryption keys
	for (j = 0; i; j++, i--) {
		tmp = encKey[j & 3 ? i : i - 4];
		if (i <= 4 || j < 4) {
			decKey[j] = tmp;
		} else {
			decKey[j] =
				decTable[0][sbox[tmp >>> 24]] ^
				decTable[1][sbox[(tmp >> 16) & 255]] ^
				decTable[2][sbox[(tmp >> 8) & 255]] ^
				decTable[3][sbox[tmp & 255]];
		}
	}
};

sjcl.cipher.aes.prototype = {
	// public
	/* Something like this might appear here eventually
  name: "AES",
  blockSize: 4,
  keySizes: [4,6,8],
  */

	/**
	 * Encrypt an array of 4 big-endian words.
	 * @param {Array} data The plaintext.
	 * @return {Array} The ciphertext.
	 */
	encrypt: function (data) {
		return this._crypt(data, 0);
	},

	/**
	 * Decrypt an array of 4 big-endian words.
	 * @param {Array} data The ciphertext.
	 * @return {Array} The plaintext.
	 */
	decrypt: function (data) {
		return this._crypt(data, 1);
	},

	/**
	 * The expanded S-box and inverse S-box tables.  These will be computed
	 * on the client so that we don't have to send them down the wire.
	 *
	 * There are two tables, _tables[0] is for encryption and
	 * _tables[1] is for decryption.
	 *
	 * The first 4 sub-tables are the expanded S-box with MixColumns.  The
	 * last (_tables[01][4]) is the S-box itself.
	 *
	 * @private
	 */
	_tables: [
		[[], [], [], [], []],
		[[], [], [], [], []]
	],

	/**
	 * Expand the S-box tables.
	 *
	 * @private
	 */
	_precompute: function () {
		var encTable = this._tables[0],
			decTable = this._tables[1],
			sbox = encTable[4],
			sboxInv = decTable[4],
			i,
			x,
			xInv,
			d = [],
			th = [],
			x2,
			x4,
			x8,
			s,
			tEnc,
			tDec;

		// Compute double and third tables
		for (i = 0; i < 256; i++) {
			th[(d[i] = (i << 1) ^ ((i >> 7) * 283)) ^ i] = i;
		}

		for (x = xInv = 0; !sbox[x]; x ^= x2 || 1, xInv = th[xInv] || 1) {
			// Compute sbox
			s = xInv ^ (xInv << 1) ^ (xInv << 2) ^ (xInv << 3) ^ (xInv << 4);
			s = (s >> 8) ^ (s & 255) ^ 99;
			sbox[x] = s;
			sboxInv[s] = x;

			// Compute MixColumns
			x8 = d[(x4 = d[(x2 = d[x])])];
			tDec =
				(x8 * 0x1010101) ^
				(x4 * 0x10001) ^
				(x2 * 0x101) ^
				(x * 0x1010100);
			tEnc = (d[s] * 0x101) ^ (s * 0x1010100);

			for (i = 0; i < 4; i++) {
				encTable[i][x] = tEnc = (tEnc << 24) ^ (tEnc >>> 8);
				decTable[i][s] = tDec = (tDec << 24) ^ (tDec >>> 8);
			}
		}

		// Compactify.  Considerable speedup on Firefox.
		for (i = 0; i < 5; i++) {
			encTable[i] = encTable[i].slice(0);
			decTable[i] = decTable[i].slice(0);
		}
	},

	/**
	 * Encryption and decryption core.
	 * @param {Array} input Four words to be encrypted or decrypted.
	 * @param dir The direction, 0 for encrypt and 1 for decrypt.
	 * @return {Array} The four encrypted or decrypted words.
	 * @private
	 */
	_crypt: function (input, dir) {
		if (input.length !== 4) {
			throw new sjcl.exception.invalid('invalid aes block size');
		}

		var key = this._key[dir],
			// state variables a,b,c,d are loaded with pre-whitened data
			a = input[0] ^ key[0],
			b = input[dir ? 3 : 1] ^ key[1],
			c = input[2] ^ key[2],
			d = input[dir ? 1 : 3] ^ key[3],
			a2,
			b2,
			c2,
			nInnerRounds = key.length / 4 - 2,
			i,
			kIndex = 4,
			out = [0, 0, 0, 0],
			table = this._tables[dir],
			// load up the tables
			t0 = table[0],
			t1 = table[1],
			t2 = table[2],
			t3 = table[3],
			sbox = table[4];

		// Inner rounds.  Cribbed from OpenSSL.
		for (i = 0; i < nInnerRounds; i++) {
			a2 =
				t0[a >>> 24] ^
				t1[(b >> 16) & 255] ^
				t2[(c >> 8) & 255] ^
				t3[d & 255] ^
				key[kIndex];
			b2 =
				t0[b >>> 24] ^
				t1[(c >> 16) & 255] ^
				t2[(d >> 8) & 255] ^
				t3[a & 255] ^
				key[kIndex + 1];
			c2 =
				t0[c >>> 24] ^
				t1[(d >> 16) & 255] ^
				t2[(a >> 8) & 255] ^
				t3[b & 255] ^
				key[kIndex + 2];
			d =
				t0[d >>> 24] ^
				t1[(a >> 16) & 255] ^
				t2[(b >> 8) & 255] ^
				t3[c & 255] ^
				key[kIndex + 3];
			kIndex += 4;
			a = a2;
			b = b2;
			c = c2;
		}

		// Last round.
		for (i = 0; i < 4; i++) {
			out[dir ? 3 & -i : i] =
				(sbox[a >>> 24] << 24) ^
				(sbox[(b >> 16) & 255] << 16) ^
				(sbox[(c >> 8) & 255] << 8) ^
				sbox[d & 255] ^
				key[kIndex++];
			a2 = a;
			a = b;
			b = c;
			c = d;
			d = a2;
		}

		return out;
	}
};

/** @fileOverview Arrays of bits, encoded as arrays of Numbers.
 *
 * @author Emily Stark
 * @author Mike Hamburg
 * @author Dan Boneh
 */

/**
 * Arrays of bits, encoded as arrays of Numbers.
 * @namespace
 * @description
 * <p>
 * These objects are the currency accepted by SJCL's crypto functions.
 * </p>
 *
 * <p>
 * Most of our crypto primitives operate on arrays of 4-byte words internally,
 * but many of them can take arguments that are not a multiple of 4 bytes.
 * This library encodes arrays of bits (whose size need not be a multiple of 8
 * bits) as arrays of 32-bit words.  The bits are packed, big-endian, into an
 * array of words, 32 bits at a time.  Since the words are double-precision
 * floating point numbers, they fit some extra data.  We use this (in a private,
 * possibly-changing manner) to encode the number of bits actually  present
 * in the last word of the array.
 * </p>
 *
 * <p>
 * Because bitwise ops clear this out-of-band data, these arrays can be passed
 * to ciphers like AES which want arrays of words.
 * </p>
 */
sjcl.bitArray = {
	/**
	 * Array slices in units of bits.
	 * @param {bitArray} a The array to slice.
	 * @param {Number} bstart The offset to the start of the slice, in bits.
	 * @param {Number} bend The offset to the end of the slice, in bits.  If this is undefined,
	 * slice until the end of the array.
	 * @return {bitArray} The requested slice.
	 */
	bitSlice: function (a, bstart, bend) {
		a = sjcl.bitArray
			._shiftRight(a.slice(bstart / 32), 32 - (bstart & 31))
			.slice(1);
		return bend === undefined ? a : sjcl.bitArray.clamp(a, bend - bstart);
	},

	/**
	 * Extract a number packed into a bit array.
	 * @param {bitArray} a The array to slice.
	 * @param {Number} bstart The offset to the start of the slice, in bits.
	 * @param {Number} blength The length of the number to extract.
	 * @return {Number} The requested slice.
	 */
	extract: function (a, bstart, blength) {
		// FIXME: this Math.floor is not necessary at all, but for some reason
		// seems to suppress a bug in the Chromium JIT.
		var x,
			sh = Math.floor((-bstart - blength) & 31);
		if (((bstart + blength - 1) ^ bstart) & -32) {
			// it crosses a boundary
			x =
				(a[(bstart / 32) | 0] << (32 - sh)) ^
				(a[(bstart / 32 + 1) | 0] >>> sh);
		} else {
			// within a single word
			x = a[(bstart / 32) | 0] >>> sh;
		}
		return x & ((1 << blength) - 1);
	},

	/**
	 * Concatenate two bit arrays.
	 * @param {bitArray} a1 The first array.
	 * @param {bitArray} a2 The second array.
	 * @return {bitArray} The concatenation of a1 and a2.
	 */
	concat: function (a1, a2) {
		if (a1.length === 0 || a2.length === 0) {
			return a1.concat(a2);
		}

		var last = a1[a1.length - 1],
			shift = sjcl.bitArray.getPartial(last);
		if (shift === 32) {
			return a1.concat(a2);
		} else {
			return sjcl.bitArray._shiftRight(
				a2,
				shift,
				last | 0,
				a1.slice(0, a1.length - 1)
			);
		}
	},

	/**
	 * Find the length of an array of bits.
	 * @param {bitArray} a The array.
	 * @return {Number} The length of a, in bits.
	 */
	bitLength: function (a) {
		var l = a.length,
			x;
		if (l === 0) {
			return 0;
		}
		x = a[l - 1];
		return (l - 1) * 32 + sjcl.bitArray.getPartial(x);
	},

	/**
	 * Truncate an array.
	 * @param {bitArray} a The array.
	 * @param {Number} len The length to truncate to, in bits.
	 * @return {bitArray} A new array, truncated to len bits.
	 */
	clamp: function (a, len) {
		if (a.length * 32 < len) {
			return a;
		}
		a = a.slice(0, Math.ceil(len / 32));
		var l = a.length;
		len = len & 31;
		if (l > 0 && len) {
			a[l - 1] = sjcl.bitArray.partial(
				len,
				a[l - 1] & (0x80000000 >> (len - 1)),
				1
			);
		}
		return a;
	},

	/**
	 * Make a partial word for a bit array.
	 * @param {Number} len The number of bits in the word.
	 * @param {Number} x The bits.
	 * @param {Number} [_end=0] Pass 1 if x has already been shifted to the high side.
	 * @return {Number} The partial word.
	 */
	partial: function (len, x, _end) {
		if (len === 32) {
			return x;
		}
		return (_end ? x | 0 : x << (32 - len)) + len * 0x10000000000;
	},

	/**
	 * Get the number of bits used by a partial word.
	 * @param {Number} x The partial word.
	 * @return {Number} The number of bits used by the partial word.
	 */
	getPartial: function (x) {
		return Math.round(x / 0x10000000000) || 32;
	},

	/**
	 * Compare two arrays for equality in a predictable amount of time.
	 * @param {bitArray} a The first array.
	 * @param {bitArray} b The second array.
	 * @return {boolean} true if a == b; false otherwise.
	 */
	equal: function (a, b) {
		if (sjcl.bitArray.bitLength(a) !== sjcl.bitArray.bitLength(b)) {
			return false;
		}
		var x = 0,
			i;
		for (i = 0; i < a.length; i++) {
			x |= a[i] ^ b[i];
		}
		return x === 0;
	},

	/** Shift an array right.
	 * @param {bitArray} a The array to shift.
	 * @param {Number} shift The number of bits to shift.
	 * @param {Number} [carry=0] A byte to carry in
	 * @param {bitArray} [out=[]] An array to prepend to the output.
	 * @private
	 */
	_shiftRight: function (a, shift, carry, out) {
		var i,
			last2 = 0,
			shift2;
		if (out === undefined) {
			out = [];
		}

		for (; shift >= 32; shift -= 32) {
			out.push(carry);
			carry = 0;
		}
		if (shift === 0) {
			return out.concat(a);
		}

		for (i = 0; i < a.length; i++) {
			out.push(carry | (a[i] >>> shift));
			carry = a[i] << (32 - shift);
		}
		last2 = a.length ? a[a.length - 1] : 0;
		shift2 = sjcl.bitArray.getPartial(last2);
		out.push(
			sjcl.bitArray.partial(
				(shift + shift2) & 31,
				shift + shift2 > 32 ? carry : out.pop(),
				1
			)
		);
		return out;
	},

	/** xor a block of 4 words together.
	 * @private
	 */
	_xor4: function (x, y) {
		return [x[0] ^ y[0], x[1] ^ y[1], x[2] ^ y[2], x[3] ^ y[3]];
	},

	/** byteswap a word array inplace.
	 * (does not handle partial words)
	 * @param {sjcl.bitArray} a word array
	 * @return {sjcl.bitArray} byteswapped array
	 */
	byteswapM: function (a) {
		var i,
			v,
			m = 0xff00;
		for (i = 0; i < a.length; ++i) {
			v = a[i];
			a[i] = (v >>> 24) | ((v >>> 8) & m) | ((v & m) << 8) | (v << 24);
		}
		return a;
	}
};
/** @fileOverview Bit array codec implementations.
 *
 * @author Emily Stark
 * @author Mike Hamburg
 * @author Dan Boneh
 */

/**
 * UTF-8 strings
 * @namespace
 */
sjcl.codec.utf8String = {
	/** Convert from a bitArray to a UTF-8 string. */
	fromBits: function (arr) {
		var out = '',
			bl = sjcl.bitArray.bitLength(arr),
			i,
			tmp;
		for (i = 0; i < bl / 8; i++) {
			if ((i & 3) === 0) {
				tmp = arr[i / 4];
			}
			out += String.fromCharCode(((tmp >>> 8) >>> 8) >>> 8);
			tmp <<= 8;
		}
		return decodeURIComponent(escape(out));
	},

	/** Convert from a UTF-8 string to a bitArray. */
	toBits: function (str) {
		str = unescape(encodeURIComponent(str));
		var out = [],
			i,
			tmp = 0;
		for (i = 0; i < str.length; i++) {
			tmp = (tmp << 8) | str.charCodeAt(i);
			if ((i & 3) === 3) {
				out.push(tmp);
				tmp = 0;
			}
		}
		if (i & 3) {
			out.push(sjcl.bitArray.partial(8 * (i & 3), tmp));
		}
		return out;
	}
};
/** @fileOverview Bit array codec implementations.
 *
 * @author Emily Stark
 * @author Mike Hamburg
 * @author Dan Boneh
 */

/**
 * Arrays of bytes
 * @namespace
 */
sjcl.codec.bytes = {
	/** Convert from a bitArray to an array of bytes. */
	fromBits: function (arr) {
		var out = [],
			bl = sjcl.bitArray.bitLength(arr),
			i,
			tmp;
		for (i = 0; i < bl / 8; i++) {
			if ((i & 3) === 0) {
				tmp = arr[i / 4];
			}
			out.push(tmp >>> 24);
			tmp <<= 8;
		}
		return out;
	},
	/** Convert from an array of bytes to a bitArray. */
	toBits: function (bytes) {
		var out = [],
			i,
			tmp = 0;
		for (i = 0; i < bytes.length; i++) {
			tmp = (tmp << 8) | bytes[i];
			if ((i & 3) === 3) {
				out.push(tmp);
				tmp = 0;
			}
		}
		if (i & 3) {
			out.push(sjcl.bitArray.partial(8 * (i & 3), tmp));
		}
		return out;
	}
};
/** @fileOverview CCM mode implementation.
 *
 * Special thanks to Roy Nicholson for pointing out a bug in our
 * implementation.
 *
 * @author Emily Stark
 * @author Mike Hamburg
 * @author Dan Boneh
 */

/**
 * CTR mode with CBC MAC.
 * @namespace
 */
sjcl.mode.ccm = {
	/** The name of the mode.
	 * @constant
	 */
	name: 'ccm',

	_progressListeners: [],

	listenProgress: function (cb) {
		sjcl.mode.ccm._progressListeners.push(cb);
	},

	unListenProgress: function (cb) {
		var index = sjcl.mode.ccm._progressListeners.indexOf(cb);
		if (index > -1) {
			sjcl.mode.ccm._progressListeners.splice(index, 1);
		}
	},

	_callProgressListener: function (val) {
		var p = sjcl.mode.ccm._progressListeners.slice(),
			i;

		for (i = 0; i < p.length; i += 1) {
			p[i](val);
		}
	},

	/** Encrypt in CCM mode.
	 * @static
	 * @param {Object} prf The pseudorandom function.  It must have a block size of 16 bytes.
	 * @param {bitArray} plaintext The plaintext data.
	 * @param {bitArray} iv The initialization value.
	 * @param {bitArray} [adata=[]] The authenticated data.
	 * @param {Number} [tlen=64] the desired tag length, in bits.
	 * @return {bitArray} The encrypted data, an array of bytes.
	 */
	encrypt: function (prf, plaintext, iv, adata, tlen) {
		var L,
			out = plaintext.slice(0),
			tag,
			w = sjcl.bitArray,
			ivl = w.bitLength(iv) / 8,
			ol = w.bitLength(out) / 8;
		tlen = tlen || 64;
		adata = adata || [];

		if (ivl < 7) {
			throw new sjcl.exception.invalid(
				'ccm: iv must be at least 7 bytes'
			);
		}

		// compute the length of the length
		for (L = 2; L < 4 && ol >>> (8 * L); L++) {}
		if (L < 15 - ivl) {
			L = 15 - ivl;
		}
		iv = w.clamp(iv, 8 * (15 - L));

		// compute the tag
		tag = sjcl.mode.ccm._computeTag(prf, plaintext, iv, adata, tlen, L);

		// encrypt
		out = sjcl.mode.ccm._ctrMode(prf, out, iv, tag, tlen, L);

		return w.concat(out.data, out.tag);
	},

	/** Decrypt in CCM mode.
	 * @static
	 * @param {Object} prf The pseudorandom function.  It must have a block size of 16 bytes.
	 * @param {bitArray} ciphertext The ciphertext data.
	 * @param {bitArray} iv The initialization value.
	 * @param {bitArray} [adata=[]] adata The authenticated data.
	 * @param {Number} [tlen=64] tlen the desired tag length, in bits.
	 * @return {bitArray} The decrypted data.
	 */
	decrypt: function (prf, ciphertext, iv, adata, tlen) {
		tlen = tlen || 64;
		adata = adata || [];
		var L,
			w = sjcl.bitArray,
			ivl = w.bitLength(iv) / 8,
			ol = w.bitLength(ciphertext),
			out = w.clamp(ciphertext, ol - tlen),
			tag = w.bitSlice(ciphertext, ol - tlen),
			tag2;

		ol = (ol - tlen) / 8;

		if (ivl < 7) {
			throw new sjcl.exception.invalid(
				'ccm: iv must be at least 7 bytes'
			);
		}

		// compute the length of the length
		for (L = 2; L < 4 && ol >>> (8 * L); L++) {}
		if (L < 15 - ivl) {
			L = 15 - ivl;
		}
		iv = w.clamp(iv, 8 * (15 - L));

		// decrypt
		out = sjcl.mode.ccm._ctrMode(prf, out, iv, tag, tlen, L);

		// check the tag
		tag2 = sjcl.mode.ccm._computeTag(prf, out.data, iv, adata, tlen, L);
		if (!w.equal(out.tag, tag2)) {
			throw new sjcl.exception.corrupt("ccm: tag doesn't match");
		}

		return out.data;
	},

	_macAdditionalData: function (prf, adata, iv, tlen, ol, L) {
		var mac,
			tmp,
			i,
			macData = [],
			w = sjcl.bitArray,
			xor = w._xor4;

		// mac the flags
		mac = [
			w.partial(
				8,
				(adata.length ? 1 << 6 : 0) | ((tlen - 2) << 2) | (L - 1)
			)
		];

		// mac the iv and length
		mac = w.concat(mac, iv);
		mac[3] |= ol;
		mac = prf.encrypt(mac);

		if (adata.length) {
			// mac the associated data.  start with its length...
			tmp = w.bitLength(adata) / 8;
			if (tmp <= 0xfeff) {
				macData = [w.partial(16, tmp)];
			} else if (tmp <= 0xffffffff) {
				macData = w.concat([w.partial(16, 0xfffe)], [tmp]);
			} // else ...

			// mac the data itself
			macData = w.concat(macData, adata);
			for (i = 0; i < macData.length; i += 4) {
				mac = prf.encrypt(
					xor(mac, macData.slice(i, i + 4).concat([0, 0, 0]))
				);
			}
		}

		return mac;
	},

	/* Compute the (unencrypted) authentication tag, according to the CCM specification
	 * @param {Object} prf The pseudorandom function.
	 * @param {bitArray} plaintext The plaintext data.
	 * @param {bitArray} iv The initialization value.
	 * @param {bitArray} adata The authenticated data.
	 * @param {Number} tlen the desired tag length, in bits.
	 * @return {bitArray} The tag, but not yet encrypted.
	 * @private
	 */
	_computeTag: function (prf, plaintext, iv, adata, tlen, L) {
		// compute B[0]
		var mac,
			i,
			w = sjcl.bitArray,
			xor = w._xor4;

		tlen /= 8;

		// check tag length and message length
		if (tlen % 2 || tlen < 4 || tlen > 16) {
			throw new sjcl.exception.invalid('ccm: invalid tag length');
		}

		if (adata.length > 0xffffffff || plaintext.length > 0xffffffff) {
			// I don't want to deal with extracting high words from doubles.
			throw new sjcl.exception.bug(
				"ccm: can't deal with 4GiB or more data"
			);
		}

		mac = sjcl.mode.ccm._macAdditionalData(
			prf,
			adata,
			iv,
			tlen,
			w.bitLength(plaintext) / 8,
			L
		);

		// mac the plaintext
		for (i = 0; i < plaintext.length; i += 4) {
			mac = prf.encrypt(
				xor(mac, plaintext.slice(i, i + 4).concat([0, 0, 0]))
			);
		}

		return w.clamp(mac, tlen * 8);
	},

	/** CCM CTR mode.
	 * Encrypt or decrypt data and tag with the prf in CCM-style CTR mode.
	 * May mutate its arguments.
	 * @param {Object} prf The PRF.
	 * @param {bitArray} data The data to be encrypted or decrypted.
	 * @param {bitArray} iv The initialization vector.
	 * @param {bitArray} tag The authentication tag.
	 * @param {Number} tlen The length of th etag, in bits.
	 * @param {Number} L The CCM L value.
	 * @return {Object} An object with data and tag, the en/decryption of data and tag values.
	 * @private
	 */
	_ctrMode: function (prf, data, iv, tag, tlen, L) {
		var enc,
			i,
			w = sjcl.bitArray,
			xor = w._xor4,
			ctr,
			l = data.length,
			bl = w.bitLength(data),
			n = l / 50,
			p = n;

		// start the ctr
		ctr = w
			.concat([w.partial(8, L - 1)], iv)
			.concat([0, 0, 0])
			.slice(0, 4);

		// en/decrypt the tag
		tag = w.bitSlice(xor(tag, prf.encrypt(ctr)), 0, tlen);

		// en/decrypt the data
		if (!l) {
			return { tag: tag, data: [] };
		}

		for (i = 0; i < l; i += 4) {
			if (i > n) {
				sjcl.mode.ccm._callProgressListener(i / l);
				n += p;
			}
			ctr[3]++;
			enc = prf.encrypt(ctr);
			data[i] ^= enc[0];
			data[i + 1] ^= enc[1];
			data[i + 2] ^= enc[2];
			data[i + 3] ^= enc[3];
		}
		return { tag: tag, data: w.clamp(data, bl) };
	}
};

export { sjcl };
