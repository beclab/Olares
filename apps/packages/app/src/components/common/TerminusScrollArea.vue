<template>
	<q-scroll-area
		class="terminus-scroll-area"
		:thumb-style="thumbStyle"
		:bar-style="barStyle"
	>
		<slot name="content"></slot>
	</q-scroll-area>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const thumbStyle = ref({
	right: '4px',
	backgroundColor: 'white',
	width: '4px'
});
const barStyle = ref({
	right: '2px',
	backgroundColor: 'white',
	width: '8px',
	opacity: 0.2
});
</script>

<style scoped lang="scss">
.terminus-scroll-area {
	width: 100%;
}
::v-deep(.q-scrollarea__content) {
	// padding: 0 !important;
	max-width: 100%;
}
</style>
