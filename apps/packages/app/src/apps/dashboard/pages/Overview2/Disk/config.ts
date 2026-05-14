import { PodItem } from '@apps/dashboard/src/types/network';
import {
	getLastMonitoringData,
	getLastMonitoringDataWithPath,
	getSuitableValue,
	getValueByUnit
} from '@apps/dashboard/src/utils/monitoring';
import { get, isNumber, round, sortBy, includes } from 'lodash';
import { t } from 'src/boot/dashboard-i18n';
import { convertTemperature } from '@apps/dashboard/src/utils/cpu';
import { getDiskSize } from '@apps/dashboard/src/utils/disk';

export type LsblkMetricRow = {
	name: string;
	node: string;
	pkname?: string;
	size?: string;
	fstype?: string;
	mountpoint?: string;
	fsused?: string;
	fsuse_percent?: string;
};

export type LsblkFlatRow = LsblkMetricRow & {
	__depth: number;
	__key: string;
	__treePrefix: string;
};

export const displayLsblkCell = (v: unknown): string => {
	if (v === null || v === undefined) return '-';
	const s = String(v).trim();
	return s === '' ? '-' : s;
};

export const LSBLK_DISK_FIXED_UNITS = [
	'Ti',
	'Gi',
	'Mi',
	'Ki',
	'Bytes'
] as const;

const capitalizeUnitSuffix = (unit: string) => {
	const c = unit.charAt(0).toUpperCase() + unit.slice(1).toLowerCase();
	return c.replace(/_/g, ' ');
};

export const formatLsblkDiskValue = (
	raw: unknown,
	unitMode: 'auto' | (typeof LSBLK_DISK_FIXED_UNITS)[number]
): string => {
	if (raw === null || raw === undefined) return '-';
	const s = String(raw).trim();
	if (s === '' || s === '-') return '-';
	if (Number.isNaN(Number(s))) return '-';

	if (unitMode === 'auto') {
		return String(getSuitableValue(s, 'disk'));
	}

	const count = getValueByUnit(s, unitMode);
	const unitText = ` ${capitalizeUnitSuffix(unitMode)}`;
	return `${count}${unitText}`;
};

