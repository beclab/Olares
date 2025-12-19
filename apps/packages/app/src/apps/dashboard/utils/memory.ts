import {
	getSuitableUnit,
	getValueByUnit
} from '@apps/dashboard/src/utils/monitoring';

export const getThroughput = (size) => {
	const unit = getSuitableUnit(size, 'throughput') || '';
	const value = getValueByUnit(String(size), unit);
	return value + ' ' + unit;
};
