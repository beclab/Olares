import { DomainCookieRecord } from 'src/constant/constants';

export type CookieFormat = 'netscape' | 'header' | 'json' | 'unknown';

export interface ParseResult {
	cookies: Map<string, DomainCookieRecord[]>;
	invalidLines: string[];
}

export interface ICookieParser {
	readonly name: CookieFormat;

	detect(text: string): number;

	parse(text: string, domain?: string): ParseResult;

	parseWithStatistics(text: string, domain?: string): ParseStatistics;
}

export interface ParseStatistics {
	totalCount: number;
	validCount: number;
	invalidCount: number;
	byDomain: Map<string, number>;
	detectedFormat: CookieFormat;
}

export abstract class BaseCookieParser implements ICookieParser {
	abstract readonly name: CookieFormat;

	abstract detect(text: string): number;

	abstract parse(text: string, domain?: string): ParseResult;

	parseWithStatistics(text: string, domain?: string): ParseStatistics {
		if (!text || !text.trim()) {
			return {
				totalCount: 0,
				validCount: 0,
				invalidCount: 0,
				byDomain: new Map(),
				detectedFormat: 'unknown'
			};
		}
		const result = this.parse(text, domain);
		return this.generateStatistics(result);
	}

	protected generateStatistics(result: ParseResult): ParseStatistics {
		let validCount = 0;
		const byDomain = new Map<string, number>();

		result.cookies.forEach((cookies, domain) => {
			const count = cookies.length;
			validCount += count;
			byDomain.set(domain, count);
		});

		return {
			totalCount: validCount + result.invalidLines.length,
			validCount,
			invalidCount: result.invalidLines.length,
			byDomain,
			detectedFormat: this.name
		};
	}
}

export class NetscapeCookieParser extends BaseCookieParser {
	readonly name: CookieFormat = 'netscape';

	detect(text: string): number {
		if (!text || !text.trim()) {
			return 0;
		}

		const lines = text.split('\n').filter((line) => {
			const trimmed = line.trim();
			return trimmed && !trimmed.startsWith('#');
		});

		if (lines.length === 0) {
			return 0;
		}

		let score = 0;
		const sampleSize = Math.min(5, lines.length);

		for (const line of lines.slice(0, sampleSize)) {
			const trimmed = line.trim();

			// tab
			const tabCount = (trimmed.match(/\t/g) || []).length;
			if (tabCount >= 6) {
				score += 10;
			} else if (tabCount >= 3) {
				score += 5;
			}

			// style
			const fields = trimmed.split('\t');
			if (fields.length >= 7) {
				const field2 = fields[1].trim();
				const field4 = fields[3].trim();
				if (
					(field2 === 'TRUE' || field2 === 'FALSE') &&
					(field4 === 'TRUE' || field4 === 'FALSE')
				) {
					score += 8;
				}
			}
		}

		return Math.min(100, score);
	}

	parse(text: string): ParseResult {
		if (!text || !text.trim()) {
			throw new Error('Cookie text cannot be empty');
		}

		const result: ParseResult = {
			cookies: new Map<string, DomainCookieRecord[]>(),
			invalidLines: []
		};

		if (!text) {
			return result;
		}

		const lines = text.split('\n');

		for (const line of lines) {
			const trimmedLine = line.trim();
			if (!trimmedLine || trimmedLine.startsWith('#')) continue;

			const fields = trimmedLine.split('\t');
			if (fields.length < 7) {
				result.invalidLines.push(
					`Insufficient fields (need at least 7): ${trimmedLine}`
				);
				continue;
			}

			const domain = fields[0].trim();
			const name = fields[5].trim();
			const value = fields.slice(6).join('\t').trim();
			const expirationStr = fields[4].trim();

			if (!domain || !/^[a-zA-Z0-9.-]+$/.test(domain)) {
				result.invalidLines.push(
					`Invalid domain: ${domain} (line: ${trimmedLine})`
				);
				continue;
			}

			if (!name) {
				result.invalidLines.push(`Empty cookie name (line: ${trimmedLine})`);
				continue;
			}

			let expiration = 0;
			if (expirationStr) {
				expiration = Number(expirationStr);
				if (isNaN(expiration) || expiration < 0) {
					result.invalidLines.push(
						`Invalid expiration time: ${expirationStr} (line: ${trimmedLine})`
					);
					continue;
				}
			}

			const path = fields[2].trim() || '/';
			if (!/^\/.*$/.test(path)) {
				result.invalidLines.push(
					`Invalid path: ${path} (line: ${trimmedLine})`
				);
				continue;
			}

			const secure = fields[3].trim() === 'TRUE';

			const cookieRecord: DomainCookieRecord = {
				domain,
				name,
				value,
				expires: expiration,
				path,
				secure
			};

			const domainList = result.cookies.get(domain) || [];
			domainList.push(cookieRecord);
			result.cookies.set(domain, domainList);
		}

		if (result.invalidLines.length > 0) {
			console.error(
				`[NetscapeParser] Detected ${result.invalidLines.length} invalid cookie lines:`,
				result.invalidLines
			);
		}

		return result;
	}
}

