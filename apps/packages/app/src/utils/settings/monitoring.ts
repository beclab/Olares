import {
	UnitKey,
	getSuitableValue as baseGetSuitableValue,
	getSuitableUnit as baseGetSuitableUnit,
	getValueByUnit as baseGetValueByUnit
} from '@bytetrade/core';

export const getValueByUnit = (
	num: string,
	unit: string | undefined,
	precision = 2
) => {
	return baseGetValueByUnit(num, unit, precision);
};

export const getSuitableUnit = (value: any, unitType: UnitKey) => {
	return baseGetSuitableUnit(value, unitType);
};

export const getSuitableValue = (
	value: string,
	unitType: any = 'default',
	defaultValue: string | number = 0
) => {
	return baseGetSuitableValue(value, unitType, defaultValue);
};
