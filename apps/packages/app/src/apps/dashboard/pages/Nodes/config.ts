import { ResourcesResponse } from '@apps/dashboard/src/types/network';
import { get, isArray, isEmpty } from 'lodash';

export const MetricTypes = {
	cpu_utilisation: 'node_cpu_utilisation',
	cpu_load1: 'node_load1',
	cpu_load5: 'node_load5',
	cpu_load15: 'node_load15',
	memory_utilisation: 'node_memory_utilisation',
	disk_utilisation: 'node_disk_size_utilisation',
	disk_inode_utilisation: 'node_disk_inode_utilisation',
	device_size_utilisation: 'node_device_size_utilisation',
	disk_inode_usage: 'node_disk_inode_usage',
	disk_inode_total: 'node_disk_inode_total',
	disk_read_iops: 'node_disk_read_iops',
	disk_write_iops: 'node_disk_write_iops',
	disk_read_throughput: 'node_disk_read_throughput',
	disk_write_throughput: 'node_disk_write_throughput',
	net_transmitted: 'node_net_bytes_transmitted',
	net_received: 'node_net_bytes_received'
};

export function chartConfig(data: ResourcesResponse) {
	const deviceUsage = get(
		data,
		`${MetricTypes.device_size_utilisation}.data.result`,
		[]
	);
	// eslint-disable-next-line @typescript-eslint/ban-ts-comment
	// @ts-ignore
	const legend = deviceUsage && deviceUsage.map((item) => item.metric.device);

	return [
		{
			type: 'utilisation',
			title: 'CPU_USAGE',
			unit: '%',
			legend: ['CPU_USAGE'],
			data: get(data, `${MetricTypes.cpu_utilisation}.data.result`)
		},
		{
			type: 'load',
			title: 'AVERAGE_CPU_LOAD',
			legend: ['1 minute ', '5 minutes', '15 minutes'],
			data: [
				get(data, `${MetricTypes.cpu_load1}.data.result[0]`, {}),
				get(data, `${MetricTypes.cpu_load5}.data.result[0]`, {}),
				get(data, `${MetricTypes.cpu_load15}.data.result[0]`, {})
			]
		},
		{
			type: 'utilisation',
			title: 'MEMORY_USAGE',
			unit: '%',
			legend: ['MEMORY_USAGE'],
			data: get(data, `${MetricTypes.memory_utilisation}.data.result`)
		},
		{
			type: 'utilisation',
			title: 'DISK_USAGE_DETAILS',
			unit: '%',
			legend: ['AVERAGE_USAGE', ...legend],
			data: [
				get(data, `${MetricTypes.disk_utilisation}.data.result[0]`),
				...deviceUsage
			]
		},
		{
			type: 'utilisation',
			title: 'INODE_USAGE',
			unit: '%',
			legend: ['INODE_USAGE'],
			data: get(data, `${MetricTypes.disk_inode_utilisation}.data.result`)
			// renderTooltip: () => {
			//   const usageData = getChartData({
			//     unit: '',
			//     legend: ['USAGE'],
			//     valuesData: [
			//       get(
			//         data,
			//         `${MetricTypes.disk_inode_usage}.data.result[0].values`,
			//         []
			//       ),
			//     ],
			//   });
			//   const totalData = getChartData({
			//     unit: '',
			//     legend: ['USAGE'],
			//     valuesData: [
			//       get(
			//         data,
			//         `${MetricTypes.disk_inode_total}.data.result[0].values`,
			//         []
			//       ),
			//     ],
			//   });

			//   return <CustomTooltip usageData={usageData} totalData={totalData} />;
			// },
		},
		{
			type: 'iops',
			title: 'IOPS',
			legend: ['READ', 'WRITE'],
			data: [
				get(data, `${MetricTypes.disk_read_iops}.data.result[0]`, {}),
				get(data, `${MetricTypes.disk_write_iops}.data.result[0]`, {})
			]
		},
		{
			type: 'throughput',
			title: 'DISK_THROUGHPUT',
			unitType: 'throughput',
			legend: ['READ', 'WRITE'],
			data: [
				get(data, `${MetricTypes.disk_read_throughput}.data.result[0]`, {}),
				get(data, `${MetricTypes.disk_write_throughput}.data.result[0]`, {})
			]
		},
		{
			type: 'bandwidth',
			title: 'NETWORK_TRAFFIC',
			unitType: 'bandwidth',
			legend: ['OUT', 'IN'],
			data: [
				get(data, `${MetricTypes.net_transmitted}.data.result[0]`, {}),
				get(data, `${MetricTypes.net_received}.data.result[0]`, {})
			]
		}
	];
}

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