export class HeaderCookieParser extends BaseCookieParser {
	readonly name: CookieFormat = 'header';

	detect(text: string): number {
		if (!text || !text.trim()) {
			return 0;
		}

		const lines = text.split('\n').filter((line) => {
			const trimmed = line.trim();
			return trimmed && !trimmed.startsWith('#');
		});

		if (lines.length === 0) {
			return 0;
		}

		let score = 0;
		const sampleSize = Math.min(5, lines.length);

		for (const line of lines.slice(0, sampleSize)) {
			const trimmed = line.trim();

			// Header "Set-Cookie:"
			if (trimmed.toLowerCase().startsWith('set-cookie:')) {
				score += 10;
				continue;
			}

			// style
			const hasHeaderAttributes =
				/;\s*(Domain|Path|Secure|HttpOnly|SameSite|Expires|Max-Age)\s*=/i.test(
					trimmed
				);
			if (hasHeaderAttributes) {
				score += 5;
			}

			// attribute
			if (trimmed.includes(';') && trimmed.includes('=')) {
				const parts = trimmed.split(';');
				if (parts.length >= 2) {
					score += 2;
				}
			}
		}

		return Math.min(100, score);
	}

	parse(text: string, inputDomain?: string): ParseResult {
		if (!text || !text.trim()) {
			throw new Error('Cookie text cannot be empty');
		}

		const result: ParseResult = {
			cookies: new Map<string, DomainCookieRecord[]>(),
			invalidLines: []
		};

		if (!text) {
			return result;
		}

		const lines = text.split('\n');

		for (const line of lines) {
			const trimmedLine = line.trim();
			if (!trimmedLine) continue;

			// remove "Set-Cookie:"
			let cookieLine = trimmedLine;
			if (cookieLine.toLowerCase().startsWith('set-cookie:')) {
				cookieLine = cookieLine.substring(11).trim();
			}

			// split the cookie body and its attributes
			const parts = cookieLine
				.split(';')
				.map((part) => part.trim())
				.filter((part) => part.length > 0);

			//browser style
			const hasStandardAttributes = parts.slice(1).some((part) => {
				const lower = part.toLowerCase();
				return (
					lower.startsWith('domain=') ||
					lower.startsWith('path=') ||
					lower.startsWith('expires=') ||
					lower.startsWith('max-age=') ||
					lower === 'secure' ||
					lower === 'httponly' ||
					lower.startsWith('samesite=')
				);
			});

			const allPartsHaveEquals = parts.every((part) => part.includes('='));

			if (parts.length > 1 && allPartsHaveEquals && !hasStandardAttributes) {
				for (const part of parts) {
					const equalIndex = part.indexOf('=');
					if (equalIndex === -1) continue;

					const name = part.substring(0, equalIndex).trim();
					const value = part.substring(equalIndex + 1).trim();

					if (!name) continue;

					const domain = inputDomain || '';
					if (!domain) {
						result.invalidLines.push(
							`Missing domain for browser cookie: ${part}`
						);
						continue;
					}

					const cookieRecord: DomainCookieRecord = {
						domain,
						name,
						value,
						expires: 0,
						path: '/',
						secure: false,
						httpOnly: false,
						sameSite: undefined
					};

					const domainList = result.cookies.get(domain) || [];
					domainList.push(cookieRecord);
					result.cookies.set(domain, domainList);
				}
				continue;
			}

			if (parts.length === 0) continue;

			// fist: name=value
			const firstPart = parts[0];
			const equalIndex = firstPart.indexOf('=');
			if (equalIndex === -1) {
				result.invalidLines.push(
					`Invalid cookie format (no '='): ${trimmedLine}`
				);
				continue;
			}

			const name = firstPart.substring(0, equalIndex).trim();
			const value = firstPart.substring(equalIndex + 1).trim();

			if (!name) {
				result.invalidLines.push(`Empty cookie name: ${trimmedLine}`);
				continue;
			}

			// parse attributes
			let domain = '';
			let path = '/';
			let secure = false;
			let httpOnly = false;
			let sameSite: string | undefined = undefined;
			let expires: number | string | Date = 0;

			for (let i = 1; i < parts.length; i++) {
				const attr = parts[i];
				const attrLower = attr.toLowerCase();

				if (attrLower === 'secure') {
					secure = true;
				} else if (attrLower === 'httponly') {
					httpOnly = true;
				} else if (attrLower.startsWith('domain=')) {
					domain = attr.substring(7).trim();
				} else if (attrLower.startsWith('path=')) {
					path = attr.substring(5).trim();
				} else if (attrLower.startsWith('samesite=')) {
					sameSite = attr.substring(9).trim();
				} else if (attrLower.startsWith('expires=')) {
					const expiresStr = attr.substring(8).trim();
					try {
						const date = new Date(expiresStr);
						if (!isNaN(date.getTime())) {
							expires = Math.floor(date.getTime() / 1000);
						}
					} catch (e) {
						result.invalidLines.push(`Invalid expires date: ${expiresStr}`);
					}
				} else if (attrLower.startsWith('max-age=')) {
					const maxAge = parseInt(attr.substring(8).trim());
					if (!isNaN(maxAge)) {
						expires = Math.floor(Date.now() / 1000) + maxAge;
					}
				}
			}

			//force
			if (inputDomain) {
				domain = inputDomain;
			}

			if (!domain) {
				result.invalidLines.push(`Missing domain attribute: ${trimmedLine}`);
				continue;
			}

			const cookieRecord: DomainCookieRecord = {
				domain,
				name,
				value,
				expires,
				path,
				secure,
				httpOnly,
				sameSite
			};

			const domainList = result.cookies.get(domain) || [];
			domainList.push(cookieRecord);
			result.cookies.set(domain, domainList);
		}

		if (result.invalidLines.length > 0) {
			console.warn(
				`[HeaderParser] Detected ${result.invalidLines.length} invalid cookie lines:`,
				result.invalidLines
			);
		}

		return result;
	}
}

