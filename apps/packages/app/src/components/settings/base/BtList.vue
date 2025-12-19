<template>
	<div class="column">
		<div
			v-if="label"
			:class="[
				deviceStore.isMobile ? 'text-subtitle1-m' : 'text-subtitle1',
				first ? 'q-mt-md' : ' q-mt-xl'
			]"
			class="text-ink-1"
		>
			{{ label }}
		</div>
		<q-list
			:class="[
				deviceStore.isMobile ? 'mobile-items-list' : 'q-list-class',
				noMarginTop ? '' : first ? 'q-mt-md' : label ? 'q-mt-sm' : 'q-mt-lg'
			]"
			class="text-background-1"
		>
			<slot />
		</q-list>
	</div>
</template>

<script setup lang="ts">
import { useDeviceStore } from 'src/stores/settings/device';
const deviceStore = useDeviceStore();

defineProps({
	label: String,
	first: {
		type: Boolean,
		default: false
	},
	noMarginTop: {
		type: Boolean,
		default: false
	}
});
</script>
