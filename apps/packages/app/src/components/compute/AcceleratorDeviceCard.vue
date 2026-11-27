<template>
	<div class="accelerator-device-card">
		<div class="row items-center justify-between no-wrap" style="gap: 12px">
			<div class="column" style="min-width: 0">
				<div class="row items-center no-wrap" style="gap: 8px">
					<span class="text-body2 text-ink-1 ellipsis">{{ deviceName }}</span>
					<compute-support-type-badge
						:support-type="String(device.supportType)"
					/>
				</div>
				<div class="text-body3 text-ink-3 q-mt-xs">{{ usageText }}</div>
			</div>

			<app-avatar-stack v-if="apps.length > 0" :apps="apps" :max="4" />
			<span v-else class="text-body3 text-ink-3 no-wrap">
				{{ t('No apps assigned') }}
			</span>
		</div>
	</div>
</template>

<script setup lang="ts">
import ComputeSupportTypeBadge from './ComputeSupportTypeBadge.vue';
import AppAvatarStack from './AppAvatarStack.vue';
import {
	ComputeResourceDevice,
	computeModeTitle,
	effectiveDeviceUsedMemory,
	formatComputeMemory,
	isVramComputeMode
} from 'src/constant/compute';
import { AcceleratorAppView } from 'src/stores/settings/accelerator';
import { computed, PropType } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	device: {
		type: Object as PropType<ComputeResourceDevice>,
		required: true
	},
	apps: {
		type: Array as PropType<AcceleratorAppView[]>,
		default: () => []
	}
});

const { t } = useI18n();

const deviceName = computed(
	() =>
		(props.device.name as string) ||
		props.device.cardModel ||
		(props.device.mode ? computeModeTitle(String(props.device.mode)) : '') ||
		props.device.id
);

const usageText = computed(() => {
	const label = isVramComputeMode(props.device.mode)
		? t('Dedicated VRAM')
		: t('Shared RAM');
	const capacity = formatComputeMemory(props.device.memory);
	const used = effectiveDeviceUsedMemory(props.device);
	const suffix =
		used > 0
			? ` (${t('{size} allocated', { size: formatComputeMemory(used) })})`
			: '';
	return `${label}: ${capacity}${suffix}`;
});
</script>

<style scoped lang="scss">
.accelerator-device-card {
	background-color: $background-6;
	border-radius: 12px;
	padding: 12px;
}
</style>
