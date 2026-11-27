<template>
	<div class="column" style="gap: 12px">
		<compute-device-row
			v-for="dev in node.devices"
			:key="dev.deviceId"
			:device-id="dev.deviceId"
			:name="dev.name"
			:support-type="dev.supportType"
			:capacity-text="capacityText(dev)"
			:selection-type="binding.allowMultiCard.value ? 'checkbox' : 'radio'"
			:selected="isChecked(dev)"
			:disabled="!selectable(dev)"
			:disabled-reason="disabledReason(dev)"
			:warning="cardWarning(dev)"
			@select="onSelect(dev)"
		>
			<!-- Inline quota for the selected card. The component self-hides for
			     Exclusive cards (whole-card allocation, nothing to adjust). -->
			<q-slide-transition>
				<compute-quota-input
					v-if="isChecked(dev)"
					:node="node"
					:device="dev"
					:binding="binding"
				/>
			</q-slide-transition>

			<template v-if="dev.boundApps.length > 0" #apps>
				<div
					class="compute-bound-panel"
					:class="{ 'compute-bound-panel--disabled': !canUnbind(dev) }"
				>
					<compute-bound-apps
						:apps="dev.boundApps"
						:busy-app-name="busyAppName"
						:unbindable="canUnbind(dev)"
						:show-badges="false"
						bg-color=""
						@unbind="(app: AcceleratorAppView) => emit('unbind', app, dev.name)"
					/>
				</div>
			</template>
		</compute-device-row>
	</div>
</template>

<script setup lang="ts">
import ComputeDeviceRow from './ComputeDeviceRow.vue';
import ComputeQuotaInput from './ComputeQuotaInput.vue';
import ComputeBoundApps from './ComputeBoundApps.vue';
import { ComputeDeviceView, ComputeNodeView } from './types';
import { ComputeBinding } from './useComputeBinding';
import {
	formatComputeMemory,
	isDeviceOperable,
	isNodeOperable,
	isVramComputeMode,
	ParsedValidationCode
} from 'src/constant/compute';
import { AcceleratorAppView } from 'src/stores/settings/accelerator';
import { PropType } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	node: {
		type: Object as PropType<ComputeNodeView>,
		required: true
	},
	binding: {
		type: Object as PropType<ComputeBinding>,
		required: true
	},
	parsedValidation: {
		type: Object as PropType<ParsedValidationCode | null>,
		default: null
	},
	busyAppName: {
		type: String,
		default: ''
	}
});

const emit = defineEmits<{
	(e: 'unbind', app: AcceleratorAppView, cardName: string): void;
}>();

const { t } = useI18n();

const capacityText = (dev: ComputeDeviceView): string => {
	const label = isVramComputeMode(props.node.gpuType)
		? t('compute.dedicated_vram')
		: t('compute.shared_ram');
	return `${label}: ${formatComputeMemory(dev.capacity)}`;
};

const canUnbind = (dev: ComputeDeviceView): boolean =>
	props.binding.allowMultiCard.value ||
	dev.capacity >= props.binding.requiredPerCard.value;

const isChecked = (dev: ComputeDeviceView) =>
	props.binding.isDeviceChecked(props.node.raw, dev.raw);

const selectable = (dev: ComputeDeviceView) =>
	props.binding.isDeviceSelectable(dev.raw, props.node.raw);

const onSelect = (dev: ComputeDeviceView) => {
	if (props.binding.allowMultiCard.value) {
		props.binding.toggleDevice(props.node.raw, dev.raw);
	} else {
		props.binding.pickSingleDevice(props.node.raw, dev.raw);
	}
};

const disabledReason = (dev: ComputeDeviceView): string => {
	if (!isNodeOperable(props.node.raw)) {
		return nodeStatusTip(props.node.status);
	}
	if (!isDeviceOperable(dev.raw)) {
		return t('compute.device_not_operable');
	}
	const keys = props.binding.selectedKeys.value;
	if (props.binding.mustSameNode.value && keys.length > 0) {
		const lockedNode = keys[0].split('::')[0];
		if (lockedNode !== props.node.nodeName) {
			return t('compute.must_same_node');
		}
	}
	return '';
};

const nodeStatusTip = (status: string): string =>
	status === 'available' ? '' : t(`compute.node_status_tip.${status}`);

const cardWarning = (dev: ComputeDeviceView): string => {
	const parsed = props.parsedValidation;
	if (
		parsed?.isDeviceScoped &&
		parsed.targetId &&
		dev.deviceId === parsed.targetId
	) {
		return parsed.i18nKey ? t(parsed.i18nKey) : '';
	}
	if (dev.operable) return '';
	if (props.node.status === 'not_enough') {
		return t('compute.card_not_enough_banner');
	}
	return t('compute.card_insufficient_banner');
};
</script>

<style scoped lang="scss">
.compute-bound-panel {
	padding: 8px 12px;
	border-radius: 8px;
	background: $background-1;

	&--disabled {
		opacity: 0.5;
	}
}
</style>
