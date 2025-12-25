import { t } from '@apps/dashboard/src/boot/i18n';

export const columns: any = [
	{
		name: 'cpu',
		label: t('FAN_OP.CPU_FAN_SPEED'),
		field: 'cpu',
		align: 'left'
	},
	{
		name: 'gpu',
		label: t('FAN_OP.GPU_FAN_SPEED'),
		field: 'gpu',
		align: 'center',
		headerStyle: 'width: 33%'
	},
	{
		name: 'cpuDn',
		label: t('FAN_OP.CPU_TEMPERATURE_RANGE'),
		field: 'cpuDn',
		align: 'center',
		headerStyle: 'width: 33%'
	},
	{
		name: 'gpuDn',
		label: t('FAN_OP.GPU_TEMPERATURE_RANGE'),
		field: 'gpuDn',
		align: 'right'
	}
];

export const FanSpeedMaxCPU = 2900;
export const FanSpeedMaxGPU = 3100;

export const cpuStop1 = 2100;
export const cpuStop2 = 2500;
export const gpuStop1 = 2300;
export const gpuStop2 = 2700;

export const cpuFanColorStops = [
	0,
	cpuStop1 / FanSpeedMaxCPU,
	cpuStop2 / FanSpeedMaxCPU
];

export const gpuFanColorStops = [
	0,
	gpuStop1 / FanSpeedMaxGPU,
	gpuStop2 / FanSpeedMaxGPU
];

export const tableData = [
	{
		uuid: '1',
		cpu: 0,
		gpu: 0,
		cpuDn: '0 - 54',
		gpuDn: '0 - 48'
	},
	{
		uuid: '2',
		cpu: 1100,
		gpu: 1300,
		cpuDn: '47 - 64',
		gpuDn: '39 - 58'
	},
	{
		uuid: '3',
		cpu: 1300,
		gpu: 1500,
		cpuDn: '54 - 71',
		gpuDn: '48 - 65'
	},
	{
		uuid: '4',
		cpu: 1500,
		gpu: 1700,
		cpuDn: '64 - 74',
		gpuDn: '58 - 68'
	},
	{
		uuid: '5',
		cpu: 1800,
		gpu: 2000,
		cpuDn: '71 - 77',
		gpuDn: '65 - 71'
	},
	{
		uuid: '6',
		cpu: cpuStop1,
		gpu: gpuStop1,
		cpuDn: '74 - 80',
		gpuDn: '68 - 74'
	},
	{
		uuid: '7',
		cpu: 2300,
		gpu: 2500,
		cpuDn: '77 - 83',
		gpuDn: '71 - 77'
	},
	{
		uuid: '8',
		cpu: gpuStop1,
		gpu: gpuStop2,
		cpuDn: '80 - 86',
		gpuDn: '75 - 80'
	},
	{
		uuid: '9',
		cpu: 2700,
		gpu: 2900,
		cpuDn: '83 - 88',
		gpuDn: '77 - 83'
	},
	{
		uuid: '10',
		cpu: 2900,
		gpu: 3100,
		cpuDn: '86 - 96',
		gpuDn: '80 - 86'
	}
];