export class JsonCookieParser extends BaseCookieParser {
	readonly name: CookieFormat = 'json';

	detect(text: string): number {
		if (!text || !text.trim()) {
			return 0;
		}

		const trimmed = text.trim();

		// json style
		if (!trimmed.startsWith('{') && !trimmed.startsWith('[')) {
			return 0;
		}

		try {
			JSON.parse(trimmed);
			return 90;
		} catch {
			return 0;
		}
	}

	parse(text: string): ParseResult {
		if (!text || !text.trim()) {
			throw new Error('Cookie text cannot be empty');
		}

		const result: ParseResult = {
			cookies: new Map<string, DomainCookieRecord[]>(),
			invalidLines: []
		};

		if (!text || !text.trim()) {
			return result;
		}

		try {
			const data = JSON.parse(text);

			// multi-style
			let cookieArray: any[] = [];

			if (Array.isArray(data)) {
				cookieArray = data;
			} else if (data.cookies && Array.isArray(data.cookies)) {
				cookieArray = data.cookies;
			} else if (typeof data === 'object') {
				// single cookie
				cookieArray = [data];
			}

			for (const item of cookieArray) {
				try {
					const cookieRecord = this.parseCookieObject(item);
					if (cookieRecord) {
						const domain = cookieRecord.domain;
						const domainList = result.cookies.get(domain) || [];
						domainList.push(cookieRecord);
						result.cookies.set(domain, domainList);
					}
				} catch (error) {
					result.invalidLines.push(
						`Invalid cookie object: ${JSON.stringify(item)}, error: ${error}`
					);
				}
			}
		} catch (error) {
			result.invalidLines.push(`Failed to parse JSON: ${error}`);
		}

		if (result.invalidLines.length > 0) {
			console.warn(
				`[JsonParser] Detected ${result.invalidLines.length} invalid cookies:`,
				result.invalidLines
			);
		}

		return result;
	}

