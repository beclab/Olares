/* eslint-disable @typescript-eslint/no-unused-vars */
import { RouteLocationNormalizedLoaded } from 'vue-router';
import { Application } from './interface/application';
import { supportLanguages, languagesShort } from '../i18n';
import { i18n } from '../boot/i18n';
import { AppPlatform } from './interface/platform';
import { Cookies, uid, copyToClipboard as quasarCopyToClipboard } from 'quasar';
import {
	ApplicationRequestInterceptor,
	ApplicationResponseInterceptor
} from './interface';
import { notifyRequestMessageError } from 'src/utils/notifyRedefinedUtil';
import { throttle } from 'lodash';

export const APPLICATION_WS_ID = 'application_ws_id';

export class SubApplication implements Application {
	ssiRule?: (() => string) | undefined;
	async copyToClipboard(text: string) {
		return await quasarCopyToClipboard(text);
	}
	applicationName = '';

	platform?: AppPlatform | undefined = undefined;

	async appLoadPrepare(_data: any): Promise<void> {
		// throw new Error('Method not implemented.');
	}
	async appMounted(): Promise<void> {
		// throw new Error('Method not implemented.');
	}
	async appUnMounted(): Promise<void> {
		// throw new Error('Method not implemented.');
	}
	async appRedirectUrl(
		_redirect: any,
		_currentRoute: RouteLocationNormalizedLoaded
	): Promise<void> {
		// throw new Error('Method not implemented.');
	}

	private getWebsocketId() {
		const lastId = localStorage.getItem(APPLICATION_WS_ID);
		if (lastId) {
			return lastId;
		}
		const websocketId = uid();
		localStorage.setItem(APPLICATION_WS_ID, websocketId);
		return websocketId;
	}

	async getWsLoginData() {
		return {
			application: this.applicationName,
			token: Cookies.get('auth_token') || '',
			id: this.getWebsocketId()
		};
	}

	getWsPongRes(data: any) {
		if (typeof data == 'string') {
			return JSON.parse(data).event === 'pong';
		}
		if (typeof data == 'object') {
			return data.event == 'pong';
		}
		return false;
	}

	getWSConnectUrl() {
		if (process.env.WS_URL) {
			return [process.env.WS_URL];
		}
		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';

		const ws_url = `${protocol}//${
			process.env.proxyTarget || window.location.host
		}/ws`;

		return [ws_url];
	}

	websocketConfig = {
		useShareWorker: false,
		shareWorkerName: '',
		externalInfo() {
			return {};
		}
	};

	requestIntercepts: ApplicationRequestInterceptor[] = [];

	responseIntercepts: ApplicationResponseInterceptor[] = [];

	responseErrorInterceps = (error: any) => {
		// throw error;
		throw error;
	};

	tokenInvalidErrorIntercep?: ((error: any) => boolean) | undefined;
	commonResponseIntercepts?: ((response: any) => boolean) | undefined;
	commonRequestIntercepts?: ((config: any) => any) | undefined;

	filesUploadConfig = {
		autoBindResumable: true,
		filesUpdate: (origin_id: number, target: any) => {},
		filesFilter: (files: FileList) => {
			return files as any;
		}
	};

	openUrl(url: string, target?: '_blank' | '_self') {
		if (!target || target === '_blank') {
			window.open(url, '_blank', 'noopener,noreferrer');
		} else {
			window.location.href = url;
		}
	}
}

let normalApplication: Application = new SubApplication();

/**
 * Set the appropriate [[Platform]] implemenation for the current environment
 */
export function setApplication(app: Application) {
	normalApplication = app;
}

/**
 * Get the current [[Platform]] implemenation
 */
export function getApplication() {
	return normalApplication;
}

export class NormalApplication extends SubApplication {
	async appLoadPrepare(_data: any): Promise<void> {
		this.initLanguage();
		this.initAxiosIntercepts();
	}
	async appMounted(): Promise<void> {
		// commonAppMounted(this);
	}
	async appUnMounted(): Promise<void> {
		// commonUnMounted(this);
	}
	async appRedirectUrl(
		_redirect: any,
		_currentRoute: RouteLocationNormalizedLoaded
	): Promise<void> {
		// throw new Error('Method not implemented.');
	}

	initLanguage() {
		let terminusLanguage = '';

		const terminusLanguageInfo = document.querySelector(
			'meta[name="terminus-language"]'
		) as any;

		if (terminusLanguageInfo && terminusLanguageInfo.content) {
			terminusLanguage = terminusLanguageInfo.content;
		} else {
			terminusLanguage = navigator.language || (navigator as any).userLanguage;
		}

		if (terminusLanguage) {
			if (languagesShort[terminusLanguage]) {
				i18n.global.locale.value = languagesShort[terminusLanguage] as any;
			} else if (supportLanguages.find((e) => e.value == terminusLanguage)) {
				i18n.global.locale.value = terminusLanguage as any;
			}
		}
	}

	commonTokenInvalidIntercept() {
		this.tokenInvalidErrorIntercep = (error) => {
			if (
				!error ||
				!error.response ||
				!error.response.status ||
				error.response.status != 459
			) {
				return false;
			}

			notifyRequestMessageError('The token has expired, please log in again');

			this.tokenInvalidRequestError = error;
			setTimeout(() => {
				this.tokenInvalidThrottle();
			}, 2000);

			return true;
		};
	}

	tokenInvalidRequestError: any = undefined;

	tokenInvalidThrottle = throttle(() => {
		if (this.tokenInvalidRequestError) {
			this.redirectToLogin({
				fa2: this.tokenInvalidRequestError.response.data.fa2,
				rm: this.tokenInvalidRequestError.response.data.method,
				rd: window.location.href
			});
		}
	}, 3000);

	initAxiosIntercepts() {
		this.commonTokenInvalidIntercept();
	}

	getBaseAuthServer(protocol = 'https:') {
		let url = protocol + '//';
		const module = 'auth';
		const parts = window.location.hostname.split('.');
		if (parts.length > 1) {
			parts[0] = module;
			const processedHostname = parts.join('.');
			url = url + processedHostname;
		} else {
			url = url + module + window.location.hostname;
		}
		return url;
	}

	redirectToLogin(params: { fa2: boolean; rm: string; rd: string }) {
		const baseUrl = this.getBaseAuthServer();
		const searchParams = new URLSearchParams();

		Object.entries(params).forEach(([key, value]) => {
			if (value !== undefined && value !== null) {
				searchParams.append(key, `${value}`);
			}
		});

		const queryString = searchParams.toString();

		const loginUrl = queryString ? `${baseUrl}?${queryString}` : baseUrl;

		window.location.replace(loginUrl);
	}
}
