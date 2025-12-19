import {
	getSuitableUnit,
	getValueByUnit
} from '@apps/dashboard/src/utils/monitoring';

export const getDiskSize = (size) => {
	const unit = getSuitableUnit(size, 'disk') || '';
	const value = getValueByUnit(String(size), unit);
	return value ? value + unit : '-';
};
