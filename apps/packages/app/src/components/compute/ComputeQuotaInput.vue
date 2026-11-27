<template>
	<!-- Exclusive cards take the whole device, so there is no adjustable quota. -->
	<div v-if="!isExclusive" class="compute-quota-block column">
		<div class="compute-quota row items-center justify-between no-wrap">
			<div class="column compute-quota__label">
				<span class="text-body3 text-ink-2">{{
					t('compute.quota_for_current_app')
				}}</span>
				<span v-if="editable" class="text-body3 text-blue-default">
					{{ t('compute.quota_available', { size: availableText }) }}
				</span>
			</div>

			<vram-amount-input
				class="compute-quota__input"
				:model-value="memoryBytes"
				:min-bytes="0"
				:max-bytes="maxBytes"
				:readonly="!editable"
				hide-help
				@update:model-value="onUpdate"
				@update:valid="onValid"
				@update:error="onError"
			/>
		</div>

		<!-- Error sits below the whole card, full width. -->
		<div v-if="editable && errorText" class="compute-quota__error">
			<span class="text-overline">{{ errorText }}</span>
		</div>
	</div>
</template>

<script setup lang="ts">
import VramAmountInput from './VramAmountInput.vue';
import { ComputeDeviceView, ComputeNodeView } from './types';
import { ComputeBinding } from './useComputeBinding';
import {
	formatComputeMemory,
	isExclusiveSupportType,
	isTimeSliceSupportType
} from 'src/constant/compute';
import { computed, PropType, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	node: {
		type: Object as PropType<ComputeNodeView>,
		required: true
	},
	device: {
		type: Object as PropType<ComputeDeviceView>,
		required: true
	},
	binding: {
		type: Object as PropType<ComputeBinding>,
		required: true
	}
});

const { t } = useI18n();

const isExclusive = computed(() =>
	isExclusiveSupportType(props.device.supportType)
);
const isTimeSlice = computed(() =>
	isTimeSliceSupportType(props.device.supportType)
);
// Only MemorySlice exposes an adjustable per-app allocation; TimeSlice renders
// the same box read-only at the full card capacity.
const editable = computed(() => !isExclusive.value && !isTimeSlice.value);

const key = computed(() =>
	props.binding.deviceKey(props.node.raw, props.device.raw)
);

const maxBytes = computed(() => props.device.available);
const availableText = computed(() =>
	formatComputeMemory(props.device.available)
);

const memoryBytes = computed(() => {
	if (!editable.value) return props.device.capacity;
	const alloc = props.binding.allocations[key.value];
	return alloc ? alloc.memory : props.binding.requiredPerCard.value;
});

const onUpdate = (bytes: number) => {
	if (!editable.value) return;
	props.binding.setAllocationMemory(props.node.raw, props.device.raw, bytes);
};

const onValid = (valid: boolean) => {
	if (!editable.value) return;
	props.binding.setAllocationValid(props.node.raw, props.device.raw, valid);
};

const errorText = ref('');
const onError = (message: string) => {
	errorText.value = message;
};
</script>

<style scoped lang="scss">
.compute-quota-block {
	padding-top: 8px;
	gap: 4px;
}

.compute-quota {
	gap: 12px;
	padding: 8px 12px;
	border-radius: 8px;
	background: $background-1;

	&__label {
		min-width: 0;
		gap: 2px;
	}

	&__input {
		flex-shrink: 0;
	}

	&__error {
		width: 100%;
		padding: 0 4px;
		color: $negative;

		span {
			word-break: break-word;
		}
	}
}
</style>
