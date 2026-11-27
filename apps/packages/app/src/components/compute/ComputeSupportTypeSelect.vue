<template>
	<!-- A single available type is not switchable — show it as plain text. -->
	<div v-if="options.length <= 1" class="text-body3 text-ink-2">
		{{ options[0]?.label ?? t(acceleratorSupportTypeLabelKey(modelValue)) }}
	</div>
	<!-- More than one type: a dropdown to switch. -->
	<bt-select-v3
		v-else
		:key="`${selectKey}-${resetKey}`"
		:model-value="modelValue"
		input-class="text-body3 text-ink-2"
		:options="options"
		:width="deviceStore.isMobile ? '100%' : width"
		show-selected-icon
		@update:model-value="(value) => emit('update:model-value', value as string)"
	/>
</template>

<script setup lang="ts">
import BtSelectV3 from 'src/components/settings/base/BtSelectV3.vue';
import { acceleratorSupportTypeLabelKey } from 'src/constant/compute';
import { useDeviceStore } from 'src/stores/settings/device';
import { computed, PropType } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	/** Currently committed support type. */
	modelValue: {
		type: String,
		required: true
	},
	/** All switchable support types for this device (falls back to modelValue). */
	availableSupportTypes: {
		type: Array as PropType<(string | undefined)[]>,
		default: () => []
	},
	/** Stable key (e.g. device id) so the inner select re-mounts per device. */
	selectKey: {
		type: [String, Number],
		default: ''
	},
	/**
	 * Bump to force the dropdown to re-sync with `modelValue` after a cancelled
	 * or failed switch (it otherwise keeps its optimistic local selection).
	 */
	resetKey: {
		type: Number,
		default: 0
	},
	/** Desktop width; mobile always stretches to 100%. */
	width: {
		type: String,
		default: '140px'
	}
});

const emit = defineEmits<{ (e: 'update:model-value', value: string): void }>();

const { t } = useI18n();
const deviceStore = useDeviceStore();

const options = computed(() => {
	const types =
		props.availableSupportTypes.length > 0
			? props.availableSupportTypes
			: [props.modelValue];
	return types
		.filter((type): type is string => !!type)
		.map((type) => ({
			label: t(acceleratorSupportTypeLabelKey(type)),
			value: type,
			enable: true
		}));
});
</script>
