import { RouteNames } from 'src/router/route-const';
import { Router } from 'vue-router';
import { browser } from 'src/platform/interface/bex/browser/target';

export const linkToSetting = (name: RouteNames, router: Router) => {
	let fullUrl = '';
	if (process.env.IS_BEX) {
		const extensionId = browser.runtime.id;
		fullUrl = `chrome-extension://${extensionId}/www/options.html#/home`;
	} else {
		fullUrl = `${window.location.origin}/options/account`;
	}
	window.open(fullUrl, '_blank');
};

export const isChromeExtension = () => {
	if (window.location.href.includes('/options/')) {
		return true;
	}
	const pageType = document.querySelector('meta[name="page-type"]') as any;
	const type = pageType?.content || '';
	if (type === 'options') {
		return true;
	}
	return false;
};
