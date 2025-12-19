export function formatDate(date: Date | string | number) {
	return new Intl.DateTimeFormat().format(new Date(date));
}

export function formatDateTime(date: Date | string | number) {
	return new Intl.DateTimeFormat(undefined, {
		dateStyle: 'short',
		timeStyle: 'medium'
	} as any).format(new Date(date));
}

export async function passwordStrength(
	pwd: string
): Promise<{ score: number }> {
	// @ts-ignore
	const { default: zxcvbn } = await import(
		/* webpackChunkName: "zxcvbn" */ 'zxcvbn'
	);
	return zxcvbn(pwd);
}

export function toggleAttribute(el: Element, attr: string, on: boolean) {
	if (on) {
		el.setAttribute(attr, '');
	} else {
		el.removeAttribute(attr);
	}
}

export function mediaType(mimeType: string) {
	const match = mimeType.match(/(.*)\/(.*)/);
	const [, type, subtype] = match || ['', '', ''];

	switch (type) {
		case 'video':
			return 'video';
		case 'audio':
			return 'audio';
		case 'image':
			return 'image';
		case 'text':
			switch (subtype) {
				case 'csv':
					// return "csv";
					break;
				case 'plain':
					return 'text';
				default:
					return 'code';
			}
			break;
		case 'application':
			switch (subtype) {
				case 'pdf':
					return 'pdf';
				case 'json':
					return 'code';
				case 'pkcs8':
				case 'pkcs10':
				case 'pkix-cert':
				case 'pkix-crl':
				case 'pkcs7-mime':
				case 'x-x509-ca-cert':
				case 'x-x509-user-cert':
				case 'x-pkcs12':
				case 'x-pkcs7-certificates':
				case 'x-pkcs7-mime':
				case 'x-pkcs7-crl':
				case 'x-pem-file':
				case 'x-pkcs7-certreqresp':
					return 'certificate';
				case 'zip':
				case 'x-7z-compressed':
				case 'x-freearc':
				case 'x-bzip':
				case 'x-bzip2':
				case 'java-archive':
				case 'x-rar-compressed':
				case 'x-tar':
					return 'archive';
			}
			break;
		default:
			return '';
	}
}

export function fileIcon(mimeType: string) {
	const mType = mediaType(mimeType);
	return mType ? `file-${mType}` : 'file';
}

export function fileSize(size = 0) {
	return size < 1e6
		? Math.ceil(size / 10) / 100 + ' KB'
		: Math.ceil(size / 10000) / 100 + ' MB';
}

export function mask(value: string): string {
	return value && value.replace(/[^\n]/g, '\u2022');
}

let translateMethod: {
	method: (key: string, params?: any) => string;
	prefix?: string;
};

export const registerTranslateMethod = (
	translate: (key: string, params: any) => string,
	prefix?: string
) => {
	if (typeof translate !== 'function') {
		throw new Error('translate must be a function');
	}
	translateMethod = {
		method: translate,
		prefix
	};
};

export const translate = (key: string, params?: any) => {
	if (!translateMethod) {
		return key;
	}
	const value = translateMethod.method(
		translateMethod.prefix ? translateMethod.prefix + key : key,
		params
	);

	if (value == translateMethod.prefix + key) {
		return key;
	}
	return value;
};
