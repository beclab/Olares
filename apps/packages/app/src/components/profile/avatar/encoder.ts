import { base64url } from 'multiformats/bases/base64';

const textEncoder = new TextEncoder();
const textDecoder = new TextDecoder();

/**
 * Utility class for encoding/converting data into various formats.
 */
export class Encoder {
	public static base64UrlToBytes(base64urlString: string): Uint8Array {
		const content = base64url.baseDecode(base64urlString);
		return content;
	}

	public static base64UrlToObject(base64urlString: string): any {
		const payloadBytes = base64url.baseDecode(base64urlString);
		const payloadString = Encoder.bytesToString(payloadBytes);
		const payloadObject = JSON.parse(payloadString);
		return payloadObject;
	}

	public static bytesToBase64Url(bytes: Uint8Array): string {
		const base64UrlString = base64url.baseEncode(bytes);
		return base64UrlString;
	}

	public static bytesToString(content: Uint8Array): string {
		const bytes = textDecoder.decode(content);
		return bytes;
	}

	public static objectToBytes(obj: { [key: string]: any }): Uint8Array {
		const objectString = JSON.stringify(obj);
		const objectBytes = textEncoder.encode(objectString);
		return objectBytes;
	}

	public static stringToBase64Url(content: string): string {
		const bytes = textEncoder.encode(content);
		const base64UrlString = base64url.baseEncode(bytes);
		return base64UrlString;
	}

	public static stringToBytes(content: string): Uint8Array {
		const bytes = textEncoder.encode(content);
		return bytes;
	}
}

export function getNftUrl(avatarUrl: string): { url: string; isNft: boolean } {
	try {
		const vp = JSON.parse(avatarUrl);
		console.log(vp);
		if (vp) {
			const vcstr = Encoder.bytesToString(
				Encoder.base64UrlToBytes(vp.verifiableCredential![0].split('.')[1])
			);
			console.log(vcstr);
			const vc = JSON.parse(vcstr);
			console.log(vc);
			console.log(vc.vc.credentialSubject.image);
			let imageUrl = vc.vc.credentialSubject.image;
			if (imageUrl.startsWith('ipfs://')) {
				imageUrl = imageUrl.replace('ipfs://', 'https://gateway.ipfs.io/ipfs/');
			}

			console.log(imageUrl);
			return { url: imageUrl, isNft: true };
		} else {
			return { url: 'https://app.cdn.olares.com/avatar3/1.png', isNft: false };
		}
	} catch (e) {
		console.log(e);
		return { url: 'https://app.cdn.olares.com/avatar3/1.png', isNft: false };
	}
}
