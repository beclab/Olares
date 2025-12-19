<template>
	<div
		v-if="!deviceStore.isMobile"
		class="bg-background-1"
		style="position: absolute; bottom: 20px; width: 100%; height: 1px"
	/>
</template>

<script lang="ts" setup>
import { useDeviceStore } from '../../stores/settings/device';

const deviceStore = useDeviceStore();
</script>

<style scoped lang="scss"></style>
