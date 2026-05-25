<template>
	<div
		class="info-content text-subtitle3 text-ink-1"
		:class="$q.platform.is.mobile ? 'q-pa-md' : 'q-pa-lg'"
	>
		{{ content }}
	</div>
</template>

<script setup lang="ts">
import { useQuasar } from 'quasar';

defineProps({
	content: {
		type: String,
		default: '',
		required: false
	}
});

const $q = useQuasar();
</script>

<style scoped lang="scss">
.info-content {
	overflow-y: scroll;
	height: 100%;
	white-space: pre-wrap;
	word-wrap: break-word;
}
</style>
