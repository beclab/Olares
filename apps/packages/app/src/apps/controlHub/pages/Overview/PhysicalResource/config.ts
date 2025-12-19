import { _capitalize } from '../../../utils/index';
import { get } from 'lodash';
const MetricTypes = {
	cpu_utilisation: 'cluster_cpu_utilisation',
	memory_utilisation: 'cluster_memory_utilisation',
	cpu_load1: 'cluster_load1',
	cpu_load5: 'cluster_load5',
	cpu_load15: 'cluster_load15',
	disk_size_usage: 'cluster_disk_size_usage',
	disk_inode_utilisation: 'cluster_disk_inode_utilisation',
	disk_inode_usage: 'cluster_disk_inode_usage',
	disk_inode_total: 'cluster_disk_inode_total',
	disk_read_iops: 'cluster_disk_read_iops',
	disk_write_iops: 'cluster_disk_write_iops',
	disk_read_throughput: 'cluster_disk_read_throughput',
	disk_write_throughput: 'cluster_disk_write_throughput',
	net_transmitted: 'cluster_net_bytes_transmitted',
	net_received: 'cluster_net_bytes_received',
	pod_running_count: 'cluster_pod_running_count',
	pod_abnormal_count: 'cluster_pod_abnormal_count',
	pod_completed_count: 'cluster_pod_succeeded_count'
};

export const getMonitoringCfgs = (data: { [key: string]: any }) => [
	{
		type: 'utilisation',
		title: 'CPU_USAGE',
		unit: '%',
		legend: ['USAGE'],
		data: get(data, `${MetricTypes.cpu_utilisation}.data.result`)
	},
	{
		type: 'utilisation',
		title: 'MEMORY_USAGE',
		unit: '%',
		legend: ['USAGE'],
		data: get(data, `${MetricTypes.memory_utilisation}.data.result`)
	},
	{
		type: 'load',
		title: 'AVERAGE_CPU_LOAD',
		legend: ['1 minute', '5 minutes', '15 minutes'],
		data: [
			get(data, `${MetricTypes.cpu_load1}.data.result[0]`, {}),
			get(data, `${MetricTypes.cpu_load5}.data.result[0]`, {}),
			get(data, `${MetricTypes.cpu_load15}.data.result[0]`, {})
		]
	},
	{
		type: 'usage',
		title: 'DISK_USAGE',
		unitType: 'disk',
		legend: ['USAGE'],
		data: get(data, `${MetricTypes.disk_size_usage}.data.result`)
	},
	{
		type: 'utilisation',
		title: 'INODE_USAGE',
		unit: '%',
		legend: ['USAGE'],
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
		type: 'iops',
		title: 'IOPS',
		legend: ['READ', 'WRITE'],
		data: [
			get(data, `${MetricTypes.disk_read_iops}.data.result[0]`, {}),
			get(data, `${MetricTypes.disk_write_iops}.data.result[0]`, {})
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
	},
	{
		type: 'count',
		title: 'POD_STATUS',
		unit: '',
		legend: ['RUNNING', 'COMPLETED', 'WARNING'],
		data: [
			get(data, `${MetricTypes.pod_running_count}.data.result[0]`, {}),
			get(data, `${MetricTypes.pod_completed_count}.data.result[0]`, {}),
			get(data, `${MetricTypes.pod_abnormal_count}.data.result[0]`, {})
		]
	}
];
