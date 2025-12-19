import { isEmpty, isNumber, round } from 'lodash';
import { date } from 'quasar';

export const chartEntervalOfWidth = (width: number) => {
	let chartInterval: number | 'auto' = 2;
	if (width > 1500) {
		chartInterval = 0;
	} else if (width > 700) {
		chartInterval = 1;
	} else if (width > 550) {
		chartInterval = 2;
	} else if (width > 450) {
		chartInterval = 3;
	} else {
		chartInterval = 'auto';
	}
};

export function formatter(params: any, unit: string) {
	let dom = '';
	let domItem = '';
	params.forEach((item: any) => {
		domItem = `<div class="chart-tooltip-item-name">
			<span>
				${item.marker}
				<span class="chart-tooltip-item-title">${item.seriesName}</span>
			</span>
			<span class="chart-tooltip-unit">${item.data} ${item.unit || unit}</span>
		</div>`;
		dom += domItem;
	});
	return `<div class="chart-tooltip-title">${params[0].axisValueLabel}</div><div class="chart-tooltip-item-container">${dom}</div>`;
}

export function dateFormate(
	value: string | number,
	formatStr = 'YYYY-MM-DD HH:mm:ss'
) {
	if (typeof value === 'number' || /^\d+$/.test(value.toString())) {
		try {
			return date.formatDate(new Date(Number(value)), formatStr);
		} catch (error) {
			console.warn(`Invalid timestamp: ${value}`);
			return value.toString();
		}
	}

	if (/^\d{1,2}(:\d{1,2}){1,2}$/.test(value)) {
		const today = new Date();
		const datePrefix = date.formatDate(today, 'YYYY-MM-DD');
		const timeParts = value.split(':');
		if (timeParts.length === 2) {
			value = `${datePrefix} ${value}:00`;
		} else {
			value = `${datePrefix} ${value}`;
		}
	}

	try {
		return date.formatDate(value, formatStr);
	} catch (error) {
		console.warn(`Invalid date format: ${value}`);
		return value;
	}
}

export type DataInter = Array<[string, number]>;
export function labelRoundValue(data: DataInter) {
	const values = data.map((item) => item[1]);
	console.log('labelRoundValue', values);
	const max = Math.max(...values);
	if (max > 10) {
		return values.map((item) => round(item, 0));
	} else {
		return values;
	}
}

const numWidth = 6.75;
const dotWidth = 3.17;
const marginLeft = 12;
const minwidth = Math.ceil(2 * numWidth + dotWidth);
export function numWidthCom(numStr: string | number) {
	if (!isNumber(numStr)) {
		return minwidth;
	} else {
		const str = numStr.toString();
		const numLength = str.replace(/\./g, '').length;
		const dotLength = str.length - numLength;
		const strWidth = Math.ceil(numLength * numWidth + dotLength * dotWidth);
		const maxLength = Math.max(minwidth, strWidth);
		return maxLength + marginLeft;
	}
}

export function labelValueWidths(data: DataInter) {
	const values = labelRoundValue(data);
	const test = values.map((item) => numWidthCom(item));
	console.log('test', test);
	return test;
}

export function labelMaxWidth(data: DataInter) {
	const values = labelValueWidths(data);
	if (isEmpty(values)) {
		return minwidth;
	}
	return Math.max(...values);
}
