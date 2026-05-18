import { ref } from 'vue';
import { t } from 'src/boot/dashboard-i18n';
export const initOptions = [
	{
		label: t('ALL'),
		value: ''
	}
];
export const options = ref(initOptions);

export const selectValue = ref(options.value[0]);
