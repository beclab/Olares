import { _capitalize } from '../../utils/index';
import { get } from 'lodash';
import { t } from '@apps/dashboard/boot/i18n';

export const MetricTypes = {
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
		title: t('CPU_USAGE'),
		unit: '%',
		legend: [t('USAGE')],
		data: get(data, `${MetricTypes.cpu_utilisation}.data.result`)
	},
	{
		type: 'utilisation',
		title: t('MEMORY_USAGE'),
		unit: '%',
		legend: [t('USAGE')],
		data: get(data, `${MetricTypes.memory_utilisation}.data.result`)
	},
	{
		type: 'load',
		title: t('AVERAGE_CPU_LOAD'),
		legend: [
			t('MINUTE_TIME', { count: 1 }),
			t('MINUTE_TIME', { count: 5 }),
			t('MINUTE_TIME', { count: 15 })
		],
		data: [
			get(data, `${MetricTypes.cpu_load1}.data.result[0]`, {}),
			get(data, `${MetricTypes.cpu_load5}.data.result[0]`, {}),
			get(data, `${MetricTypes.cpu_load15}.data.result[0]`, {})
		]
	},
	{
		type: 'usage',
		title: t('DISK_USAGE'),
		unitType: 'disk',
		legend: [t('USAGE')],
		data: get(data, `${MetricTypes.disk_size_usage}.data.result`)
	},
	{
		type: 'utilisation',
		title: t('INODE_USAGE'),
		unit: '%',
		legend: [t('USAGE')],
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
		title: t('DISK_THROUGHPUT'),
		unitType: 'throughput',
		legend: [t('READ'), t('WRITE')],
		data: [
			get(data, `${MetricTypes.disk_read_throughput}.data.result[0]`, {}),
			get(data, `${MetricTypes.disk_write_throughput}.data.result[0]`, {})
		]
	},
	{
		type: 'iops',
		title: t('IOPS'),
		legend: [t('READ'), t('WRITE')],
		data: [
			get(data, `${MetricTypes.disk_read_iops}.data.result[0]`, {}),
			get(data, `${MetricTypes.disk_write_iops}.data.result[0]`, {})
		]
	},
	{
		type: 'bandwidth',
		title: t('NETWORK_TRAFFIC'),
		unitType: 'bandwidth',
		legend: [t('OUT'), t('IN')],
		data: [
			get(data, `${MetricTypes.net_transmitted}.data.result[0]`, {}),
			get(data, `${MetricTypes.net_received}.data.result[0]`, {})
		]
	},
	{
		type: 'count',
		title: t('POD_STATUS'),
		unit: '',
		legend: [t('RUNNING'), t('COMPLETED'), t('WARNING')],
		data: [
			get(data, `${MetricTypes.pod_running_count}.data.result[0]`, {}),
			get(data, `${MetricTypes.pod_completed_count}.data.result[0]`, {}),
			get(data, `${MetricTypes.pod_abnormal_count}.data.result[0]`, {})
		]
	}
];
