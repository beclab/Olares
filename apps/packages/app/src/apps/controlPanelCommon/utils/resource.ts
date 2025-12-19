import { toNumber, get, orderBy } from 'lodash-es';

export const resourceMetricsSort = (
	monitoringData: any,
	list: any,
	type: 'pod' | 'container',
	metric: string
) => {
	const sortBy = `${type}_${metric}`;
	const data = monitoringData[sortBy].data.result;
	const podsLastMetricList = data.map((item: any) => ({
		name: item.metric[type],
		value: toNumber(get(item, `values[${item.values.length - 1}][1]`, 0))
	}));
	const sortData = orderBy(podsLastMetricList, 'value', 'desc');
	let target = undefined;
	const sortPodsList = sortData.map((item) => {
		target = list.find((child: any) => child.name === item.name);
		return target;
	});
	return sortPodsList;
};
