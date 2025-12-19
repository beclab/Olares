<template>
	<div
		class="text-ink-1"
		:class="deviceStore.isMobile ? 'text-subtitle1-m' : 'text-subtitle1'"
	>
		<slot />
	</div>
</template>

<script setup lang="ts">
import { useDeviceStore } from 'src/stores/settings/device';
const deviceStore = useDeviceStore();
</script>

<style scoped lang="scss"></style>
