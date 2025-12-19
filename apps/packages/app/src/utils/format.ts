import { formatDistanceToNow } from 'date-fns';
import { zhCN, enUS } from 'date-fns/locale';
import { i18n } from 'src/boot/i18n';

const formatFileSize = (
	bytes: number,
	fixed = 2,
	space = '',
	isUnit?: boolean
) => {
	const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
	isUnit = isUnit ?? true;

	if (bytes === 0) {
		return isUnit ? '0B' : '0';
	}

	// console.log('bytes ===>', bytes);

	const i = Math.floor(Math.log(bytes) / Math.log(1024));
	if (i === 0) {
		return bytes.toFixed(0) + (isUnit ? sizes[i] : '');
	}
	return (bytes / 1024 ** i).toFixed(fixed) + space + (isUnit ? sizes[i] : '');
};
export const format = {
	humanStorageSize: (bytes: number) => formatFileSize(bytes),
	formatFileSize
};

export function formatDateFromNow(
	date: Date | string | number,
	addSuffix = true
) {
	try {
		const locale = i18n.global.locale.value === 'zh-CN' ? zhCN : enUS;
		return formatDistanceToNow(new Date(date), { addSuffix, locale: locale });
	} catch (error) {
		return '';
	}
}

const hp = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';

export function generatePasword(length: number, range = hp) {
	let password = '';
	for (let i = 0; i < length; ++i) {
		const index = Math.floor(Math.random() * range.length);
		password = password + range[index];
	}
	return password;
}

export const stringToIntHash = (
	str: string,
	lowerbound = 0,
	upperbound = 32
) => {
	if (!str) {
		return lowerbound;
	}

	let result = 0;
	for (let i = 0; i < str.length; i++) {
		result = result + str.charCodeAt(i);
	}

	if (!lowerbound) lowerbound = 0;
	if (!upperbound) upperbound = 500;

	return (result % (upperbound - lowerbound)) + lowerbound;
};
