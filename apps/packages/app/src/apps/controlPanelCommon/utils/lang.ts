export const supportLanguages = ['en-US', 'zh-CN'];

export function getLangWithMeta() {
	let terminusLanguage: string | null = null;
	let locale = navigator.language;
	const terminusLanguageInfo = document.querySelector(
		'meta[name="terminus-language"]'
	);
	if (terminusLanguageInfo && terminusLanguageInfo.getAttribute('content')) {
		terminusLanguage = terminusLanguageInfo.getAttribute('content');
	}

	if (terminusLanguage) {
		if (supportLanguages.find((e) => e == terminusLanguage)) {
			locale = terminusLanguage;
		}
	}

	return locale.includes('zh') ? 'zh-CN' : 'en-US';
}