export const getLsblkColumns = () => [
	{
		name: 'name',
		label: t('DISK_OP.LSBLK_NAME'),
		align: 'left',
		field: 'name'
	},
	{
		name: 'size',
		align: 'left',
		label: t('DISK_OP.LSBLK_SIZE'),
		field: (row: LsblkFlatRow) => displayLsblkCell(row.size)
	},
	{
		name: 'fstype',
		align: 'left',
		label: t('DISK_OP.LSBLK_FSTYPE'),
		field: (row: LsblkFlatRow) => displayLsblkCell(row.fstype)
	},
	{
		name: 'mountpoint',
		align: 'left',
		label: t('DISK_OP.MOUNT_POINT'),
		field: (row: LsblkFlatRow) => displayLsblkCell(row.mountpoint)
	},
	{
		name: 'fsused',
		align: 'left',
		label: t('DISK_OP.LSBLK_FSUSED'),
		field: (row: LsblkFlatRow) => displayLsblkCell(row.fsused)
	},
	{
		name: 'fsuse_percent',
		align: 'right',
		label: t('DISK_OP.LSBLK_FSUSE_PERCENT'),
		field: (row: LsblkFlatRow) => displayLsblkCell(row.fsuse_percent)
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
	device_size_utilisation: 'node_device_size_utilisation',
	disk_lsblk_info: 'node_disk_lsblk_info'
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

const metricFromSample = (m: Record<string, string>): LsblkMetricRow => ({
	name: m.name ?? '',
	node: m.node ?? '',
	pkname: m.pkname,
	size: m.size,
	fstype: m.fstype,
	mountpoint: m.mountpoint,
	fsused: m.fsused,
	fsuse_percent: m.fsuse_percent
});

const hasPknameLabels = (rows: LsblkMetricRow[]) =>
	rows.some((r) => {
		const p = r.pkname?.trim();
		return !!p;
	});

const collectSubtreeByPkname = (
	allRows: LsblkMetricRow[],
	rootName: string
): LsblkMetricRow[] => {
	const byName = new Map(allRows.map((r) => [r.name, r]));
	const seen = new Set<string>();
	const queue: string[] = [];

	if (byName.has(rootName)) {
		seen.add(rootName);
		queue.push(rootName);
	}

	while (queue.length) {
		const n = queue.shift()!;
		for (const r of allRows) {
			const pk = (r.pkname || '').trim();
			if (pk === n && !seen.has(r.name)) {
				seen.add(r.name);
				queue.push(r.name);
			}
		}
	}

	if (seen.size === 0) {
		const addDesc = (parent: string) => {
			for (const r of allRows) {
				const pk = (r.pkname || '').trim();
				if (pk === parent && !seen.has(r.name)) {
					seen.add(r.name);
					addDesc(r.name);
				}
			}
		};
		addDesc(rootName);
	}

	return allRows.filter((r) => seen.has(r.name));
};

const resolveParent = (
	r: LsblkMetricRow,
	rootName: string,
	nameSet: Set<string>
): string | undefined => {
	if (r.name === rootName) return undefined;
	const pk = r.pkname?.trim();
	if (pk && nameSet.has(pk)) return pk;

	const prefixes = [...nameSet].filter(
		(n) => n.length > 0 && n !== r.name && r.name.startsWith(n)
	);
	prefixes.sort((a, b) => b.length - a.length);
	const best = prefixes[0];
	if (best) return best;
	if (nameSet.has(rootName)) return rootName;
	return undefined;
};

const buildLsblkTreePrefix = (depth: number, lastStack: boolean[]): string => {
	if (depth === 0) return '';
	let s = '';
	for (let i = 0; i < depth - 1; i++) {
		s += lastStack[i] ? '    ' : '│   ';
	}
	s += lastStack[depth - 1] ? '└── ' : '├── ';
	return s;
};

const flattenLsblkHierarchy = (
	rows: LsblkMetricRow[],
	rootName: string,
	nodeId: string
): LsblkFlatRow[] => {
	const nameSet = new Set(rows.map((r) => r.name));
	const byName = new Map(rows.map((r) => [r.name, r]));

	if (!nameSet.has(rootName)) {
		return rows
			.slice()
			.sort((a, b) => a.name.localeCompare(b.name))
			.map((r) => ({
				...r,
				__depth: 0,
				__treePrefix: '',
				__key: `${nodeId}::${r.name}`
			}));
	}

	const children = new Map<string, LsblkMetricRow[]>();
	for (const r of rows) {
		if (r.name === rootName) continue;
		let p = resolveParent(r, rootName, nameSet);
		if (!p || !nameSet.has(p)) p = rootName;
		if (!children.has(p)) children.set(p, []);
		children.get(p)!.push(r);
	}
	for (const [, list] of children) {
		list.sort((a, b) => a.name.localeCompare(b.name));
	}

	const out: LsblkFlatRow[] = [];
	const walk = (rname: string, depth: number, lastStack: boolean[]) => {
		const r = byName.get(rname);
		if (!r) return;
		const treePrefix =
			depth === 0 ? '' : buildLsblkTreePrefix(depth, lastStack);
		out.push({
			...r,
			__depth: depth,
			__treePrefix: treePrefix,
			__key: `${nodeId}::${r.name}`
		});
		const ch = children.get(rname);
		if (ch) {
			ch.forEach((c, idx) => {
				const isLast = idx === ch.length - 1;
				walk(c.name, depth + 1, [...lastStack, isLast]);
			});
		}
	};
	walk(rootName, 0, []);
	return out;
};

export const getDiskPartitionRows = (
	data: { [key: string]: string },
	name: string,
	node: string
): LsblkFlatRow[] => {
	const disk_lsblk_info: any = get(
		data,
		`${MetricTypes.disk_lsblk_info}.data.result`,
		[]
	);

	const allForNode: LsblkMetricRow[] = disk_lsblk_info
		.filter(
			(item: { metric: Record<string, string> }) => item.metric.node === node
		)
		.map((item: { metric: Record<string, string> }) =>
			metricFromSample(item.metric)
		);

	let subset: LsblkMetricRow[];
	if (hasPknameLabels(allForNode)) {
		subset = collectSubtreeByPkname(allForNode, name);
	} else {
		subset = allForNode.filter(
			(r) => r.name === name || r.name.startsWith(name)
		);
	}

	return flattenLsblkHierarchy(subset, name, node);
};
