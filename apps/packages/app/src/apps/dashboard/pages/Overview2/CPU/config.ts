import { PodItem } from '@apps/dashboard/src/types/network';
import {
	getAreaChartOps,
	getLastMonitoringData,
	getValueByUnit
} from '@apps/dashboard/src/utils/monitoring';
import { get, round } from 'lodash';
import { t } from 'src/boot/dashboard-i18n';
import { formatFrequency } from '@apps/dashboard/src/utils/cpu';
import { timeRangeFormate } from '@apps/control-panel-common/src/containers/Monitoring/utils';

export const MetricTypes = {
	cpu_base_frequency_hertz_max: 'node_cpu_base_frequency_hertz_max',
	cpu_frequency_max_hertz: 'node_cpu_frequency_max_hertz',
	cpu_temp_celsius: 'node_cpu_temp_celsius',
	acpitz_temp_celsius: 'node_acpitz_temp_celsius',
	user_cpu_usage: 'node_user_cpu_usage',
	system_cpu_usage: 'node_system_cpu_usage',
	iowait_cpu_usage: 'node_iowait_cpu_usage',
	load1: 'node_load1',
	load5: 'node_load5',
	load15: 'node_load15',
	cpu_info: 'node_cpu_info',
	cpu_dmidecode_info: 'node_cpu_dmidecode_info',
	cpu_total: 'node_cpu_total',
	cpu_utilisation: 'node_cpu_utilisation'
};
export type MetricTypesType = typeof MetricTypes;

interface CpuInfoMetric {
	model_name?: string;
	version?: string;
}

export const getValue = (data: PodItem) => get(data, 'value[1]', 0);

export const hasValue = (data: PodItem) => get(data, 'value[1]') !== undefined;

const usagePerCore = (value, core) => round(core ? value / core : value, 2);

export const getCpuList = (
	data: { [key: string]: string },
	MetricTypes: MetricTypesType
) => {
	const firstData: any = get(data, `${MetricTypes.load1}.data.result`, []);
	return firstData.map((item, index) => {
		const cpuBase = getCpuBaseOptions(data, MetricTypes, index);
		const cpuChartDataTemp = getAreaChartOps(getMonitoringCfgs(data, index));
		const cpuChartData = {
			...cpuChartDataTemp,
			data: cpuChartDataTemp.data.map((chartData) =>
				chartData.map((child) => [
					child[0],
					usagePerCore(child[1], cpuBase.core)
				])
			)
		};

		const usageTateListTemp = getCpuRateOptions(data, MetricTypes, index);
		const usageTateList = usageTateListTemp.map((usageInfo) => ({
			...usageInfo
		}));
		return {
			cpuChartData,
			usageTateList,
			AverageLoad: getCpuAverageLoadOptions(data, MetricTypes, index),
			temperature: getCpuTemperatureOptions(data, MetricTypes, index),
			cpuBase
		};
	});
};

export const getMonitoringCfgs = (
	data: { [key: string]: string },
	index = 0
) => ({
	type: 'load',
	unit: '%',
	title: get(
		data,
		`${MetricTypes.user_cpu_usage}.data.result[${index}].metric.node`,
		undefined
	),
	legend: [t('CPU_OP.UTILIZATION_RATE')],
	data: [get(data, `${MetricTypes.user_cpu_usage}.data.result[${index}]`, {})]
});

