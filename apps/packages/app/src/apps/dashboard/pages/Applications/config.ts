import { getLastMonitoringData } from '@apps/dashboard/src/utils/monitoring';

import memory_icon from '../../assets/memory.svg';
import memory_active_icon from '../../assets/memory_active.svg';
import memory_alt_active_icon from '../../assets/memory_alt_active.svg';
import memory_alt_icon from '../../assets/memory_alt.svg';
import { getValue } from '../Overview2/config';

export const MetricTypesFormat = (type = 'cluster') => ({
	cpu_usage: `${type}_cpu_usage`,
	cpu_total: `${type}_cpu_limit_hard`,
	memory_usage: `${type}_memory_usage_wo_cache`,
	memory_total: `${type}_memory_limit_hard`
});

export const getTabOptions = (
	data: { [key: string]: string },
	MetricTypes: any,
	index = 0
) => {
	const lastData: { [key: string]: any } = getLastMonitoringData(data, index);
	const result = [
		{
			name: 'CPU',
			unitType: 'cpu',
			used: getValue(lastData[MetricTypes.cpu_usage]),
			total: getValue(lastData[MetricTypes.cpu_total]),
			img: memory_icon,
			img_active: memory_active_icon
		},
		{
			name: 'MEMORY',
			unitType: 'memory',
			used: getValue(lastData[MetricTypes.memory_usage]),
			total: getValue(lastData[MetricTypes.memory_total]),
			img: memory_alt_icon,
			img_active: memory_alt_active_icon
		}
	];

	return result;
};
