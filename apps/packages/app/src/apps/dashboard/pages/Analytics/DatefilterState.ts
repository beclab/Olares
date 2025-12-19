import { reactive } from 'vue';
import { parseDateRange } from './date';
import { get } from 'lodash';
import { t } from '@apps/dashboard/src/boot/i18n';

export const dateFilterOption = [
	{
		label: t('ANALYTICS_DATE_OPTION.TODAY'),
		value: '1day',
		unit: 'hour'
	},
	{
		label: t('ANALYTICS_DATE_OPTION.LAST_24_HOURS'),
		value: '24hour',
		unit: 'hour'
	},
	{
		label: t('ANALYTICS_DATE_OPTION.YESTERDAY'),
		value: '-1day',
		unit: 'hour'
	},
	{
		label: t('ANALYTICS_DATE_OPTION.THIS_WEEK'),
		value: '1week',
		unit: 'day',
		divider: true
	},
	{
		label: t('ANALYTICS_DATE_OPTION.LAST_7_DAYS'),
		value: '7day',
		unit: 'day'
	},
	{
		label: t('ANALYTICS_DATE_OPTION.THIS_MONTH'),
		value: '1month',
		unit: 'day',
		divider: true
	},
	{
		label: t('ANALYTICS_DATE_OPTION.LAST_30_DAYS'),
		value: '30day',
		unit: 'day'
	},
	{
		label: t('ANALYTICS_DATE_OPTION.LAST_90_DAYS'),
		value: '90day',
		unit: 'day'
	},
	{
		label: t('ANALYTICS_DATE_OPTION.THIS_YEAR'),
		value: '1year',
		unit: 'month'
	}
	// {
	//   label: 'All time',
	//   value: 'all',
	//   divider: true,
	// },
	// {
	//   label: 'Custom-range',
	//   value: 'custom',
	//   divider: true,
	// },
];

export const DEFAULT_DATE_OPTIONS_VALUE = dateFilterOption[1].value;

export const dateFormatStr: any = {
	hour: 'hA',
	day: 'MMMD',
	month: 'MMM'
};

export const dateValue = reactive<Record<string, any>>({});

export const websiteDateValue = (websiteId: string) =>
	dateValue[websiteId] || DEFAULT_DATE_OPTIONS_VALUE;

export const dateRange = (websiteId: string) => {
	return parseDateRange(
		get(dateValue, `[${websiteId}].value`) || DEFAULT_DATE_OPTIONS_VALUE
	);
};

export const setWebsiteDateDefault = (websites: any) => {
	websites.forEach((item: any) => {
		if (!dateValue[item.id]) {
			dateValue[item.id] = { ...dateFilterOption[1] };
		}
	});
};
