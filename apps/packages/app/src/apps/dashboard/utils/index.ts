import { isString, isUndefined, isNull } from 'lodash';
import moment from 'moment-mini';
export { _capitalize } from '@apps/control-panel-common/src/utils/index';

export const getLocalTime = (time: any) => {
	let formatTime = time;

	if (time && isString(time) && time.indexOf(' +0000 UTC') !== -1) {
		formatTime = time.replace(' +0000 UTC', 'Z').replace(' ', 'T');
	}

	return moment.utc(formatTime).local();
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
