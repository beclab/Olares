<template>
	<div class="empty_root column justify-center items-center">
		<img
			class="empty-image"
			:src="
				getRequireImage(
					$q.dark.isActive ? 'content_empty_dark.svg' : 'content_empty.svg'
				)
			"
		/>
		<div
			v-if="label"
			class="q-mt-lg text-subtitle2 text-ink-1"
			style="white-space: pre-line"
		>
			{{ label }}
		</div>
	</div>
</template>

<script lang="ts" setup>
import { getRequireImage } from 'src/utils/imageUtils';
import { useQuasar } from 'quasar';

defineProps({
	label: {
		type: String,
		default: ''
	}
});

const $q = useQuasar();
</script>

<style lang="scss" scoped>
.empty_root {
	.empty-image {
		width: 240px;
		height: 240px;
	}
}
</style>
