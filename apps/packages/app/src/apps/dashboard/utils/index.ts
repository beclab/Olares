import { isString, capitalize, isUndefined, isNull } from 'lodash';
import moment from 'moment-mini';

export const getLocalTime = (time: any) => {
	let formatTime = time;

	if (time && isString(time) && time.indexOf(' +0000 UTC') !== -1) {
		formatTime = time.replace(' +0000 UTC', 'Z').replace(' ', 'T');
	}

	return moment.utc(formatTime).local();
};

export const _capitalize = (value: string | undefined) => {
	return capitalize(value).replace(/_/g, ' ');
};

/**
 * join selector
 * @param {Object} selector
 */
export const joinSelector = (selector = {}) =>
	Object.entries(selector)
		.filter((entry) => !isUndefined(entry[1]) && !isNull(entry[1]))
		.map(([key, value]) => `${key}=${value}`)
		.join(',');
