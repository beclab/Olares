import 'axios';

declare module 'axios' {
	export interface AxiosRequestConfig {
		noToast?: boolean;
	}
}
