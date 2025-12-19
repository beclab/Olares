import { PodItem } from '@apps/dashboard/src/types/network';
import {
	getLastMonitoringData,
	getLastMonitoringDataWithPath
} from '@apps/dashboard/src/utils/monitoring';
import { get, isNumber, round, sortBy, includes } from 'lodash';
import { t } from '@apps/dashboard/boot/i18n';
import { convertTemperature } from '@apps/dashboard/src/utils/cpu';
import { getDiskSize } from '@apps/dashboard/src/utils/disk';

export const columns: any = [
	{
		name: 'device',
		label: t('DISK_OP.FILE_SYSTEM'),
		align: 'left',
		field: 'device'
	},
	{
		name: 'total',
		align: 'left',
		label: t('DISK_OP.TOTAL_CAPACITY'),
		field: 'total'
	},
	{
		name: 'usage',
		align: 'left',
		label: t('DISK_OP.USED_SPACE'),
		field: 'usage'
	},
	{
		name: 'available',
		align: 'left',
		label: t('DISK_OP.AVAILABLE_SPACE'),
		field: 'available'
	},
	{
		name: 'utilisation',
		align: 'left',
		label: t('DISK_OP.USAGE_RATE'),
		field: 'utilisation'
	},
	{
		name: 'mountpoint',
		align: 'right',
		label: t('DISK_OP.MOUNT_POINT'),
		field: 'mountpoint'
	}
];

export const MetricTypes = {
	disk_smartctl_info: 'node_disk_smartctl_info',
	disk_temp_celsius: 'node_disk_temp_celsius',
	one_disk_utilization_ratio: 'node_one_disk_utilization_ratio',
	one_disk_capacity_size: 'node_one_disk_capacity_size',
	one_disk_avail_size: 'node_one_disk_avail_size',
	disk_power_on_hours: 'node_disk_power_on_hours',
	one_disk_data_bytes_read: 'node_one_disk_data_bytes_read',
	one_disk_data_bytes_written: 'node_one_disk_data_bytes_written',
	device_size_usage: 'node_device_size_usage',
	device_partition_size_total: 'node_device_partition_size_total',
	filesystem_size_bytes: 'node_filesystem_size_bytes',
	device_size_utilisation: 'node_device_size_utilisation'
};
type DiskType = 'HDD' | 'SSD';

export type MetricTypesType = typeof MetricTypes;

export const getValue = (data: PodItem) => get(data, 'value[1]', 0);

const getDiskType = (rotational: string) => {
	return rotational === '1' ? t('DISK_OP.HDD') : t('DISK_OP.SSD');
};

