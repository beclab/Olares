import { TermiPassDeviceInfo } from '@bytetrade/core';

const browserInfo = (async () => {
	const { default: UAParser } = await import(
		/* webpackChunkName: "ua-parser" */ 'ua-parser-js'
	);
	return new UAParser(navigator.userAgent).getResult();
})();

export class WebPlatform {
	async getDeviceInfo() {
		const { os, browser } = await browserInfo;
		const platform = (os.name && os.name.replace(' ', '')) || '';
		return new TermiPassDeviceInfo({
			platform,
			osVersion: (os.version && os.version.replace(' ', '')) || '',
			id: '',
			manufacturer: '',
			model: '',
			browser: browser.name || '',
			browserVersion: browser.version,
			userAgent: navigator.userAgent,
			locale: navigator.language || 'en',
			description:
				browser.name && browser.name !== 'Electron'
					? browser.name + ' on ' + platform
					: platform + ' Device',
			runtime: 'web'
		});
	}
}
