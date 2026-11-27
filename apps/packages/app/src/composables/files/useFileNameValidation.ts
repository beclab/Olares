import { computed, ComputedRef, Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { disabledFolderNames } from 'src/utils/interface/files';

export function useFileNameValidation(name: Ref<string>): {
	errorMessage: ComputedRef<string>;
} {
	const { t } = useI18n();

	const errorMessage = computed(() => {
		const value = name.value ?? '';
		if (value.includes('\\') || value.includes('/')) {
			return t('files.backslash_create');
		}
		if (disabledFolderNames.includes(value)) {
			return t('files.File/folder/library names cannot be {name}', {
				name: value
			});
		}
		return '';
	});

	return { errorMessage };
}
