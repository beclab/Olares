<template>
	<q-input
		class="searchInput text-body3 text-ink-3"
		:model-value="modelValue"
		dense
		borderless
		:name="t('base.search')"
		debounce="500"
		@update:model-value="(value) => emit('update:modelValue', value)"
		input-style="height: 32px"
		:placeholder="placeholder"
		autocomplete="off"
	>
		<template v-slot:prepend>
			<q-icon class="q-ml-sm" color="ink-2" size="16px" name="search" />
		</template>
	</q-input>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { i18n } from '../../boot/i18n';

defineProps({
	placeholder: {
		type: String,
		default: i18n.global.t('base.search'),
		required: false
	},
	modelValue: {
		type: String,
		required: true
	}
});

const emit = defineEmits(['update:modelValue']);
const { t } = useI18n();
</script>

<style lang="scss" scoped>
.searchInput {
	border-radius: var(--8, 8px);
	margin-bottom: 2px;
	border: 1px solid $input-stroke;
	height: 32px !important;
	gap: 10px;
	margin-left: 10px;
	margin-right: 20px;
	width: calc(100% - 60px);
}

::v-deep(.q-field__marginal) {
	height: 32px !important;
}

::v-deep(.q-field__inner) {
	height: 32px !important;
}
</style>

<style lang="scss"></style>
