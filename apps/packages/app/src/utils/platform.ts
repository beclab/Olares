import { i18n } from 'src/boot/i18n';

// import { Platform } from 'quasar';
export const isPad = () => {
	// return Platform.is.ipad || isAndroidTablet();
	return false;
};

// const isAndroidTablet = () => {
// 	return (
// 		/android/i.test(navigator.userAgent) && !/mobile/i.test(navigator.userAgent)
// 	);
// };

const services = {
	en: {
		serviceAgreement: 'https://cdn.bttcdn.com/os/en/LarePass-agreement.html',
		privacyPolicy: 'https://cdn.bttcdn.com/os/en/LarePass-privacy.html'
	},
	zh: {
		serviceAgreement:
			'https://cdn.api.jointerminus.cn/os/zh/LarePass-agreement.html',
		privacyPolicy: 'https://cdn.api.jointerminus.cn/os/zh/LarePass-privacy.html'
	}
};

export const appServices = () => {
	const language = i18n.global.locale.value.split('-')[0] || 'en';
	return services[language];
};

export const displayAppServices = () => {
	return (
		process.env.APP_SERVICES != undefined &&
		process.env.APP_SERVICES == 'APP_SERVICES'
	);
};
