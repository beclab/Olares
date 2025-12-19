<template>
	<div class="column">
		<div :class="[titleClass]" class="text-ink-1">
			{{ $t('language') }}
		</div>
		<bt-select
			class="q-mt-md"
			v-model="currentLanguage"
			:options="supportLanguages"
			:border="true"
			@update:modelValue="updateLocale"
		/>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useUserStore } from 'src/stores/user';
import BtSelect from 'src/components/base/BtSelect.vue';
import { SupportLanguageType, supportLanguages } from 'src/i18n';

interface Props {
	titleClass?: string;
}

withDefaults(defineProps<Props>(), {
	titleClass: 'text-h6'
});

const { locale } = useI18n();
const userStore = useUserStore();
const currentLanguage = ref(userStore.locale || locale.value);

const updateLocale = async (language: SupportLanguageType) => {
	if (language) {
		await userStore.updateLanguageLocale(language);
	}
};
</script>

<style></style>
