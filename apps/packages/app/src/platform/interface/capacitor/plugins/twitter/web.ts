import { WebPlugin } from '@capacitor/core';
import { TwitterLoginPlugin } from './definitions';
// import { Dialog } from 'quasar';

export class TwitterLoginWeb extends WebPlugin implements TwitterLoginPlugin {
	resole: any;
	constructor() {
		super();
	}

	// eslint-disable-next-line @typescript-eslint/no-unused-vars
	async login(_options: {
		oauthUrl: string;
	}): Promise<{ status: boolean; message: string }> {
		return {
			status: false,
			message: ''
		};
	}
}
