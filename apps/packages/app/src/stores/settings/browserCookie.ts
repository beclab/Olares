import { defineStore } from 'pinia';
import { browser } from 'src/platform/interface/bex/browser/target';
import parser from 'tld-extract';
import { useCookieStore } from './cookie';
import { notifyFailed, notifySuccess } from 'src/utils/settings/btNotify';
import { useCollectSiteStore } from '../collect-site';
import { DomainCookie } from 'src/constant/constants';
import { useCollectStore } from 'src/stores/collect';

export type BrowserCookieState = {
	domain: string;
	rootDomain: string;
	pushLoading: boolean;
	cookieList: any[];
	current_cookie: DomainCookie[];
	current_tab:
		| {
				url: string;
				title: string;
				favicon: string;
		  }
		| undefined;
	olaresId?: string;
	url?: string;
};

export const useBrowserCookieStore = defineStore('browserCookie', {
	state: () => {
		return {
			domain: '',
			rootDomain: '',
			pushLoading: false,
			cookieList: [],
			current_cookie: [],
			current_tab: undefined,
			olaresId: undefined,
			url: undefined
		} as BrowserCookieState;
	},

	getters: {
		hasCookie(state): boolean {
			const cookieStore = useCookieStore();
			return cookieStore.hasDomain(state.rootDomain);
		},
		hasExpiredRecords(state): boolean {
			const currentTime = Math.floor(Date.now() / 1000);
			console.log('mainDomain-3', state.current_cookie);

			return state.current_cookie.some((domainCookie) =>
				domainCookie.records?.some((record) => {
					const expires =
						typeof record.expires === 'number' ? record.expires : 0;
					return expires > 0 && currentTime > expires;
				})
			);
		},
		cookieStatus(): number {
			if (this.hasCookie) {
				return this.hasExpiredRecords ? 1 : 2;
			}
			return 0;
		}
	},

	actions: {
		async init(tab, olaresId, url) {
			this.current_tab = tab;
			this.olaresId = olaresId;
			this.url = url;
			const cookieStore = useCookieStore();
			this.resetCookieStatus();
			await cookieStore.init(olaresId, url, false);

			const urlNew = new URL(tab.url);
			this.domain = urlNew.hostname;
			this.rootDomain = this.getPrimaryDomain(urlNew.origin);
			this.getBrowserCookies(urlNew, olaresId);
		},
		getPrimaryDomain(url: string): string {
			const urlParser = parser(url);
			return urlParser.domain;
		},
		updateCookieStatus() {
			const collectStore = useCollectStore();
			collectStore.updateCookieStatus();
		},
		resetCookieStatus() {
			const collectStore = useCollectStore();
			collectStore.resetCookieStatus();
		},
		async getBrowserCookies(url: URL, userAccount?: string) {
			this.cookieList = [];
			// 	active: true,
			// 	currentWindow: true
			// });
			const tab = this.current_tab;
			if (tab?.url) {
				try {
					const cookies = await browser.cookies.getAll({
						domain: this.rootDomain
					});

					const cookieIgnores = ['auth_token', 'authelia_session'];
					const selfDomain = userAccount?.split('@').join('.');
					const isOlaresWeb = !selfDomain
						? false
						: tab?.url.includes(selfDomain);
					const urlFilterStr = `.${url.hostname}`;

					const data = cookies.filter((item) => {
						if (isOlaresWeb) {
							return (
								urlFilterStr.includes(item.domain) &&
								!cookieIgnores.includes(item.name)
							);
						} else {
							return urlFilterStr.includes(item.domain);
						}
					});

					console.log('Successfully get cookie:', data);
					this.cookieList = data.map((cookie) => ({
						name: cookie.name,
						value: cookie.value,
						domain: cookie.domain,
						path: cookie.path || '/',
						expires: cookie.expirationDate || 0,
						secure: cookie.secure || false,
						httpOnly: cookie.httpOnly || false,
						sameSite: cookie.sameSite
					}));
					console.log('cookieList', this.cookieList);
				} catch (e) {
					console.log('Error getting cookies:', e);
				}
			}
		},

		getCookietList(tab) {
			const url = new URL(tab.url);
			const rootDomain = this.getPrimaryDomain(url.origin);

			const cookieStore = useCookieStore();
			const mainDomain = cookieStore._getMainDomain(rootDomain);
			this.current_cookie = cookieStore.getDomainCookies(mainDomain);
			console.log('mainDomain', mainDomain);
			console.log('mainDomain-2', this.current_cookie);
		},
		async pushCookie() {
			const collectSiteStore = useCollectSiteStore();

			this.pushLoading = true;
			const cookieStore = useCookieStore();
			const cookieJson = JSON.stringify(this.cookieList);
			cookieStore
				.addJsonCookies(cookieJson)
				.then(() => {
					collectSiteStore.updateCookieStatus();
					this.updateCookieStatus();
					notifySuccess('success');
				})
				.catch((err) => {
					notifyFailed(err.message);
				})
				.finally(() => {
					this.pushLoading = false;
				});
		}
	}
});
