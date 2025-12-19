import { NormalApplication } from './base';
import { useTokenStore } from '../stores/token';
import { CurrentView } from '../utils/constants';
import { useDevice, onDeviceChange } from '@bytetrade/core';

export class LoginApplication extends NormalApplication {
	applicationName = 'login';
	async appLoadPrepare(data): Promise<void> {
		await super.appLoadPrepare(data);

		const tokenStore = useTokenStore();
		const { state } = useDevice();
		tokenStore.deviceInfo = state;

		onDeviceChange((state) => {
			tokenStore.deviceInfo = state;
		});
	}

	async appRedirectUrl(redirect: any): Promise<void> {
		const tokenStore = useTokenStore();
		let host = '';
		if (typeof window !== 'undefined') {
			host = window.location.origin;
		}

		tokenStore.currentView = CurrentView.FIRST_FACTOR;
		tokenStore.setUrl(host);

		const urlParams = new URLSearchParams(window.location.search);
		const logout = urlParams.get('logout');
		const url_redirect = urlParams.get('redirect') || urlParams.get('rd');
		const fa2 = urlParams.get('fa2') || 'false';

		console.log('logout:', logout);
		console.log('url_redirect:', url_redirect);
		console.log('redirect:', redirect);
		console.log('urlParams ===>', urlParams);

		return await tokenStore
			.refresh_token(logout, fa2 == 'true')
			.then(async () => {
				const hostUrl = new URL(host);
				let defaultPath = host;
				if (hostUrl.host.startsWith('test.')) {
					defaultPath =
						'https://' +
						hostUrl.host.split(':')[0].replace('test.', 'desktop.');
				} else {
					defaultPath = host.replace('auth.', 'desktop.');
				}

				const path = url_redirect || defaultPath;
				window.location.replace(path);
			})
			.catch(async () => {
				await tokenStore.loadData().then(async () => {
					if (document.getElementById('Loading'))
						document.getElementById('Loading')?.remove();
					if (fa2 == 'true') {
						tokenStore.currentView = CurrentView.SECOND_FACTOR;
					}
				});
			});
	}

	async appUnMounted() {
		const { cleanup } = useDevice();
		cleanup();
	}

	initAxiosIntercepts(): void {
		this.requestIntercepts.push((config) => {
			config.headers['Access-Control-Allow-Origin'] = '*';
			config.headers['Access-Control-Allow-Headers'] =
				'X-Requested-With,Content-Type';
			config.headers['Access-Control-Allow-Methods'] =
				'PUT,POST,GET,DELETE,OPTIONS';

			return config;
		});

		this.responseIntercepts.push((response, router) => {
			if (response && response.status == 401) {
				router.push({ path: '/login' });
				return;
			}

			if (!response || response.status != 200 || !response.data) {
				throw Error('Network error, please try again later');
			}

			const data = response.data;
			if (data.code == 100001) {
				router.push({ path: '/login' });
				throw Error(data.message);
			}

			if (data.status) {
				if (data.status === 'OK') {
					return data.data;
				}
				throw Error(data.status);
			} else {
				if (data.code != 0) {
					throw Error(data.message);
				}

				return data.data;
			}
		});
	}
}
