import { t } from '@apps/control-panel-common/src/boot/i18n';
import { isEmpty, isNil, isNumber, isUndefined } from 'lodash';

export const timeReflection = {
	[10]: {
		label: '10m',
		step: '1m'
	},
	[20]: {
		label: '20m',
		step: '1m'
	},
	[30]: {
		label: '30m',
		step: '1m'
	},
	[60]: {
		label: '1h',
		step: '10m'
	},
	[2 * 60]: {
		label: '2h',
		step: '20m'
	},
	[3 * 60]: {
		label: '3h',
		step: '10m'
	},
	[5 * 60]: {
		label: '5h',
		step: '10m'
	},
	[8 * 60]: {
		label: '8h',
		step: '30m'
	},
	[12 * 60]: {
		label: '12h',
		step: '30m'
	},
	[24 * 60]: {
		label: '1d',
		step: '60m'
	},
	[3 * 24 * 60]: {
		label: '3d',
		step: '60m'
	},
	[7 * 24 * 60]: {
		label: '7d',
		step: '60m'
	}
};

export const timeOption = Object.values(timeReflection).map(
	(item) => item.label
);

export const specialTime = (value) => {
	if (!value) {
		return;
	}
	const minutes = getMinutes(value);
	if (timeReflection[minutes]) {
		return timeReflection[minutes].step;
	}
};

export const getMinutes = (timeStr) => {
	const unit = timeStr.slice(-1);
	const value = parseFloat(timeStr);

	switch (unit) {
		default:
		case 's':
			return value / 60;
		case 'm':
			return value;
		case 'h':
			return value * 60;
		case 'd':
			return value * 24 * 60;
	}
};

export const getStep = (timeStr, times) =>
	`${parseInt(getMinutes(timeStr) / times, 10)}m`;

export const getTimes = (timeStr, step) =>
	Math.floor(getMinutes(timeStr) / getMinutes(step));

export const getTimeStr = (seconds) => {
	let value = Math.round(parseFloat(seconds) / 60);

	if (value < 60) {
		return `${value}m`;
	}

	value = Math.round(value / 60);
	if (value < 24) {
		return `${value}h`;
	}

	return `${Math.round(value / 24)}d`;
};

export const getLastTimeStr = (step, times) => {
	const unit = step.slice(-1);
	const timeStr = `${parseFloat(step) * times}${unit}`;
	const value = getMinutes(timeStr) * 60;
	return getTimeStr(value);
};

export const getTimeLabel = (timeStr) => {
	const unit = timeStr.slice(-1).toUpperCase();
	return t(`LAST_TIME_${unit}`, { count: parseInt(timeStr, 10) });
};

export const getTimeOptions = (times) =>
	times.map((time) => ({
		label: getTimeLabel(time),
		value: time
	}));

export function timeRangeFormate(value, times) {
	let step = getStep(value, times);
	const stepNum = parseInt(step, 10);
	if (specialTime(value)) {
		step = specialTime(value);
		times = getTimes(value, step);
	} else {
		if (stepNum < 1) {
			times = 10;
			step = getStep(value, times);
		}

		if (stepNum > 60) {
			step = '60m';
			times = getTimes(value, step);
		}
	}
	return { step, times, lastTime: value };
}
