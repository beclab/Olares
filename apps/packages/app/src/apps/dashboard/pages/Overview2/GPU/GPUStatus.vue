<template>
	<div
		class="gpu-status-container row inline justify-center items-center q-py-xs q-px-md"
		:class="[
			isExternal
				? 'bg-yellow-alpha text-yellow-default'
				: !isBoolean(health)
				? ''
				: health
				? 'bg-green-alpha text-green-default'
				: 'bg-red-alpha text-red-default'
		]"
	>
		{{
			isExternal
				? $t('GPU_OP.UNMANAGED')
				: !isBoolean(health)
				? ''
				: health
				? $t('GPU_OP.HEALTHY')
				: $t('GPU_OP.HARDWARE_ERROR')
		}}
	</div>
</template>

<script setup lang="ts">
import { isBoolean } from 'lodash';

interface Props {
	health: boolean;
	isExternal: boolean;
}

const props = defineProps<Props>();
</script>

<style lang="scss" scoped>
.gpu-status-container {
	border-radius: 4px;
}
</style>
