/* eslint-disable @typescript-eslint/no-unused-vars */
import { RouteLocationNormalizedLoaded } from 'vue-router';
import { AppPlatform } from './platform';
import {
	ApplicationRequestInterceptor,
	ApplicationResponseInterceptor
} from '.';

export interface Application {
	applicationName: string;

	platform?: AppPlatform;

	/******** app start *************/
	/**
	 * app.vue start
	 * @param data params
	 */
	appLoadPrepare(data: any): Promise<void>;

	/**
	 * app.vue mounted
	 */
	appMounted(): Promise<void>;

	/**
	 * app.vue unmounted
	 */
	appUnMounted(): Promise<void>;

	/**
	 * app.vue
	 * @param redirect
	 */
	appRedirectUrl(
		redirect: any,
		currentRoute: RouteLocationNormalizedLoaded
	): Promise<void>;

	/**
	 * auth token
	 */
	getWsLoginData(): Promise<any>;

	getWsPongRes(data: any): boolean;

	getWSConnectUrl(): string[];

	websocketConfig: {
		useShareWorker: boolean;
		shareWorkerName: string;

		externalInfo(): any;

		responseShareWorkerMessage?: (data: any) => void;
	};

	requestIntercepts: ApplicationRequestInterceptor[];

	responseIntercepts: ApplicationResponseInterceptor[];

	responseErrorInterceps?: (error: any) => any;

	tokenInvalidErrorIntercep?: (error: any) => boolean;

	commonResponseIntercepts?: (response: any) => boolean;

	commonRequestIntercepts?: (config: any) => any;

	ssiRule?: () => string;

	filesUploadConfig?: {
		autoBindResumable: boolean;
		filesUpdate?: (origin_id: number, target: any) => void;
		filesFilter?: (files: FileList) => any;
		toastDeleteFiles?: (files: File[]) => void;
	};

	copyToClipboard(text: string): Promise<void>;

	openUrl(url: string, target?: '_blank' | '_self'): void;
}
