import { NormalApplication } from './base';
import { useTokenStore } from './../stores/wizard-step';
import { commonInterceptValue } from 'src/utils/response';

export class WizardApplication extends NormalApplication {
	applicationName = 'wizard';
	async appLoadPrepare(data): Promise<void> {
		await super.appLoadPrepare(data);

		const tokenStore = useTokenStore();
		const host = window.location.origin;
		tokenStore.setUrl(host);
	}

	async appRedirectUrl(): Promise<void> {
		if (document.getElementById('Loading'))
			document.getElementById('Loading')?.remove();
		const tokenStore = useTokenStore();

		return new Promise((resolve) => {
			tokenStore.loadData().then(() => {
				tokenStore.loadWizard().then(() => {
					resolve();
				});
			});
		});
	}

	initAxiosIntercepts(): void {
		this.requestIntercepts.push((config) => {
			const tokenStore = useTokenStore();
			config.headers['Access-Control-Allow-Origin'] = '*';
			config.headers['Access-Control-Allow-Headers'] =
				'X-Requested-With,Content-Type';
			config.headers['Access-Control-Allow-Methods'] =
				'PUT,POST,GET,DELETE,OPTIONS';

			if (tokenStore.token?.access_token) {
				console.log(config.baseURL);
				if (config.baseURL?.startsWith('https://auth.')) {
					console.log('start with auth');
					return config;
				} else {
					config.headers['X-Authorization'] = tokenStore.token?.access_token;
					return config;
				}
			} else {
				return config;
			}
		});

		this.responseIntercepts.push((response) => {
			const data = response.data;

			if (
				!response ||
				(response.status != 200 &&
					response.status != 201 &&
					response.status != 421 &&
					response.status != 304) ||
				!data
			) {
				throw Error('Network error, please try again later');
			}

			if (data.code == 100001) {
				throw Error(data.message);
			}

			if (typeof data == 'string' && commonInterceptValue.includes(data)) {
				throw Error('Default');
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
		this.responseErrorInterceps = (error: any) => {
			console.log('wizard error ====>', error.response);
			if (
				error.response &&
				error.response.status == 421 &&
				typeof error.response.data == 'string' &&
				commonInterceptValue.includes(error.response.data)
			) {
				console.log('default error 1');

				throw Error('Default');
			}
			throw error;
		};
	}
}
