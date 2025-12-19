import { ref } from 'vue';
import { t } from '@apps/dashboard/boot/i18n';
export const initOptions = [
	{
		label: t('ALL'),
		value: '',
		type: 'cluster'
	}
];
export const options = ref(initOptions);

export const selectValue = ref(options.value[0]);