export const getDiskOptions = (
	data: { [key: string]: string },
	MetricTypes: MetricTypesType
) => {
	const firstData: any = get(
		data,
		`${MetricTypes.disk_smartctl_info}.data.result`,
		[]
	);

	let lastData: { [key: string]: any } = getLastMonitoringData({});
	const result = firstData.map((item, index) => {
		lastData = getLastMonitoringDataWithPath(data, (metric) => {
			const device =
				get(metric, 'metric.device') || get(metric, 'metric.disk_name');
			return (
				device &&
				device.includes(item.metric.device) &&
				item.metric.node === get(metric, 'metric.node')
			);
		});
		const celsius = getValue(lastData[MetricTypes.disk_temp_celsius]);
		const fahrenheit = convertTemperature(celsius, 'F');
		const rotational = get(item, 'metric.rotational', '0');
		const logical_block_size = get(item, 'metric.logical_block_size', '512');
		const physical_block_size = get(item, 'metric.physical_block_size', '512');

		const LOGICAL_SIZE = '4096';
		const is4K =
			(rotational == '0' && logical_block_size == LOGICAL_SIZE) ||
			(rotational == '1' &&
				logical_block_size == LOGICAL_SIZE &&
				physical_block_size == LOGICAL_SIZE);

		const capacity_size = getValue(
			lastData[MetricTypes.one_disk_capacity_size]
		);
		const capacity_value = getDiskSize(capacity_size);

		const avail_size = getValue(lastData[MetricTypes.one_disk_avail_size]);
		const used_size = capacity_size - avail_size;
		const used_value = getDiskSize(used_size);

		const capacity_all_size = get(item, 'metric.capacity', '0');

		const capacity_all_value = getDiskSize(capacity_all_size);
		const health_ok_status = get(item, 'metric.health_ok') === 'true';
		const headerData = {
			device: get(item, 'metric.device', '-'),
			health_ok: health_ok_status
				? t('DISK_OP.NORMAL')
				: t('DISK_OP.EXCEPTION'),
			health_ok_status,
			disk_size_ratio: used_size / capacity_size,
			used_size: used_value,
			capacity_show: !!capacity_size,
			capacity_size: t('DISK_OP.AVAILABLE', { count: capacity_value }),
			rotational: getDiskType(get(item, 'metric.rotational', '-')),
			name: get(item, 'metric.name', ''),
			node: get(item, 'metric.node', '')
		};

		const disk_power_on_hours = getValue(
			lastData[MetricTypes.disk_power_on_hours]
		);

		const one_disk_data_bytes_written = getValue(
			lastData[MetricTypes.one_disk_data_bytes_written]
		);

		const contentData = [
			{
				name: t('DISK_OP.TOTAL_CAPACITY'),
				value: capacity_all_value
			},
			{
				name: t('DISK_OP.MODEL_TYPE'),
				value: get(item, 'metric.model', '-')
			},
			{
				name: t('DISK_OP.SERIAL_NUMBER'),
				value: get(item, 'metric.serial', '-')
			},
			{
				name: t('DISK_OP.INTERFACE_PROTOCOL'),
				value: get(item, 'metric.protocol', '-')
			},
			{
				name: t('DISK_OP.TEMPERATURE'),
				value:
					celsius && Number(celsius)
						? round(celsius, 1) + '°C' + '/' + round(fahrenheit, 1) + '°F'
						: '-/-'
			},
			{
				name: t('DISK_OP.FIRMWARE_VERSION'),
				value: get(item, 'metric.firmware', '-')
			},
			{
				name: t('DISK_OP.FOUR_K_NATIVE_HDD'),
				value: is4K ? t('DISK_OP.YES_4K') : t('DISK_OP.NO_4K')
			},
			{
				name: t('OWNED_NODE'),
				value: get(item, 'metric.node', '-')
			},
			{
				name: t('DISK_OP.POWER_ON_DURATION'),
				value: isNumber(Number(disk_power_on_hours))
					? t('DISK_OP.POWER_ON_DURATION_VALUE', {
							count: disk_power_on_hours
					  })
					: '-'
			},
			{
				name: t('DISK_OP.WRITE_VOLUME'),
				value: getDiskSize(one_disk_data_bytes_written)
			}
		];

		return { headerData, contentData };
	});

	return result;
};

export const getDiskPartitionRows = (
	data: { [key: string]: string },
	name: string,
	node: string
) => {
	//
	const device_partition_size_total: any = get(
		data,
		`${MetricTypes.device_partition_size_total}.data.result`,
		[]
	);

	const device_size_usage: any = get(
		data,
		`${MetricTypes.device_size_usage}.data.result`,
		[]
	);

	const device_size_utilisation: any = get(
		data,
		`${MetricTypes.device_size_utilisation}.data.result`,
		[]
	);

	const filesystem_size_bytes: any = get(
		data,
		`${MetricTypes.filesystem_size_bytes}.data.result`,
		[]
	);
	const rows = device_partition_size_total
		.filter(
			(item) => item.metric.device.startsWith(name) && item.metric.node === node
		)
		.map((item) => ({
			device: item.metric.device,
			total: getDiskSize(item.value[1]),
			totalSize: item.value[1]
		}));
	const result = rows.map((item) => {
		const utilisationTarget = device_size_utilisation.find(
			(info) => info.metric.device === item.device
		);

		const filesystemBytesTarget = filesystem_size_bytes.filter(
			(info) => info.metric.device === item.device
		);

		const sortedData = sortBy(
			filesystemBytesTarget,
			(obj) =>
				obj.metric.mountpoint.split('/').filter((part) => part !== '').length
		);
		const currentTotalSize = item.totalSize;
		const utilisation = utilisationTarget.value[1];
		const usageSize = round(currentTotalSize * utilisation, 0);

		return {
			...item,
			usage: getDiskSize(usageSize),
			available: getDiskSize(currentTotalSize - usageSize),
			utilisation: round(utilisation * 100, 1) + '%',
			mountpoint: sortedData[0].metric.mountpoint
		};
	});

	return result;
};
