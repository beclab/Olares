import { WebPlugin } from '@capacitor/core';
import { DropboxAuthPluginInterface, DropboxAuthResult } from './definitions';

export class DropboxWeb
	extends WebPlugin
	implements DropboxAuthPluginInterface
{
	resole: any;
	constructor() {
		super();
	}
	initialize(): Promise<void> {
		return new Promise((resolve) => {
			resolve();
		});
	}
	signIn(): Promise<DropboxAuthResult> {
		return new Promise((resolve) => {
			resolve({
				accessToken: '',
				uid: ''
			});
		});
	}
}
