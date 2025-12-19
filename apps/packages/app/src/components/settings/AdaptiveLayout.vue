<template>
	<div
		:class="auto ? 'adaptive-auto' : 'adaptive-root'"
		v-if="deviceStore.isMobile"
	>
		<slot name="mobile" />
	</div>
	<div :class="auto ? 'adaptive-auto' : 'adaptive-root'" v-else>
		<slot name="pc" />
	</div>
</template>

<script lang="ts" setup>
import { useDeviceStore } from 'src/stores/settings/device';

defineProps({
	auto: {
		type: Boolean,
		default: false
	}
});

const deviceStore = useDeviceStore();
</script>

<style lang="scss" scoped>
.adaptive-root {
	width: 100%;
	height: 100%;
}

.adaptive-auto {
	width: auto;
	height: auto;
}
</style>
