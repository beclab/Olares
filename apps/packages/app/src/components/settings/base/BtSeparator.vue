<template>
	<div
		class="separator"
		:class="bgColor"
		:style="{
			'--offset': `${offset}px`,
			'--height': `${height}px`
		}"
	></div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useQuasar } from 'quasar';

defineProps({
	bgColor: {
		type: String,
		default: 'bg-separator',
		required: false
	},
	offset: {
		type: Number,
		default: 0,
		required: false
	}
});
const $q = useQuasar();
const height = ref($q.platform.is.win ? 0.5 : 1);
</script>

<style scoped lang="scss">
.separator {
	width: calc(100% - var(--offset) * 2);
	height: var(--height);
	margin-left: var(--offset);
	margin-right: var(--offset);
}
</style>
