import { getLastMonitoringData } from '@apps/control-hub/src/utils/monitoring';
import memory_icon from '@apps/dashboard/src/assets/memory.svg';
import memory_active_icon from '@apps/dashboard/src/assets/memory_active.svg';
import memory_alt_active_icon from '@apps/dashboard/src/assets/memory_alt_active.svg';
import memory_alt_icon from '@apps/dashboard/src/assets/memory_alt.svg';
import hard_drive_active_icon from '@apps/dashboard/src/assets/hard_drive_active.svg';
import hard_drive_icon from '@apps/dashboard/src/assets/hard_drive.svg';
import package_2_active_icon from '@apps/dashboard/src/assets/package_2_active.svg';
import package_2_icon from '@apps/dashboard/src/assets/package_2.svg';
import { get } from 'lodash';
import { PodItem } from '@apps/control-panel-common/src/network/network';
import { t } from '@apps/control-hub/src/boot/i18n';
const getValue = (data: PodItem) => get(data, 'value[1]', 0);

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
	const lastData: { [key: string]: any } = getLastMonitoringData(data);
	const result = [
		{
			name: t('CPU'),
			unitType: 'cpu',
			used: getValue(lastData[MetricTypes.cpu_usage]),
			total: getValue(lastData[MetricTypes.cpu_total]),
			img: memory_icon,
			img_active: memory_active_icon
		},
		{
			name: t('MEMORY'),
			unitType: 'memory',
			used: getValue(lastData[MetricTypes.memory_usage]),
			total: getValue(lastData[MetricTypes.memory_total]),
			img: memory_alt_icon,
			img_active: memory_alt_active_icon
		}
	];

	return result;
};
