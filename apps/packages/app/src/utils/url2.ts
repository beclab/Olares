import { i18n } from 'src/boot/i18n';

export enum URL_VALID_STATUS {
	BLOCKED = 'BLOCKED',
	PROTOCOL = 'PROTOCOL'
}
export interface UrlValidationResult {
	valid: boolean;
	status?: `${URL_VALID_STATUS}`;
	reason?: string;
}

export function getBlacklistRules(): Array<{ pattern: RegExp }> {
	try {
		const { useBlacklistStore } = require('src/stores/settings/blacklist');
		const blacklistStore = useBlacklistStore();
		return blacklistStore.parsedBlacklistRules || [];
	} catch (e) {
		return [];
	}
}

export async function getBlacklistRulesAsync(): Promise<
	Array<{ pattern: RegExp }>
> {
	try {
		const { useBlacklistStore } = require('src/stores/settings/blacklist');
		const blacklistStore = useBlacklistStore();
		await blacklistStore.ensureInitialized();
		return blacklistStore.parsedBlacklistRules || [];
	} catch (e) {
		console.error('get blacklist error', e);
		return [];
	}
}

export function parseBlacklistData(data) {
	const rules: Array<{
		pattern: RegExp;
	}> = [];
	data.urls.forEach((url: string) => {
		const escapedUrl = url.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
		rules.push({
			pattern: new RegExp(`^${escapedUrl}$`, 'i')
		});
	});

	Object.entries(data.domains).forEach(([domain, regexList]) => {
		if (Array.isArray(regexList) && regexList.length > 0) {
			regexList.forEach((regexStr: string) => {
				const pureRegex = regexStr.replace(/^\/|\/$/g, '');
				rules.push({
					pattern: new RegExp(pureRegex, 'i')
				});
			});
		} else {
			const escapedDomain = domain.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
			const domainRegex = `^(https?:\\/\\/)?([a-zA-Z0-9-]+\\.)*${escapedDomain}\\/?.*`;
			rules.push({
				pattern: new RegExp(domainRegex, 'i')
			});
		}
	});

	return rules;
}

export async function validateUrlWithReasonAsync(
	url: string
): Promise<UrlValidationResult> {
	if (!url || typeof url !== 'string' || url.trim().length === 0) {
		return { valid: false, reason: i18n.global.t('url_value_empty') };
	}

	const trimmedUrl = url.trim();
	const lowerUrl = trimmedUrl.toLowerCase();

	const allowedProtocols = ['youtube:', 'rsshub:', 'http:', 'https:'];
	const hasValidProtocol = allowedProtocols.some((protocol) =>
		lowerUrl.startsWith(protocol)
	);
	if (!hasValidProtocol) {
		return {
			valid: false,
			status: URL_VALID_STATUS.PROTOCOL,
			reason: i18n.global.t('url_unsupported')
		};
	}

	try {
		new URL(trimmedUrl);
	} catch (e) {
		return { valid: false, reason: i18n.global.t('url_malformed_format') };
	}

	const currentBlacklistRules = await getBlacklistRulesAsync();
	console.log('parsedBlacklistRules--2', currentBlacklistRules, lowerUrl);
	for (const rule of currentBlacklistRules) {
		if (rule.pattern.test(lowerUrl)) {
			return {
				valid: false,
				status: URL_VALID_STATUS.BLOCKED,
				reason: i18n.global.t('this page in blocklist')
			};
		}
	}

	return { valid: true };
}

export function replaceOriginDomain(
	origin,
	newDomain = 'file',
	removePort = false
) {
	const urlParts = origin.split('://');
	if (urlParts.length < 2) {
		return origin;
	}

	const protocol = urlParts[0];
	const domainWithPort = urlParts[1];

	let domain, port;
	const portIndex = domainWithPort.indexOf(':');
	if (portIndex !== -1) {
		domain = domainWithPort.substring(0, portIndex);
		port = domainWithPort.substring(portIndex);
	} else {
		domain = domainWithPort;
		port = '';
	}

	const domainParts = domain.split('.');

	if (domainParts.length >= 2) {
		domainParts[0] = newDomain;
	} else if (domainParts.length === 1) {
		domainParts[0] = newDomain;
	}

	const newDomainStr = domainParts.join('.');

	const portStr = removePort ? '' : port;

	return `${protocol}://${newDomainStr}${portStr}`;
}
