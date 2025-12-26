<template>
	<div class="info-filed-root items-center">
		<div class="row items-center justify-start">
			<span class="text-body3 text-cyan-4">
				{{ filed.name }}
			</span>
		</div>

		<div v-if="filed.value" class="text-left">
			<span class="text-white text-body3 field-value">
				{{ filed.format(true) }}
			</span>
		</div>

		<div class="text-body3 field-value" v-else>[{{ t('empty') }}]</div>
	</div>
</template>

<script lang="ts" setup>
import { PropType } from 'vue';
import { Field } from '@didvault/sdk/src/core';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

defineProps({
	filed: {
		type: Object as PropType<Field>,
		required: true
	}
});
</script>

<style scoped lang="scss">
.info-filed-root {
	height: 35px;
	width: 100%;
	margin-right: 40px;

	.field-value {
		margin-top: 4px;
	}
}
</style>
