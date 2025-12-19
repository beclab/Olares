import { get, isEmpty, last, set, isArray, flatten, isUndefined } from 'lodash';
import { getLocalTime, _capitalize } from '.';
import { getValueByUnit, getSuitableUnit } from '@bytetrade/core';
export {
	getValueByUnit,
	getZeroValues,
	getSuitableUnit,
	getSuitableValue,
	UnitKey
} from '@bytetrade/core';

export const getFormatTime = (ms: string, showDay: boolean) =>
	getLocalTime(Number(ms))
		.format(showDay ? 'MM-DD HH:mm' : 'HH:mm:ss')
		.replace(/(\d+:\d+)(:00)$/g, '$1');

export const getLastMonitoringData = (data: any, index = 0) => {
	const result = {};

	Object.entries(data).forEach(([key, value]) => {
		const values = get(value, `data.result[${index}].values`, []) || [];
		const _value = isEmpty(values)
			? get(value, `data.result[${index}].value`, []) || []
			: last(values);
		set(result, `[${key}].value`, _value);
	});

	return result;
};

export const getLastMonitoringDataWithPath = (data: any, callback) => {
	const result = {};

	Object.entries(data).forEach(([key, value]) => {
		const list = get(value, `data.result`, []);
		const target = list.find((item) => callback(item));
		if (target) {
			const values = get(target, 'values', []);
			const _value = isEmpty(values)
				? get(target, `value`, []) || []
				: last(values);
			set(result, `[${key}].value`, _value);
		}
	});

	return result;
};

type ObjTypes = { [key: string]: any };
export const getChartData = ({
	type,
	unit,
	xKey = 'name',
	legend = [],
	valuesData = [],
	dot = 2
}: ObjTypes) => {
	/*
    build a value map => { 1566289260: {...} }
    e.g. { 1566289260: { 'utilisation': 30.2 } }
  */
	let minX = 0;
	let maxX = 0;
	const valueMap: { [key: string]: any } = {};
	const tempData = valuesData.map((values: any, index: number) => {
		return values.map((item: any) => {
			const time = parseInt(get(item, [0], 0), 10);
			const value = get(item, [1]);

			if (!minX || minX > time) minX = time;
			if (!maxX || maxX < time) maxX = time;
			const newValue =
				value === '-1'
					? null
					: getValueByUnit(value, isUndefined(unit) ? type : unit, dot);
			return { name: time, value: newValue };
		});
	});

	const showDay = maxX - minX > 3600 * 24;
	const formatter = (key: any) =>
		xKey === 'name' ? getFormatTime((key * 1000).toString(), showDay) : key;

	const chartData = tempData.map((item: any) =>
		item.map((child: any) => [formatter(child.name), child.value])
	);
	return chartData;
};

export const getAreaChartOps = ({
	type,
	title,
	unitType,
	xKey = 'name',
	legend = [],
	data = [],
	...rest
}: {
	[key: string]: any;
}) => {
	const seriesData = isArray(data) ? data : [];
	const valuesData = seriesData.map((result) => get(result, 'values') || []);
	const unit = unitType
		? getSuitableUnit(flatten(valuesData), unitType)
		: rest.unit;

	const chartData = getChartData({
		type,
		unit,
		xKey,
		legend,
		valuesData,
		dot: rest.dot
	});

	return {
		...rest,
		title,
		unit,
		legend,
		data: chartData
	};
};

export function getResult(result: any) {
	const data: any = {};
	const results = isArray(result) ? result : get(result, 'results', []) || [];

	if (isEmpty(results)) {
		const metricName = get(result, 'metric_name');

		if (metricName) {
			data[metricName] = result;
		}
	} else {
		results.forEach((item: any) => {
			data[item.metric_name] = item;
		});
	}

	return data;
}