	private parseCookieObject(obj: any): DomainCookieRecord | null {
		if (!obj.name || !obj.domain) {
			throw new Error('Missing required fields: name or domain');
		}

		// parse time
		let expires: number | string | Date = 0;
		if (obj.expires !== undefined) {
			if (typeof obj.expires === 'number') {
				expires = obj.expires;
			} else if (typeof obj.expires === 'string') {
				const date = new Date(obj.expires);
				if (!isNaN(date.getTime())) {
					expires = Math.floor(date.getTime() / 1000);
				}
			} else if (obj.expires instanceof Date) {
				expires = Math.floor(obj.expires.getTime() / 1000);
			}
		} else if (obj.expirationDate !== undefined) {
			// be compatible with the export formats of certain browsers
			expires = typeof obj.expirationDate === 'number' ? obj.expirationDate : 0;
		}

		const domain = obj.domain.trim();

		return {
			domain,
			name: obj.name,
			value: obj.value !== undefined ? obj.value : '',
			expires,
			path: obj.path || '/',
			secure: obj.secure === true,
			httpOnly: obj.httpOnly === true,
			sameSite: obj.sameSite
		};
	}
}

export class CookieParserFactory {
	private static parsers: ICookieParser[] = [
		new NetscapeCookieParser(),
		new HeaderCookieParser(),
		new JsonCookieParser()
	];

	static registerParser(parser: ICookieParser): void {
		this.parsers.push(parser);
	}

	static getAllParsers(): ICookieParser[] {
		return [...this.parsers];
	}

	static detectParser(text: string): ICookieParser | null {
		if (!text || !text.trim()) {
			return null;
		}

		let bestParser: ICookieParser | null = null;
		let bestScore = 0;

		for (const parser of this.parsers) {
			const score = parser.detect(text);
			if (score > bestScore) {
				bestScore = score;
				bestParser = parser;
			}
		}

		// at least 5 score
		return bestScore >= 5 ? bestParser : null;
	}

	static getParser(format: CookieFormat): ICookieParser | null {
		return this.parsers.find((p) => p.name === format) || null;
	}

	static parse(text: string): ParseResult {
		if (!text || !text.trim()) {
			throw new Error('Cookie text cannot be empty');
		}
		const parser = this.detectParser(text);

		if (!parser) {
			console.warn('Unable to detect cookie format, trying all parsers...');

			for (const p of this.parsers) {
				try {
					const result = p.parse(text);
					if (result.cookies.size > 0) {
						console.log(`Successfully parsed with ${p.name} parser`);
						return result;
					}
				} catch (error) {
					console.warn(`Failed to parse with ${p.name} parser:`, error);
				}
			}

			throw new Error('Unable to parse cookies with any available parser');
		}

		console.log(`Detected format: ${parser.name}`);
		return parser.parse(text);
	}

	static parseWithStatistics(text: string): ParseStatistics {
		if (!text || !text.trim()) {
			return {
				totalCount: 0,
				validCount: 0,
				invalidCount: 0,
				byDomain: new Map(),
				detectedFormat: 'unknown'
			};
		}

		const parser = this.detectParser(text);

		if (!parser) {
			for (const p of this.parsers) {
				try {
					return p.parseWithStatistics(text);
				} catch (error) {
					console.warn(`Failed to parse with ${p.name} parser:`, error);
				}
			}

			return {
				totalCount: 0,
				validCount: 0,
				invalidCount: 0,
				byDomain: new Map(),
				detectedFormat: 'unknown'
			};
		}

		return parser.parseWithStatistics(text);
	}
}
