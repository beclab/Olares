import { ref } from 'vue';
import { t } from '@apps/dashboard/src/boot/i18n';
export const initOptions = [
	{
		label: t('ALL'),
		value: ''
	}
];
export const options = ref(initOptions);

export const selectValue = ref(options.value[0]);
