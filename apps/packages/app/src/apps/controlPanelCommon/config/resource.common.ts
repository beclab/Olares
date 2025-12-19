import { timeRangeFormate } from '@apps/control-panel-common/src/containers/Monitoring/utils';

export const timeRangeDefault = timeRangeFormate('8h', 16);

export const timeParams = {
	step: '1800s',
	times: 16,
	lastTime: '8h'
};
