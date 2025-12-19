<template>
	<div class="row items-center flex-gap-sm no-wrap">
		<div
			:style="{
				height: '12px',
				width: '12px',
				borderRadius: '50%',
				display: 'inline-block'
			}"
			:class="[current.color]"
		></div>
		<span>{{ current.label }}</span>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { computed } from 'vue';
import { TaskStatusOptions } from './config';

interface Props {
	status: string;
}
const props = withDefaults(defineProps<Props>(), {
	status: 'unknown'
});

const { t } = useI18n();

const current = computed(() => {
	const target = TaskStatusOptions.value.find(
		(item) => item.value === props.status
	);
	if (target) {
		return target;
	} else {
		return TaskStatusOptions[TaskStatusOptions.value.length - 1];
	}
});
</script>

<style></style>
