<template>
	<div class="row items-end justify-between title-part text-body3">
		<div class="text-ink-2">{{ title }}</div>
		<div class="text-light-blue-default reset" @click="emit('reset')">
			{{ t('reset') }}
		</div>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	}
});

const emit = defineEmits(['reset']);
const { t } = useI18n();
</script>

<style scoped lang="scss">
.title-part {
	height: 24px;
	width: 100%;

	.reset {
		cursor: pointer;
	}
}
</style>
