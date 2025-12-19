import { boot } from 'quasar/wrappers';
import axios, {
	AxiosInstance,
	AxiosRequestConfig,
	AxiosResponse,
	InternalAxiosRequestConfig
} from 'axios';
import { getApplication } from 'src/application/base';

declare module '@vue/runtime-core' {
	interface ComponentCustomProperties {
		$axios: AxiosInstance;
	}
}

const api = axios.create({
	withCredentials: true,
	maxRedirects: 0
});

export default boot(({ app, router }) => {
	app.config.globalProperties.$axios = axios;
	app.config.globalProperties.$api = api;

	app.config.globalProperties.$axios.interceptors.request.use(
		async (config: InternalAxiosRequestConfig) => {
			const application = getApplication();
			if (application.requestIntercepts) {
				for (const interceptor of application.requestIntercepts) {
					config = (await interceptor(config)) || config;
				}
			}
			return config;
		}
	);

	app.config.globalProperties.$axios.interceptors.response.use(
		async (response: AxiosResponse) => {
			const application = getApplication();
			if (application.responseIntercepts) {
				for (const interceptor of application.responseIntercepts) {
					response = (await interceptor(response, router)) || response;
				}
			}
			return response;
		},
		(error: any) => {
			const application = getApplication();
			if (
				application.tokenInvalidErrorIntercep &&
				application.tokenInvalidErrorIntercep(error)
			) {
				return;
			}
			if (application.responseErrorInterceps) {
				application.responseErrorInterceps(error);
			} else {
				throw error;
			}
		}
	);
});

export { api };
