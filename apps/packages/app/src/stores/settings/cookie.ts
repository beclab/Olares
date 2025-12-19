import {
	createCookieList,
	createOrUpdateCookie,
	deleteCookie,
	getCookies
} from 'src/api/settings/cookie';
import { defineStore } from 'pinia';
import {
	AggregatedDomain,
	DomainCookie,
	DomainCookieRecord
} from 'src/constant/constants';
import psl from 'psl';
import { CookieParserFactory, CookieFormat } from './cookieParser';

export type CookieState = {
	base_url: string;
	account: string;
	aggregatedDomainMap: Map<string, AggregatedDomain>;
	loading: boolean;
};

export const useCookieStore = defineStore('cookie', {
	state: () => {
		return {
			base_url: '',
			account: '',
			aggregatedDomainMap: new Map<string, AggregatedDomain>(),
			loading: false
		} as CookieState;
	},

	getters: {
		getAggregatedDomainList(state): AggregatedDomain[] {
			return Array.from(state.aggregatedDomainMap.values());
		}
	},

	actions: {
		async init(account: string, base_url = '', fetch = false) {
			this.account = account;
			this.base_url = base_url;
			if (fetch) {
				await this.getAllCookies();
			}
		},

		getDomainCookies(mainDomain: string): DomainCookie[] {
			const aggregatedDomain = this.aggregatedDomainMap.get(mainDomain);
			return aggregatedDomain?.subDomains ?? [];
		},

		async deleteAggregatedDomain(mainDomain: string) {
			const aggregatedDomain = this.aggregatedDomainMap.get(mainDomain);
			if (aggregatedDomain?.subDomains.length) {
				await Promise.all(
					aggregatedDomain.subDomains.map((domain) =>
						deleteCookie(this.base_url, domain.get_store_key())
					)
				);
			}
			return this.aggregatedDomainMap.delete(mainDomain);
		},

		async deleteDomainCookie(mainDomain: string, subDomain: string) {
			const aggregatedDomain = this.aggregatedDomainMap.get(mainDomain);
			if (!aggregatedDomain) return;

			const targetCookie = aggregatedDomain.subDomains.find(
				(domain) => domain.domain === subDomain
			);
			if (targetCookie) {
				await deleteCookie(this.base_url, targetCookie.get_store_key());
				aggregatedDomain.subDomains = aggregatedDomain.subDomains.filter(
					(domain) => domain.domain !== subDomain
				);
			}
		},

		async updateDomainCookie(domain: string, records: DomainCookieRecord[]) {
			if (!domain) {
				throw new Error('need domain');
			}

			const domainCookie = new DomainCookie({
				account: this.account,
				domain: domain,
				records: records,
				updateTime: Date.now()
			});

			const currentStoreKey = domainCookie.get_store_key();

			await createOrUpdateCookie(this.base_url, domainCookie);

			console.log(domainCookie);

			const mainDomain = this._getMainDomain(domainCookie.domain);
			const aggregatedDomain = this.aggregatedDomainMap.get(mainDomain);

			if (aggregatedDomain) {
				const existingIndex = aggregatedDomain.subDomains.findIndex(
					(oldCookie) => oldCookie.get_store_key() === currentStoreKey
				);

				if (existingIndex > -1) {
					aggregatedDomain.subDomains[existingIndex] = domainCookie;
				} else {
					aggregatedDomain.subDomains.push(domainCookie);
				}
			} else {
				this.aggregatedDomainMap.set(mainDomain, {
					mainDomain,
					subDomains: [domainCookie]
				});
			}

			return domainCookie;
		},

		async getAllCookies() {
			this.aggregatedDomainMap.clear();
			this.loading = true;
			try {
				const cookieList = await getCookies(this.base_url);
				if (cookieList?.length) {
					for (let i = 0; i < cookieList.length; i++) {
						const cookie = cookieList[i];
						const domainCookie = new DomainCookie(cookie);

						if (!domainCookie.domain || !domainCookie.account) {
							continue;
						}

						if (domainCookie.account !== this.account) {
							continue;
						}

						const mainDomain = this._getMainDomain(domainCookie.domain);
						if (!this.aggregatedDomainMap.has(mainDomain)) {
							this.aggregatedDomainMap.set(mainDomain, {
								mainDomain,
								subDomains: []
							});
						}
						this.aggregatedDomainMap
							.get(mainDomain)!
							.subDomains.push(domainCookie);
					}
				}
			} catch (error) {
				//
			}
			this.loading = false;
		},

		hasDomain(domain: string, checkSubDomain = false): boolean {
			if (!domain) return false;

			const cleanedDomain = domain.trim().replace(/^\.+/, '');

			if (checkSubDomain) {
				const mainDomain = this._getMainDomain(cleanedDomain);
				const aggregatedDomain = this.aggregatedDomainMap.get(mainDomain);

				if (!aggregatedDomain) return false;

				return aggregatedDomain.subDomains.some(
					(cookie) => cookie.domain === cleanedDomain
				);
			} else {
				const mainDomain = this._getMainDomain(cleanedDomain);
				return this.aggregatedDomainMap.has(mainDomain);
			}
		},

		getCookieCount(cookieText: string): number {
			try {
				const stats = CookieParserFactory.parseWithStatistics(cookieText);
				return stats.validCount;
			} catch (error) {
				console.error('Failed to count cookies:', error);
				return 0;
			}
		},

		async addCookies(cookieText: string): Promise<DomainCookie[]> {
			try {
				const parseResult = CookieParserFactory.parse(cookieText);
				return await this._uploadParsedCookies(parseResult);
			} catch (error) {
				return Promise.reject(error);
			}
		},

		getNetscapeCookieCount(cookieText: string): number {
			try {
				const parser = CookieParserFactory.getParser('netscape');
				if (!parser) {
					console.error('Netscape parser not found');
					return 0;
				}

				const stats = parser.parseWithStatistics(cookieText);
				console.log(stats);
				return stats.validCount;
			} catch (error) {
				console.error('Failed to count Netscape cookies:', error);
				return 0;
			}
		},

		getHeaderCookieCount(cookieText: string, domain?: string): number {
			try {
				const parser = CookieParserFactory.getParser('header');
				if (!parser) {
					console.error('Header parser not found');
					return 0;
				}

				const stats = parser.parseWithStatistics(cookieText, domain);
				console.log(stats);
				return stats.validCount;
			} catch (error) {
				console.error('Failed to count Header cookies:', error);
				return 0;
			}
		},

		getJsonCookieCount(cookieText: string): number {
			try {
				const parser = CookieParserFactory.getParser('json');
				if (!parser) {
					console.error('JSON parser not found');
					return 0;
				}

				const stats = parser.parseWithStatistics(cookieText);
				console.log(stats);
				return stats.validCount;
			} catch (error) {
				console.error('Failed to count JSON cookies:', error);
				return 0;
			}
		},

		addNetscapeCookies(netscapeText: string): Promise<DomainCookie[]> {
			try {
				const parser = CookieParserFactory.getParser('netscape');
				if (!parser) {
					throw new Error('Netscape parser not found');
				}
				const parseResult = parser.parse(netscapeText);
				console.log('addNetscapeCookies', parseResult);
				return this._uploadParsedCookies(parseResult);
			} catch (error) {
				return Promise.reject(error);
			}
		},

		addHeaderCookies(
			headerText: string,
			domain: string
		): Promise<DomainCookie[]> {
			try {
				const valid = psl.isValid(domain);
				if (!valid) {
					throw new Error('The domain is invalid.');
				}
				const parser = CookieParserFactory.getParser('header');
				if (!parser) {
					throw new Error('Header parser not found');
				}
				if (!domain || !domain.trim()) {
					throw new Error('domain cannot be empty');
				}
				const parseResult = parser.parse(headerText, domain);
				console.log('addHeaderCookies', parseResult);
				return this._uploadParsedCookies(parseResult);
			} catch (error) {
				return Promise.reject(error);
			}
		},

		addJsonCookies(jsonText: string): Promise<DomainCookie[]> {
			try {
				const parser = CookieParserFactory.getParser('json');
				if (!parser) {
					throw new Error('JSON parser not found');
				}
				const parseResult = parser.parse(jsonText);
				console.log('addJsonCookies', parseResult);
				return this._uploadParsedCookies(parseResult);
			} catch (error) {
				return Promise.reject(error);
			}
		},

		detectCookieFormat(cookieText: string): CookieFormat {
			const parser = CookieParserFactory.detectParser(cookieText);
			return parser ? parser.name : 'unknown';
		},

		async _uploadParsedCookies(parseResult: any): Promise<DomainCookie[]> {
			if (!this.account) {
				throw new Error('account not found');
			}
			if (parseResult.cookies.size === 0) {
				throw new Error('No valid cookies found');
			}

			return await parseResult.cookies
				.entries()
				.reduce(async (prevPromise, [domain, cookieData]) => {
					const results = await prevPromise;
					const currentResult = await this.updateDomainCookie(
						domain,
						cookieData
					);
					return [...results, currentResult];
				}, Promise.resolve([] as DomainCookie[]));
		},

		_getMainDomain(domain: string): string {
			const cleanedDomain = domain.trim().replace(/^\.+/, '');
			try {
				// const parsed = psl.parse(cleanedDomain);
				// if ('error' in parsed) {
				// 	const message = `domain parse error: ${domain}`;
				// 	console.error(message, parsed);
				// 	throw new Error(message);
				// }
				// console.log(parsed);
				// return parsed.domain || domain;
				const mainDomain = psl.get(cleanedDomain);
				if (!mainDomain) {
					throw new Error('psl get main domain error');
				}
				console.log(mainDomain);
				return mainDomain;
			} catch (error) {
				console.error(`Could not parse domain ${domain}`, error);
				if (cleanedDomain.includes('.')) {
					const parts = cleanedDomain.split('.').filter(Boolean);
					if (parts.length <= 2) return cleanedDomain;
					return parts.slice(-2).join('.');
				} else {
					console.error(`Could not split domain ${domain}`, error);
					return cleanedDomain;
				}
			}
		}
	}
});