export const getCpuRateOptions = (
	data: { [key: string]: string },
	MetricTypes: MetricTypesType,
	index = 0
) => {
	const lastData: { [key: string]: any } = getLastMonitoringData(data, index);
	const user_cpu_usage_value = getValue(lastData[MetricTypes.user_cpu_usage]);
	const system_cpu_usage_value = getValue(
		lastData[MetricTypes.system_cpu_usage]
	);
	const iowait_cpu_usage_value = getValue(
		lastData[MetricTypes.iowait_cpu_usage]
	);

	const cpu_utilisation_value = getValue(lastData[MetricTypes.cpu_utilisation]);
	const total = getValue(lastData[MetricTypes.cpu_total]);

	const result = [
		{
			title: t('CPU_OP.USER'),
			unit: '%',
			data: [
				get(data, `${MetricTypes.user_cpu_usage}.data.result[${index}]`, {})
			],
			value: round(Number(cpu_utilisation_value) * 100, 2)
		},
		{
			title: t('CPU_OP.USER'),
			unit: '%',
			data: [
				get(data, `${MetricTypes.user_cpu_usage}.data.result[${index}]`, {})
			],
			value: round((user_cpu_usage_value / total) * 100, 2),
			info: t('CPU_OP.USER_INFO')
		},
		{
			title: t('CPU_OP.SYSTEM'),
			unit: '%',
			data: [
				get(data, `${MetricTypes.system_cpu_usage}.data.result[${index}]`, {})
			],
			value: round((system_cpu_usage_value / total) * 100, 2),
			info: t('CPU_OP.SYSTEM_INFO')
		},
		{
			title: t('CPU_OP.IO_WAIT'),
			unit: '%',
			data: [
				get(data, `${MetricTypes.iowait_cpu_usage}.data.result[${index}]`, {})
			],
			value: round((iowait_cpu_usage_value / total) * 100, 2),
			info: t('CPU_OP.IO_WAIT_INFO')
		}
	];

	return result;
};

export const getCpuTemperatureOptions = (
	data: { [key: string]: string },
	MetricTypes: MetricTypesType,
	index = 0
) => {
	const lastData: { [key: string]: any } = getLastMonitoringData(data, index);
	const tempData = hasValue(lastData[MetricTypes.cpu_temp_celsius])
		? lastData[MetricTypes.cpu_temp_celsius]
		: lastData[MetricTypes.acpitz_temp_celsius];
	const temperature = getValue(tempData);
	const result = {
		name: t('CPU_OP.CPU_TEMPERATURE'),
		unit: '°C',
		value: round(temperature, 1)
	};

	return result;
};

export const getCpuAverageLoadOptions = (
	data: { [key: string]: string },
	MetricTypes: MetricTypesType,
	index = 0
) => {
	const lastData: { [key: string]: any } = getLastMonitoringData(data, index);
	const result = [
		{
			unit: t('CPU_OP.MINUTES', 1),
			value: round(getValue(lastData[MetricTypes.load1]), 2)
		},
		{
			unit: t('CPU_OP.MINUTES', 5),
			value: round(getValue(lastData[MetricTypes.load5]), 2)
		},
		{
			unit: t('CPU_OP.MINUTES', 15),
			value: round(getValue(lastData[MetricTypes.load15]), 2)
		}
	];

	return result;
};

export const getCpuBaseOptions = (
	data: { [key: string]: string },
	MetricTypes: MetricTypesType,
	index = 0
) => {
	const lastData: { [key: string]: any } = getLastMonitoringData(data, index);
	const frequencyData = hasValue(
		lastData[MetricTypes.cpu_base_frequency_hertz_max]
	)
		? lastData[MetricTypes.cpu_base_frequency_hertz_max]
		: lastData[MetricTypes.cpu_frequency_max_hertz];
	const value = formatFrequency(getValue(frequencyData));

	const cpuInfo = get(
		data,
		`${MetricTypes.cpu_info}.data.result[${index}].metric`,
		{}
	) as CpuInfoMetric;
	const dmidecodeInfo = get(
		data,
		`${MetricTypes.cpu_dmidecode_info}.data.result[${index}].metric`,
		{}
	) as CpuInfoMetric;
	const modelName =
		cpuInfo.model_name ||
		dmidecodeInfo.model_name ||
		cpuInfo.version ||
		dmidecodeInfo.version;
	const core = getValue(lastData[MetricTypes.cpu_total]);
	const thread = core;

	const result = [
		modelName,
		value,
		t('CORES', Number(core)),
		t('THREADS', Number(thread))
	];

	return { list: result, core: core };
};
