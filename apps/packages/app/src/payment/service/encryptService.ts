import { ref } from 'vue';
import { ENCRYPT_CONFIG } from './../config';
import { logger } from './../logger';
import type { OrderDataBase } from './../types';

class MessagePackLoader {
	private msgpack = ref<null | { encode: (data: unknown) => Uint8Array }>(null);

	async load(): Promise<{ encode: (data: unknown) => Uint8Array } | null> {
		if (this.msgpack.value) return this.msgpack.value;

		if (window.msgpack5) {
			this.msgpack.value = window.msgpack5();
			logger.log('✅ MessagePack library loaded (global)');
			return this.msgpack.value;
		}

		return new Promise((resolve, reject) => {
			const script = document.createElement('script');
			script.src =
				'https://cdn.jsdelivr.net/npm/msgpack5@4.0.2/dist/msgpack5.min.js';
			document.head.appendChild(script);

			script.onload = () => {
				if (window.msgpack5) {
					this.msgpack.value = window.msgpack5();
					logger.log('✅ MessagePack library loaded (CDN)');
					resolve(this.msgpack.value);
				} else {
					reject(new Error('MessagePack library failed to load'));
				}
			};

			script.onerror = () =>
				reject(new Error('MessagePack library CDN load failed'));
		});
	}
}

export class EncryptService {
	private msgpackLoader = new MessagePackLoader();
	private encryptedData = ref('');
	private shouldEncrypt = ref(true);

	private async importRSAPublicKey(
		publicKeyString: string
	): Promise<CryptoKey> {
		// Complete the PEM headers and footers if missing
		let pemString = publicKeyString.trim();
		if (!pemString.includes('-----BEGIN PUBLIC KEY-----')) {
			pemString = `-----BEGIN PUBLIC KEY-----\n${pemString}\n-----END PUBLIC KEY-----`;
		}

		// PEM to DER format
		const pemHeader = '-----BEGIN PUBLIC KEY-----';
		const pemFooter = '-----END PUBLIC KEY-----';
		const pemContents = pemString.substring(
			pemHeader.length,
			pemString.length - pemFooter.length
		);
		// Remove spaces and decode Base64
		const binaryDerString = atob(pemContents.replace(/\s/g, ''));
		const binaryDer = new Uint8Array(binaryDerString.length);
		for (let i = 0; i < binaryDerString.length; i++) {
			binaryDer[i] = binaryDerString.charCodeAt(i);
		}

		// Import public key (for encryption only)
		return crypto.subtle.importKey(
			'spki', // Subject Public Key Info
			binaryDer,
			{
				name: 'RSA-OAEP', // Encryption algorithm
				hash: ENCRYPT_CONFIG.RSA_HASH_ALG
			},
			false, // Non-extractable (security restriction)
			['encrypt'] // Only granted encryption permission
		);
	}

	/**
	 * RSA block encryption (handles large data to avoid exceeding encryption length limits)
	 * @param data Data to be encrypted (string/Buffer/Uint8Array)
	 * @param publicKey Imported RSA public key
	 */
	private async encryptWithRSA(
		data: string | Buffer | Uint8Array,
		publicKey: CryptoKey
	): Promise<string> {
		// Convert to Uint8Array
		let dataBytes: Uint8Array;
		if (typeof data === 'string') {
			dataBytes = new TextEncoder().encode(data);
		} else if (data instanceof Buffer) {
			dataBytes = new Uint8Array(data);
		} else {
			dataBytes = data;
		}

		// Block encryption (RSA-OAEP has length limitations)
		const maxLength = ENCRYPT_CONFIG.RSA_MAX_PLAINTEXT_LENGTH;
		const chunks: Uint8Array[] = [];

		for (let i = 0; i < dataBytes.length; i += maxLength) {
			const chunk = dataBytes.slice(i, i + maxLength);
			const encryptedChunk = await crypto.subtle.encrypt(
				{ name: 'RSA-OAEP' },
				publicKey,
				chunk
			);
			chunks.push(new Uint8Array(encryptedChunk));
		}

		// Merge chunks and convert to a hexadecimal string
		const totalLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
		const result = new Uint8Array(totalLength);
		let offset = 0;
		for (const chunk of chunks) {
			result.set(chunk, offset);
			offset += chunk.length;
		}

		return Array.from(result)
			.map((b) => b.toString(16).padStart(2, '0'))
			.join('');
	}

	/**
	 * Create Olares Pay header (8-byte fixed format)
	 * @param isEncrypted Whether it is encrypted (used for header identification)
	 */
	private createOlaresPayHeader(isEncrypted: boolean): string {
		const { identifier, version, reserved } = ENCRYPT_CONFIG.OLARES_PAY_HEADER;
		const header = new Uint8Array(8);

		// Fill in header fields
		header.set(identifier, 0); // Bytes 0-3: "OLRP" identifier
		header[4] = version; // Byte 4: version number
		header[5] = isEncrypted ? 0x01 : 0x00; // Byte 5: encryption flag
		header.set(reserved, 6); // Bytes 6-7: reserved fields

		// Convert to a hexadecimal string
		return Array.from(header)
			.map((b) => b.toString(16).padStart(2, '0'))
			.join('');
	}

	/**
	 * Process order data (MessagePack encoding + optional RSA encryption + Olares Pay format)
	 * @param orderData Original order data
	 * @param rsaPublicKey RSA public key (no encryption if empty)
	 */
	async processOrderData(
		orderData: OrderDataBase,
		rsaPublicKey: string
	): Promise<void> {
		logger.log('Starting order data processing...');
		try {
			// 1. Load the MessagePack library
			const msgpackInstance = await this.msgpackLoader.load();
			if (!msgpackInstance) {
				throw new Error('MessagePack library not available');
			}

			// 2. Encode order data with MessagePack
			const encodedData = msgpackInstance.encode(orderData);
			logger.log('✅ MessagePack encoding completed');

			// 3. Optional RSA encryption
			let finalData: string;
			const isEncrypted = this.shouldEncrypt.value && !!rsaPublicKey;
			if (isEncrypted) {
				const publicKey = await this.importRSAPublicKey(rsaPublicKey);
				logger.log('✅ RSA public key imported successfully');

				finalData = await this.encryptWithRSA(encodedData, publicKey);
				logger.log('✅ Order data encrypted with RSA');
			} else {
				// If not encrypted: directly convert to hexadecimal
				finalData = Array.from(encodedData)
					.map((b) => b.toString(16).padStart(2, '0'))
					.join('');
				logger.log('✅ Order data encoded to hex (no encryption)');
			}

			// 4. Generate Olares Pay header and merge with data
			const header = this.createOlaresPayHeader(isEncrypted);
			this.encryptedData.value = header + finalData;

			logger.log(`✅ Final Olares Pay data prepared`);
			logger.log(`Header length: 16 chars (8 bytes)`);
			logger.log(`Data length: ${finalData.length} chars`);
			logger.log(`Total length: ${this.encryptedData.value.length} chars`);
		} catch (error) {
			const errorMsg = error instanceof Error ? error.message : String(error);
			logger.log(`❌ Order data processing failed: ${errorMsg}`);
			throw error;
		}
	}

	setShouldEncrypt(value: boolean): void {
		this.shouldEncrypt.value = value;
		logger.log(`Encryption switch set to: ${value}`);
	}

	getEncryptedData() {
		return this.encryptedData;
	}

	getShouldEncrypt() {
		return this.shouldEncrypt;
	}
}

export const encryptService = new EncryptService();
