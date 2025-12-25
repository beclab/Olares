import { PodItem } from '@apps/dashboard/src/types/network';
import { get } from 'lodash';
import { t } from '@apps/dashboard/boot/i18n';
import { timeRangeFormate } from '@apps/control-panel-common/src/containers/Monitoring/utils';
import { getAreaChartOps } from '@apps/control-hub/src/utils/monitoring';

export const MetricTypes = {
	pod_count: `node_pod_running_count`
};
export type MetricTypesType = typeof MetricTypes;

export const getValue = (data: PodItem) => get(data, 'value[1]', 0);

export const getPodsList = (data: { [key: string]: string }) => {
	const firstData: any = get(data, `${MetricTypes.pod_count}.data.result`, []);
	return firstData.map((item, index) =>
		getAreaChartOps(getMonitoringCfgs(data, index))
	);
};

export const getMonitoringCfgs = (
	data: { [key: string]: string },
	index = 0
) => ({
	type: 'load',
	unit: '',
	title: get(
		data,
		`${MetricTypes.pod_count}.data.result[${index}].metric.node`,
		'node'
	),
	legend: [t('COUNT')],
	data: [get(data, `${MetricTypes.pod_count}.data.result[${index}]`, {})]
});
