import { HttpOptions, HttpResponse } from '@capacitor/core';
import { AxiosResponse, InternalAxiosRequestConfig } from 'axios';
import { Router } from 'vue-router';
export interface TabbarItem {
	name: string;
	identify: any;
	normalImage: string;
	activeImage: string;
	darkActiveImage?: string;
	to: string;
	tabChanged?: () => boolean;
	badge?: string;
	hoverImage?: string;
}

export interface HookCapacitorHttpPlugin {
	request(options: HttpOptions): Promise<HttpResponse>;
	get(options: HttpOptions): Promise<HttpResponse>;
	post(options: HttpOptions): Promise<HttpResponse>;
	put(options: HttpOptions): Promise<HttpResponse>;
	patch(options: HttpOptions): Promise<HttpResponse>;
	delete(options: HttpOptions): Promise<HttpResponse>;
}

export type ApplicationRequestInterceptor = (
	config: InternalAxiosRequestConfig
) => InternalAxiosRequestConfig | Promise<InternalAxiosRequestConfig>;

export type ApplicationResponseInterceptor = (
	response: AxiosResponse,
	router: Router
) => AxiosResponse | Promise<AxiosResponse>;
